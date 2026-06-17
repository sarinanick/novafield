package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
	"novafield-api/store"
)

var allowedExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

var allowedArchiveExts = map[string]bool{
	".zip": true,
}

var maxUploadSize int64 = 10 << 20

var (
	maxImageWidth = envInt("IMAGE_MAX_WIDTH", 1200)
	thumbWidth    = envInt("THUMB_WIDTH", 300)
	thumbHeight   = envInt("THUMB_HEIGHT", 300)
)

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		Error(w, 405, "Method not allowed")
		return
	}

	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		Error(w, 400, "File too large (max 10MB)")
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		Error(w, 400, "No file provided")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(handler.Filename))
	isImage := allowedExts[ext]
	isArchive := allowedArchiveExts[ext]

	if !isImage && !isArchive {
		Error(w, 400, "File type not allowed. Use: jpg, png, gif, webp, zip")
		return
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		Error(w, 500, "Failed to create uploads directory")
		return
	}

	base := strings.TrimSuffix(handler.Filename, filepath.Ext(handler.Filename))
	filename := fmt.Sprintf("%s_%s%s", store.NewID()[:8], sanitizeFilename(base), ext)
	dstPath := filepath.Join("uploads", filename)

	if isImage {
		data, err := io.ReadAll(file)
		if err != nil {
			Error(w, 500, "Failed to read file")
			return
		}

		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			Error(w, 500, "Failed to save file")
			return
		}

		resizedPath, thumbPath, err := processImage(data, filename, ext)
		if err != nil {
			JSON(w, 201, H{
				"url":      "/uploads/" + filename,
				"filename": filename,
			})
			return
		}

		resp := H{
			"url":      "/uploads/" + filename,
			"filename": filename,
		}
		if resizedPath != "" {
			resp["resizedUrl"] = "/uploads/" + resizedPath
		}
		if thumbPath != "" {
			resp["thumbnailUrl"] = "/uploads/" + thumbPath
		}
		JSON(w, 201, resp)
		return
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		Error(w, 500, "Failed to save file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		Error(w, 500, "Failed to write file")
		return
	}

	JSON(w, 201, H{"url": "/uploads/" + filename, "filename": filename})
}

func processImage(data []byte, filename, ext string) (resizedPath, thumbPath string, err error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", "", err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= maxImageWidth && width <= thumbWidth {
		return "", "", nil
	}

	if err := os.MkdirAll("uploads/thumbs", 0755); err != nil {
		return "", "", err
	}

	baseName := strings.TrimSuffix(filename, ext)

	if width > maxImageWidth {
		resized := resize.Resize(uint(maxImageWidth), 0, img, resize.Lanczos3)
		resizedName := baseName + "_resized" + ext
		resizedPath = filepath.Join("uploads", resizedName)
		if err := encodeImage(resizedPath, ext, resized); err != nil {
			return "", "", err
		}
		resizedPath = resizedName
	}

	thumb := resize.Thumbnail(uint(thumbWidth), uint(thumbHeight), img, resize.Lanczos3)
	thumbName := baseName + "_thumb" + ext
	thumbPathFull := filepath.Join("uploads/thumbs", thumbName)
	if err := encodeImage(thumbPathFull, ext, thumb); err != nil {
		return resizedPath, "", err
	}
	thumbPath = "thumbs/" + thumbName

	_ = height
	return resizedPath, thumbPath, nil
}

func encodeImage(path, ext string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(f, img, &jpeg.Options{Quality: 85})
	case ".png":
		return png.Encode(f, img)
	case ".gif":
		return gif.Encode(f, img, nil)
	default:
		return png.Encode(f, img)
	}
}

var unsafeChars = regexp.MustCompile(`[^a-zA-Z0-9_\-]`)

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	ext := filepath.Ext(name)
	name = strings.TrimSuffix(name, ext)
	name = unsafeChars.ReplaceAllString(name, "_")
	if name == "" {
		name = "file"
	}
	if len(name) > 50 {
		name = name[:50]
	}
	return name
}
