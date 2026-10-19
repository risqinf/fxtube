package main

import (
	"os/exec"
	"sync"
	"time"
)

// StreamCacheItem holds resolved stream URLs and expiration
type StreamCacheItem struct {
	VideoURL   string
	AudioURL   string
	IsCombined bool
	ExpiresAt  time.Time
}

// FeedCacheItem holds cached feed/search results to prevent excessive yt-dlp launches
type FeedCacheItem struct {
	Videos    []VideoItem
	ExpiresAt time.Time
}

// Thumbnail represents video thumbnail data
type Thumbnail struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// VideoItem represents a single YouTube video
type VideoItem struct {
	ID             string      `json:"id"`
	Title          string      `json:"title"`
	URL            string      `json:"url"`
	Channel        string      `json:"channel"`
	Uploader       string      `json:"uploader"`
	Duration       int64       `json:"duration"`
	DurationString string      `json:"duration_string"`
	Thumbnails     []Thumbnail `json:"thumbnails"`
}

// PlayRequest represents a play action from the frontend
type PlayRequest struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Mode    string `json:"mode"`
	Quality string `json:"quality"`
}

// AccountModeRequest changes YouTube login mode
type AccountModeRequest struct {
	Mode string `json:"mode"`
}

// PlayerCommandRequest represents an MPV IPC command from the UI
type PlayerCommandRequest struct {
	Command string `json:"command"`
}

// PlayerStatusResponse represents MPV state sent to the UI
type PlayerStatusResponse struct {
	Playing  bool    `json:"playing"`
	Paused   bool    `json:"paused"`
	Title    string  `json:"title"`
	URL      string  `json:"url"`
	Quality  string  `json:"quality"`
	TimePos  float64 `json:"time_pos"`
	Duration float64 `json:"duration"`
}

// VideoInfoResponse contains video metadata and available streaming qualities
type VideoInfoResponse struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	URL            string   `json:"url"`
	Channel        string   `json:"channel"`
	Duration       int64    `json:"duration"`
	DurationString string   `json:"duration_string"`
	Qualities      []string `json:"qualities"`
}

// PlayItem represents the currently playing media
type PlayItem struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Quality string `json:"quality"`
	Mode    string `json:"mode"`
}

// App holds the central application state
type App struct {
	firefoxProfile string
	accountMode    string // "firefox", "guest", "custom"
	customCookies  string
	mpvCmd         *exec.Cmd
	mpvMu          sync.Mutex
	mpvSockPath    string
	currentPlaying *PlayItem

	streamCache map[string]StreamCacheItem
	streamMu    sync.RWMutex

	feedCache map[string]FeedCacheItem
	feedMu    sync.RWMutex

	qualitiesCache map[string][]string
	qualitiesMu    sync.RWMutex

	currentRemux *exec.Cmd
	remuxMu      sync.Mutex

	ytdlpSemaphore chan struct{} // Concurrency limiter to prevent CPU starvation
}
