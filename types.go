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
	IsShorts       bool        `json:"is_shorts"`
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

	googleProfile GoogleProfile
	profileMu     sync.RWMutex

	commentsCache map[string][]CommentItem
	commentsMu    sync.RWMutex

	downloadDir string
	downloads   map[string]*DownloadItem
	downloadMu  sync.RWMutex
}

// GoogleProfile represents the logged in Google/YouTube user
type GoogleProfile struct {
	Name       string `json:"name"`
	Handle     string `json:"handle"`
	Avatar     string `json:"avatar"`
	ChannelURL string `json:"channel_url"`
	IsLoggedIn bool   `json:"is_logged_in"`
}

// CommentItem represents a single YouTube video comment
type CommentItem struct {
	Author string `json:"author"`
	Avatar string `json:"avatar"`
	Text   string `json:"text"`
	Time   string `json:"time"`
	Likes  string `json:"likes"`
}

// CreateCommentRequest payload for posting comments
type CreateCommentRequest struct {
	VideoURL string `json:"video_url"`
	VideoID  string `json:"video_id"`
	Text     string `json:"text"`
}

// GoogleLoginRequest payload for logging in to Google / YouTube
type GoogleLoginRequest struct {
	Type       string `json:"type"`        // "handle", "channel_url", "browser", "manual"
	Handle     string `json:"handle"`      // e.g. "@risqinf"
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	ChannelURL string `json:"channel_url"`
	Browser    string `json:"browser"`     // "firefox", "chrome", "brave"
}

// DownloadItem represents a downloaded or active download video
type DownloadItem struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Channel      string  `json:"channel"`
	Thumbnail    string  `json:"thumbnail"`
	Quality      string  `json:"quality"`
	FilePath     string  `json:"file_path"`
	FileName     string  `json:"file_name"`
	FileSize     int64   `json:"file_size"`
	FileSizeText string  `json:"file_size_text"`
	Progress     float64 `json:"progress"`
	Status       string  `json:"status"` // "downloading", "completed", "error"
	Error        string  `json:"error,omitempty"`
	CreatedAt    int64   `json:"created_at"`
}

// StartDownloadRequest payload for starting a video download
type StartDownloadRequest struct {
	URL       string `json:"url"`
	ID        string `json:"id"`
	Title     string `json:"title"`
	Channel   string `json:"channel"`
	Thumbnail string `json:"thumbnail"`
	Quality   string `json:"quality"`
}

// DownloadConfigRequest payload for configuring download directory
type DownloadConfigRequest struct {
	DownloadDir string `json:"download_dir"`
}
