package handlers

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestJPEG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes()
}

func createTestPNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

type multipartData struct {
	buf         *bytes.Buffer
	contentType string
}

func createMultipartRequest(filename string, data []byte) multipartData {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, _ := w.CreateFormFile("file", filename)
	part.Write(data)
	w.Close()
	return multipartData{buf: &buf, contentType: w.FormDataContentType()}
}

func cleanupUploads() {
	os.RemoveAll("uploads")
}

func TestUploadJPEG_LargeImage(t *testing.T) {
	resetDB()
	cleanupUploads()
	defer cleanupUploads()

	_, token := createTestUser("client")
	imgData := createTestJPEG(2000, 1500)

	md := createMultipartRequest("photo.jpg", imgData)
	req := httptest.NewRequest("POST", "/api/v1/upload", md.buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", md.contentType)
	rr := httptest.NewRecorder()

	UploadHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["url"] == nil {
		t.Fatal("expected url")
	}
	if result["resizedUrl"] == nil {
		t.Fatal("expected resizedUrl for 2000px image")
	}
	if result["thumbnailUrl"] == nil {
		t.Fatal("expected thumbnailUrl for 2000px image")
	}

	resizedUrl := result["resizedUrl"].(string)
	if !strings.Contains(resizedUrl, "_resized") {
		t.Errorf("expected resized url to contain '_resized', got %s", resizedUrl)
	}

	thumbUrl := result["thumbnailUrl"].(string)
	if !strings.Contains(thumbUrl, "thumbs/") {
		t.Errorf("expected thumb url to contain 'thumbs/', got %s", thumbUrl)
	}
	if !strings.Contains(thumbUrl, "_thumb") {
		t.Errorf("expected thumb url to contain '_thumb', got %s", thumbUrl)
	}
}

func TestUploadJPEG_SmallImage(t *testing.T) {
	resetDB()
	cleanupUploads()
	defer cleanupUploads()

	_, token := createTestUser("client")
	imgData := createTestJPEG(800, 600)

	md := createMultipartRequest("small.jpg", imgData)
	req := httptest.NewRequest("POST", "/api/v1/upload", md.buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", md.contentType)
	rr := httptest.NewRecorder()

	UploadHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["url"] == nil {
		t.Fatal("expected url")
	}
	if result["resizedUrl"] != nil {
		t.Error("should not have resizedUrl for 800px image")
	}
	if result["thumbnailUrl"] == nil {
		t.Fatal("should have thumbnailUrl (800 > 300 thumbWidth)")
	}
}

func TestUploadPNG_LargeImage(t *testing.T) {
	resetDB()
	cleanupUploads()
	defer cleanupUploads()

	_, token := createTestUser("client")
	imgData := createTestPNG(1920, 1080)

	md := createMultipartRequest("screenshot.png", imgData)
	req := httptest.NewRequest("POST", "/api/v1/upload", md.buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", md.contentType)
	rr := httptest.NewRecorder()

	UploadHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["resizedUrl"] == nil {
		t.Fatal("expected resizedUrl for 1920px image")
	}
	if result["thumbnailUrl"] == nil {
		t.Fatal("expected thumbnailUrl for 1920px image")
	}
}

func TestUploadZip(t *testing.T) {
	resetDB()
	cleanupUploads()
	defer cleanupUploads()

	_, token := createTestUser("client")

	md := createMultipartRequest("project.zip", []byte("PK\x03\x04fake zip content"))
	req := httptest.NewRequest("POST", "/api/v1/upload", md.buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", md.contentType)
	rr := httptest.NewRecorder()

	UploadHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["url"] == nil {
		t.Fatal("expected url")
	}
	if result["resizedUrl"] != nil {
		t.Error("zip should not have resizedUrl")
	}
}

func TestUpload_DisallowedExtension(t *testing.T) {
	resetDB()

	_, token := createTestUser("client")

	md := createMultipartRequest("malware.exe", []byte("bad stuff"))
	req := httptest.NewRequest("POST", "/api/v1/upload", md.buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", md.contentType)
	rr := httptest.NewRecorder()

	UploadHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpload_Unauthorized(t *testing.T) {
	resetDB()

	md := createMultipartRequest("test.jpg", createTestJPEG(100, 100))
	req := httptest.NewRequest("POST", "/api/v1/upload", md.buf)
	req.Header.Set("Content-Type", md.contentType)
	rr := httptest.NewRecorder()

	UploadHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestUpload_CreatesResizedFile(t *testing.T) {
	resetDB()
	cleanupUploads()
	defer cleanupUploads()

	_, token := createTestUser("client")
	imgData := createTestJPEG(2000, 1500)

	md := createMultipartRequest("big.jpg", imgData)
	req := httptest.NewRequest("POST", "/api/v1/upload", md.buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", md.contentType)
	rr := httptest.NewRecorder()

	UploadHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	resizedUrl := result["resizedUrl"].(string)
	resizedPath := filepath.Join("uploads", strings.TrimPrefix(resizedUrl, "/uploads/"))

	info, err := os.Stat(resizedPath)
	if err != nil {
		t.Fatalf("resized file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("resized file is empty")
	}
}

func TestUpload_CreatesThumbnailFile(t *testing.T) {
	resetDB()
	cleanupUploads()
	defer cleanupUploads()

	_, token := createTestUser("client")
	imgData := createTestJPEG(2000, 1500)

	md := createMultipartRequest("big.jpg", imgData)
	req := httptest.NewRequest("POST", "/api/v1/upload", md.buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", md.contentType)
	rr := httptest.NewRecorder()

	UploadHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	thumbUrl := result["thumbnailUrl"].(string)
	thumbPath := filepath.Join("uploads", strings.TrimPrefix(thumbUrl, "/uploads/"))

	info, err := os.Stat(thumbPath)
	if err != nil {
		t.Fatalf("thumbnail file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("thumbnail file is empty")
	}
}

func TestUpload_ThumbnailDimensions(t *testing.T) {
	resetDB()
	cleanupUploads()
	defer cleanupUploads()

	_, token := createTestUser("client")
	imgData := createTestJPEG(2000, 1500)

	md := createMultipartRequest("big.jpg", imgData)
	req := httptest.NewRequest("POST", "/api/v1/upload", md.buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", md.contentType)
	rr := httptest.NewRecorder()

	UploadHandler(rr, req)

	result := decodeJSON(rr)
	thumbUrl := result["thumbnailUrl"].(string)
	thumbPath := filepath.Join("uploads", strings.TrimPrefix(thumbUrl, "/uploads/"))

	f, err := os.Open(thumbPath)
	if err != nil {
		t.Fatalf("open thumbnail: %v", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode thumbnail: %v", err)
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w > thumbWidth || h > thumbHeight {
		t.Errorf("thumbnail too large: %dx%d, max %dx%d", w, h, thumbWidth, thumbHeight)
	}
}

func TestUpload_ResizedDimensions(t *testing.T) {
	resetDB()
	cleanupUploads()
	defer cleanupUploads()

	_, token := createTestUser("client")
	imgData := createTestJPEG(3000, 2000)

	md := createMultipartRequest("huge.jpg", imgData)
	req := httptest.NewRequest("POST", "/api/v1/upload", md.buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", md.contentType)
	rr := httptest.NewRecorder()

	UploadHandler(rr, req)

	result := decodeJSON(rr)
	resizedUrl := result["resizedUrl"].(string)
	resizedPath := filepath.Join("uploads", strings.TrimPrefix(resizedUrl, "/uploads/"))

	f, err := os.Open(resizedPath)
	if err != nil {
		t.Fatalf("open resized: %v", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode resized: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() > maxImageWidth {
		t.Errorf("resized width %d exceeds max %d", bounds.Dx(), maxImageWidth)
	}
}

func bodyWriterContentType(buf *bytes.Buffer) string {
	return "multipart/form-data; boundary=" + buf.String()[:strings.Index(buf.String(), "\r\n")]
}
