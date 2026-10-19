package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// sendMpvCommand sends a JSON IPC command to the active MPV unix socket
func (a *App) sendMpvCommand(cmd []interface{}) (map[string]interface{}, error) {
	if a.mpvSockPath == "" || runtime.GOOS == "windows" {
		return nil, fmt.Errorf("no socket configured")
	}
	conn, err := net.DialTimeout("unix", a.mpvSockPath, 250*time.Millisecond)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	payload, err := json.Marshal(map[string]interface{}{"command": cmd})
	if err != nil {
		return nil, err
	}
	payload = append(payload, '\n')

	_ = conn.SetDeadline(time.Now().Add(400 * time.Millisecond))
	if _, err := conn.Write(payload); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	for {
		respLine, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}

		var res map[string]interface{}
		if err := json.Unmarshal(respLine, &res); err == nil {
			if _, isEvent := res["event"]; isEvent && res["error"] == nil {
				continue
			}
			return res, nil
		}
	}
}

// startMpv launches an ultra-lightweight MPV player tailored for ThinkPad X201 hardware
func (a *App) startMpv(req PlayRequest) error {
	a.stopCurrentPlayback()
	a.stopCurrentRemux()

	var mpvArgs []string
	if a.accountMode == "firefox" && a.firefoxProfile != "" {
		mpvArgs = append(mpvArgs, "--ytdl-raw-options=cookies-from-browser=firefox:"+a.firefoxProfile)
	} else if a.accountMode == "custom" && a.customCookies != "" {
		mpvArgs = append(mpvArgs, "--ytdl-raw-options=cookies="+a.customCookies)
	}

	maxHeight := "480"
	if req.Quality == "360" || req.Quality == "720" {
		maxHeight = req.Quality
	}

	_ = os.Remove(a.mpvSockPath)

	if req.Mode == "audio" {
		mpvArgs = append(mpvArgs,
			"--no-video",
			"--ytdl-format=bestaudio[acodec^=mp4a]/bestaudio/best",
			"--input-ipc-server="+a.mpvSockPath,
			fmt.Sprintf("--title=FX Tube Audio: %s", req.Title),
			req.URL,
		)
	} else {
		// Hardware-safe MPV configuration for ThinkPad X201:
		// - hwdec=no prevents GPU driver lockup on Ironlake
		// - vd-lavc-threads=4 utilizes all 4 threads of Core i7 M 620
		// - demuxer-max-bytes=24MiB and demuxer-readahead-secs=10 saves memory & bandwidth
		// - autofit and geometry provide a clean floating player
		formatString := fmt.Sprintf("bestvideo[vcodec^=avc1][height<=?%s][fps<=?30]+bestaudio[acodec^=mp4a]/bestvideo[height<=?%s][fps<=?30]+bestaudio/best[height<=?%s]/best", maxHeight, maxHeight, maxHeight)
		mpvArgs = append(mpvArgs,
			"--hwdec=no",
			"--vd-lavc-threads=4",
			"--vo=gpu,wlshm,x11",
			"--demuxer-max-bytes=24MiB",
			"--demuxer-readahead-secs=10",
			"--audio-display=no",
			"--input-ipc-server="+a.mpvSockPath,
			"--autofit=960x540",
			"--geometry=65%x65%",
			"--ytdl-format="+formatString,
			fmt.Sprintf("--title=FX Tube: %s", req.Title),
			req.URL,
		)
	}

	a.mpvMu.Lock()
	a.mpvCmd = exec.Command("mpv", mpvArgs...)
	a.currentPlaying = &PlayItem{
		URL:     req.URL,
		Title:   req.Title,
		Quality: maxHeight,
		Mode:    req.Mode,
	}
	if err := a.mpvCmd.Start(); err != nil {
		a.mpvMu.Unlock()
		return err
	}
	a.mpvMu.Unlock()

	go func() {
		_ = a.mpvCmd.Wait()
		a.mpvMu.Lock()
		a.mpvCmd = nil
		a.currentPlaying = nil
		a.mpvMu.Unlock()
		_ = os.Remove(a.mpvSockPath)
	}()

	return nil
}

// stopCurrentPlayback gracefully terminates any active MPV session
func (a *App) stopCurrentPlayback() {
	a.mpvMu.Lock()
	defer a.mpvMu.Unlock()
	if a.mpvCmd != nil && a.mpvCmd.Process != nil {
		_, _ = a.sendMpvCommand([]interface{}{"quit"})
		time.Sleep(40 * time.Millisecond)
		_ = a.mpvCmd.Process.Kill()
		a.mpvCmd = nil
		a.currentPlaying = nil
	}
	_ = os.Remove(a.mpvSockPath)
}

// getPlayerStatus returns the live playback state of the MPV engine
func (a *App) getPlayerStatus() PlayerStatusResponse {
	a.mpvMu.Lock()
	playing := a.mpvCmd != nil && a.mpvCmd.Process != nil && a.currentPlaying != nil
	cur := a.currentPlaying
	a.mpvMu.Unlock()

	if !playing || cur == nil {
		return PlayerStatusResponse{Playing: false}
	}

	status := PlayerStatusResponse{
		Playing: true,
		Title:   cur.Title,
		URL:     cur.URL,
		Quality: cur.Quality,
	}

	if res, err := a.sendMpvCommand([]interface{}{"get_property", "pause"}); err == nil {
		if paused, ok := res["data"].(bool); ok {
			status.Paused = paused
		}
	}
	if res, err := a.sendMpvCommand([]interface{}{"get_property", "time-pos"}); err == nil {
		if pos, ok := res["data"].(float64); ok {
			status.TimePos = pos
		}
	}
	if res, err := a.sendMpvCommand([]interface{}{"get_property", "duration"}); err == nil {
		if dur, ok := res["data"].(float64); ok {
			status.Duration = dur
		}
	}

	return status
}

// executePlayerCommand dispatches user commands to the MPV IPC socket
func (a *App) executePlayerCommand(command string) {
	switch command {
	case "toggle_pause":
		_, _ = a.sendMpvCommand([]interface{}{"cycle", "pause"})
	case "seek_forward":
		_, _ = a.sendMpvCommand([]interface{}{"seek", 10})
	case "seek_backward":
		_, _ = a.sendMpvCommand([]interface{}{"seek", -10})
	case "stop":
		a.stopCurrentPlayback()
	}
}
