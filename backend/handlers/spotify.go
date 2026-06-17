package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"novafield-api/database"
	"novafield-api/models"
	"os"
	"strings"
	"time"
)

const (
	spotifyAuthURL  = "https://accounts.spotify.com/authorize"
	spotifyTokenURL = "https://accounts.spotify.com/api/token"
	spotifyAPIBase  = "https://api.spotify.com/v1"
)

type SpotifyTrackInfo struct {
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	AlbumArt string `json:"albumArt"`
	TrackURL string `json:"trackUrl"`
	IsPlaying bool  `json:"isPlaying"`
}

func getSpotifyClientID() string {
	return os.Getenv("SPOTIFY_CLIENT_ID")
}

func getSpotifyClientSecret() string {
	return os.Getenv("SPOTIFY_CLIENT_SECRET")
}

func getSpotifyRedirectURI() string {
	uri := os.Getenv("SPOTIFY_REDIRECT_URI")
	if uri == "" {
		uri = "http://localhost:3001/api/v1/spotify/callback"
	}
	return uri
}

func SpotifyAuthHandler(w http.ResponseWriter, r *http.Request) {
	clientID := getSpotifyClientID()
	if clientID == "" {
		Error(w, 500, "Spotify client ID not configured")
		return
	}

	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	state := base64.URLEncoding.EncodeToString([]byte(user.ID))
	redirectURI := getSpotifyRedirectURI()

	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("response_type", "code")
	params.Add("redirect_uri", redirectURI)
	params.Add("scope", "user-read-playback-state user-modify-playback-state user-read-currently-playing")
	params.Add("state", state)
	params.Add("show_dialog", "true")

	http.Redirect(w, r, spotifyAuthURL+"?"+params.Encode(), http.StatusTemporaryRedirect)
}

func SpotifyCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errParam := r.URL.Query().Get("error")

	if errParam != "" {
		http.Redirect(w, r, "http://localhost:3000/world?spotify=error", http.StatusTemporaryRedirect)
		return
	}

	if code == "" || state == "" {
		http.Redirect(w, r, "http://localhost:3000/world?spotify=error", http.StatusTemporaryRedirect)
		return
	}

	userIDBytes, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		http.Redirect(w, r, "http://localhost:3000/world?spotify=error", http.StatusTemporaryRedirect)
		return
	}
	userID := string(userIDBytes)

	clientID := getSpotifyClientID()
	clientSecret := getSpotifyClientSecret()
	redirectURI := getSpotifyRedirectURI()

	data := url.Values{}
	data.Add("grant_type", "authorization_code")
	data.Add("code", code)
	data.Add("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", spotifyTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		http.Redirect(w, r, "http://localhost:3000/world?spotify=error", http.StatusTemporaryRedirect)
		return
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Redirect(w, r, "http://localhost:3000/world?spotify=error", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil || tokenResp.AccessToken == "" {
		http.Redirect(w, r, "http://localhost:3000/world?spotify=error", http.StatusTemporaryRedirect)
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Users {
		if d.Users[i].ID == userID {
			d.Users[i].SpotifyAccessToken = tokenResp.AccessToken
			d.Users[i].SpotifyRefreshToken = tokenResp.RefreshToken
			d.Users[i].SpotifyTokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Unix()
			d.Users[i].SpotifyConnected = true
			break
		}
	}
	d.Mu.Unlock()
	d.Save()

	http.Redirect(w, r, "http://localhost:3000/world?spotify=connected", http.StatusTemporaryRedirect)
}

func refreshSpotifyToken(user *models.User) (string, error) {
	if user.SpotifyRefreshToken == "" {
		return "", fmt.Errorf("no refresh token")
	}

	clientID := getSpotifyClientID()
	clientSecret := getSpotifyClientSecret()

	data := url.Values{}
	data.Add("grant_type", "refresh_token")
	data.Add("refresh_token", user.SpotifyRefreshToken)

	req, err := http.NewRequest("POST", spotifyTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil || tokenResp.AccessToken == "" {
		return "", fmt.Errorf("failed to refresh token")
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Users {
		if d.Users[i].ID == user.ID {
			d.Users[i].SpotifyAccessToken = tokenResp.AccessToken
			if tokenResp.RefreshToken != "" {
				d.Users[i].SpotifyRefreshToken = tokenResp.RefreshToken
			}
			d.Users[i].SpotifyTokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Unix()
			break
		}
	}
	d.Mu.Unlock()
	d.Save()

	return tokenResp.AccessToken, nil
}

func getValidSpotifyToken(user *models.User) (string, error) {
	if user.SpotifyAccessToken == "" {
		return "", fmt.Errorf("not connected")
	}

	if time.Now().Unix() >= user.SpotifyTokenExpiry {
		return refreshSpotifyToken(user)
	}

	return user.SpotifyAccessToken, nil
}

func SpotifyStatusHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	if !user.SpotifyConnected {
		JSON(w, 200, H{"connected": false})
		return
	}

	token, err := getValidSpotifyToken(user)
	if err != nil {
		JSON(w, 200, H{"connected": false, "error": "Token expired, please reconnect"})
		return
	}

	req, err := http.NewRequest("GET", spotifyAPIBase+"/me/player/currently-playing", nil)
	if err != nil {
		Error(w, 500, "Failed to create request")
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Error(w, 500, "Failed to get playback")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		JSON(w, 200, H{"connected": true, "playing": false})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Error(w, 500, "Failed to read response")
		return
	}

	var playback struct {
		IsPlaying bool `json:"is_playing"`
		Item      struct {
			Name   string `json:"name"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
			Album struct {
				Images []struct {
					URL string `json:"url"`
				} `json:"images"`
			} `json:"album"`
			ExternalURLs struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
		} `json:"item"`
	}

	if err := json.Unmarshal(body, &playback); err != nil {
		JSON(w, 200, H{"connected": true, "playing": false})
		return
	}

	artist := ""
	if len(playback.Item.Artists) > 0 {
		artist = playback.Item.Artists[0].Name
	}

	albumArt := ""
	if len(playback.Item.Album.Images) > 0 {
		albumArt = playback.Item.Album.Images[0].URL
	}

	JSON(w, 200, H{
		"connected": true,
		"playing":   playback.IsPlaying,
		"track": SpotifyTrackInfo{
			Name:      playback.Item.Name,
			Artist:    artist,
			AlbumArt:  albumArt,
			TrackURL:  playback.Item.ExternalURLs.Spotify,
			IsPlaying: playback.IsPlaying,
		},
	})
}

func SpotifyShareHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Share bool `json:"share"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Users {
		if d.Users[i].ID == user.ID {
			d.Users[i].SpotifySharing = req.Share
			if !req.Share {
				d.Users[i].SpotifyTrackName = ""
				d.Users[i].SpotifyTrackArtist = ""
				d.Users[i].SpotifyAlbumArt = ""
				d.Users[i].SpotifyTrackURL = ""
			}
			break
		}
	}
	d.Mu.Unlock()
	d.Save()

	if req.Share {
		token, err := getValidSpotifyToken(user)
		if err == nil {
			updateSharedTrack(user.ID, token)
		}
	}

	JSON(w, 200, H{"sharing": req.Share})
}

func updateSharedTrack(userID, token string) {
	req, err := http.NewRequest("GET", spotifyAPIBase+"/me/player/currently-playing", nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode == 204 {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var playback struct {
		Item struct {
			Name   string `json:"name"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
			Album struct {
				Images []struct {
					URL string `json:"url"`
				} `json:"images"`
			} `json:"album"`
			ExternalURLs struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
		} `json:"item"`
	}

	if err := json.Unmarshal(body, &playback); err != nil {
		return
	}

	artist := ""
	if len(playback.Item.Artists) > 0 {
		artist = playback.Item.Artists[0].Name
	}

	albumArt := ""
	if len(playback.Item.Album.Images) > 0 {
		albumArt = playback.Item.Album.Images[0].URL
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Users {
		if d.Users[i].ID == userID {
			d.Users[i].SpotifyTrackName = playback.Item.Name
			d.Users[i].SpotifyTrackArtist = artist
			d.Users[i].SpotifyAlbumArt = albumArt
			d.Users[i].SpotifyTrackURL = playback.Item.ExternalURLs.Spotify
			break
		}
	}
	d.Mu.Unlock()
	d.Save()
}

func SpotifyDisconnectHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Users {
		if d.Users[i].ID == user.ID {
			d.Users[i].SpotifyAccessToken = ""
			d.Users[i].SpotifyRefreshToken = ""
			d.Users[i].SpotifyTokenExpiry = 0
			d.Users[i].SpotifyConnected = false
			d.Users[i].SpotifySharing = false
			d.Users[i].SpotifyTrackName = ""
			d.Users[i].SpotifyTrackArtist = ""
			d.Users[i].SpotifyAlbumArt = ""
			d.Users[i].SpotifyTrackURL = ""
			break
		}
	}
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"message": "Spotify disconnected"})
}

func SpotifyPlaybackHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	token, err := getValidSpotifyToken(user)
	if err != nil {
		Error(w, 400, "Spotify not connected")
		return
	}

	var req struct {
		Action string `json:"action"` // play, pause, next, previous
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	var spotifyReq *http.Request
	switch req.Action {
	case "play":
		spotifyReq, _ = http.NewRequest("PUT", spotifyAPIBase+"/me/player/play", nil)
	case "pause":
		spotifyReq, _ = http.NewRequest("PUT", spotifyAPIBase+"/me/player/pause", nil)
	case "next":
		spotifyReq, _ = http.NewRequest("POST", spotifyAPIBase+"/me/player/next", nil)
	case "previous":
		spotifyReq, _ = http.NewRequest("POST", spotifyAPIBase+"/me/player/previous", nil)
	default:
		Error(w, 400, "Invalid action")
		return
	}

	spotifyReq.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(spotifyReq)
	if err != nil {
		Error(w, 500, "Failed to control playback")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 || resp.StatusCode == 200 {
		JSON(w, 200, H{"message": "Playback updated"})
	} else {
		Error(w, resp.StatusCode, "Spotify API error")
	}
}

func SpotifySearchHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	token, err := getValidSpotifyToken(user)
	if err != nil {
		Error(w, 400, "Spotify not connected")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		Error(w, 400, "Search query required")
		return
	}

	spotifyReq, err := http.NewRequest("GET", spotifyAPIBase+"/search?q="+url.QueryEscape(query)+"&type=track&limit=10", nil)
	if err != nil {
		Error(w, 500, "Failed to create request")
		return
	}
	spotifyReq.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(spotifyReq)
	if err != nil {
		Error(w, 500, "Failed to search")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Error(w, 500, "Failed to read response")
		return
	}

	var searchResult struct {
		Tracks struct {
			Items []struct {
				Name   string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
				Album struct {
					Images []struct {
						URL string `json:"url"`
					} `json:"images"`
				} `json:"album"`
				URI            string `json:"uri"`
				ExternalURLs struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
			} `json:"items"`
		} `json:"tracks"`
	}

	if err := json.Unmarshal(body, &searchResult); err != nil {
		Error(w, 500, "Failed to parse response")
		return
	}

	var tracks []map[string]interface{}
	for _, item := range searchResult.Tracks.Items {
		artist := ""
		if len(item.Artists) > 0 {
			artist = item.Artists[0].Name
		}
		albumArt := ""
		if len(item.Album.Images) > 0 {
			albumArt = item.Album.Images[0].URL
		}
		tracks = append(tracks, map[string]interface{}{
			"name":      item.Name,
			"artist":    artist,
			"albumArt":  albumArt,
			"uri":       item.URI,
			"trackUrl":  item.ExternalURLs.Spotify,
		})
	}

	JSON(w, 200, H{"tracks": tracks})
}
