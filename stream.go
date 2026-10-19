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
	targetURL := r.URL.Query().Get("url")
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

	// 1. Single combined stream: proxy directly with HTTP Range byte-seeking support
	if item.IsCombined {
		proxyDirectStream(w, r, item.VideoURL)
		return
	}

	// 2. Separate video and audio: remux on the fly via local FFmpeg per connection
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var ffmpegArgs []string
	ffmpegArgs = append(ffmpegArgs, "-y")
	if start != "" && start != "0" {
		ffmpegArgs = append(ffmpegArgs, "-ss", start)
	}
	ffmpegArgs = append(ffmpegArgs,
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", item.VideoURL,
	)
	if start != "" && start != "0" {
		ffmpegArgs = append(ffmpegArgs, "-ss", start)
	}
	ffmpegArgs = append(ffmpegArgs,
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", item.AudioURL,
		"-c", "copy",
		"-f", "mp4",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof",
		"-loglevel", "error",
		"pipe:1",
	)

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
