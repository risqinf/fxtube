package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

// stopCurrentRemux terminates any existing FFmpeg background remuxing job on app shutdown
func (a *App) stopCurrentRemux() {
	a.remuxMu.Lock()
	defer a.remuxMu.Unlock()
	if a.currentRemux != nil && a.currentRemux.Process != nil {
		_ = a.currentRemux.Process.Kill()
		a.currentRemux = nil
	}
}

// handleStream serves the direct or remuxed video stream for embedded HTML5 playback
func (a *App) handleStream(w http.ResponseWriter, r *http.Request) {
	targetURL := normalizeYouTubeURL(r.URL.Query().Get("url"))
	quality := r.URL.Query().Get("quality")
	start := strings.TrimSpace(r.URL.Query().Get("start"))
	if quality == "" {
		quality = "480"
	}
	if targetURL == "" {
		http.Error(w, "Missing url parameter", http.StatusBadRequest)
		return
	}

	item, err := a.resolveStreamURLs(targetURL, quality)
	if err != nil {
		log.Printf("[FX Tube Stream] Gagal resolve stream: %v\n", err)
		http.Error(w, "Failed to resolve stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 1. Single combined stream: proxy directly if standard MP4/WebM with HTTP Range byte-seeking support
	// If it is HLS / .m3u8, it MUST be remuxed by FFmpeg because desktop Chromium cannot decode m3u8 natively
	isHLS := strings.Contains(item.VideoURL, ".m3u8") || strings.Contains(item.VideoURL, "manifest") || strings.Contains(item.VideoURL, "live")
	if item.IsCombined && !isHLS {
		proxyDirectStream(w, r, item.VideoURL)
		return
	}

	// 2. Remux on the fly via local FFmpeg per connection (handles separate video+audio or HLS live streams)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ua := "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0"
	var ffmpegArgs []string
	ffmpegArgs = append(ffmpegArgs, "-y")
	if start != "" && start != "0" {
		ffmpegArgs = append(ffmpegArgs, "-ss", start)
	}
	ffmpegArgs = append(ffmpegArgs,
		"-user_agent", ua,
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-analyzeduration", "2M",
		"-probesize", "2M",
		"-i", item.VideoURL,
	)

	// If separate audio stream exists, add it as second input
	if item.AudioURL != "" && item.AudioURL != item.VideoURL && !isHLS {
		if start != "" && start != "0" {
			ffmpegArgs = append(ffmpegArgs, "-ss", start)
		}
		ffmpegArgs = append(ffmpegArgs,
			"-user_agent", ua,
			"-reconnect", "1",
			"-reconnect_streamed", "1",
			"-reconnect_delay_max", "5",
			"-analyzeduration", "2M",
			"-probesize", "2M",
			"-i", item.AudioURL,
			"-c:v", "copy",
			"-c:a", "aac",
			"-b:a", "128k",
			"-f", "mp4",
			"-movflags", "frag_keyframe+empty_moov+default_base_moof",
			"-loglevel", "error",
			"pipe:1",
		)
	} else {
		// Single multiplexed stream or HLS live stream
		ffmpegArgs = append(ffmpegArgs,
			"-c:v", "copy",
			"-c:a", "aac",
			"-b:a", "128k",
			"-f", "mp4",
			"-movflags", "frag_keyframe+empty_moov+default_base_moof",
			"-loglevel", "error",
			"pipe:1",
		)
	}

	cmd := exec.CommandContext(ctx, getFfmpegPath(), ffmpegArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, "Remux error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := cmd.Start(); err != nil {
		http.Error(w, "FFmpeg start error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	_, _ = io.Copy(w, stdout)
	_ = cmd.Wait()
}

// proxyDirectStream transparently proxies single-stream MP4 with HTTP Range byte-seeking support
func proxyDirectStream(w http.ResponseWriter, r *http.Request, targetURL string) {
	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Forward Range header from browser
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		lk := strings.ToLower(k)
		if lk == "connection" || lk == "keep-alive" || lk == "transfer-encoding" || lk == "upgrade" {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
