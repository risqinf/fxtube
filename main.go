package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const DefaultPort = "18201"

func isServerRunning(port string) bool {
	client := http.Client{Timeout: 400 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:" + port + "/api/ping")
	if err == nil && resp.StatusCode == 200 {
		var b bytes.Buffer
		_, _ = io.Copy(&b, resp.Body)
		resp.Body.Close()
		return b.String() == "pong"
	}
	return false
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			log.Println("FX Tube v1.4.2 (Ultra Lightweight YouTube Client)")
			return
		}
	}

	port := DefaultPort

	// If instance is already running, just launch the GUI and exit immediately
	if isServerRunning(port) {
		launchGUI(port)
		return
	}

	home, _ := os.UserHomeDir()
	profile := findFirefoxProfile()
	customCookies := filepath.Join(home, ".config", "fxtube", "cookies.txt")
	mpvSock := filepath.Join(os.TempDir(), "fxtube-mpv.sock")

	log.Printf("[FX Tube] Inisialisasi: Firefox Profile: %s\n", profile)

	app := &App{
		firefoxProfile: profile,
		accountMode:    "guest",
		customCookies:  customCookies,
		mpvSockPath:    mpvSock,
		streamCache:    make(map[string]StreamCacheItem),
		feedCache:      make(map[string]FeedCacheItem),
		qualitiesCache: make(map[string][]string),
		commentsCache:  make(map[string][]CommentItem),
		ytdlpSemaphore: make(chan struct{}, 2), // Cap concurrent yt-dlp workers to 2 for CPU stability
	}

	// Load stored Google profile if available (defaults to Guest)
	app.loadProfile()

	handler := app.setupRoutes()

	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	localIP := getLocalIP()
	log.Printf("[FX Tube] Server Desktop: http://127.0.0.1:%s\n", port)
	log.Printf("[FX Tube] Akses Mobile / Android (LAN): http://%s:%s\n", localIP, port)

	// Ignore SIGHUP so backend persists cleanly when terminal/launcher exits
	signal.Ignore(syscall.SIGHUP)

	// Clean shutdown hook on SIGINT or SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("[FX Tube] Menutup proses...")
		app.stopCurrentPlayback()
		app.stopCurrentRemux()
		_ = listener.Close()
		os.Exit(0)
	}()

	// Launch GUI window and verify yt-dlp binary presence
	go func() {
		time.Sleep(250 * time.Millisecond)
		launchGUI(port)
		ensureYtDlp()
	}()

	if err := http.Serve(listener, handler); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
