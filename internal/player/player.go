package player

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var playerPriority = []string{"mpv", "iina", "vlc", "haruna"}

func DetectPlayer() string {
	if envPlayer := os.Getenv("ANITUI_PLAYER"); envPlayer != "" {
		return envPlayer
	}

	for _, player := range playerPriority {
		if playerAvailable(player) {
			return player
		}
	}
	return ""
}

func Play(url string) error {
	player := DetectPlayer()
	if player == "" {
		return fmt.Errorf("no supported video player found. Please install mpv, vlc, iina (macOS), or haruna")
	}

	var args []string

	switch player {
	case "mpv":
		args = []string{
			"--really-quiet",
			"--no-terminal",
			"--force-window=yes",
			"--http-header-fields=Referer: https://allmanga.to",
			url,
		}
	case "iina":
		args = []string{"--no-stdin", "--keep-running", url}
	case "vlc":
		args = []string{"--quiet", "--play-and-exit", "--http-referrer=https://allmanga.to", url}
	case "haruna":
		args = []string{url}
	}

	cmd := exec.Command(player, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", player, err)
	}

	go func() {
		cmd.Wait()
	}()

	return nil
}

func playerAvailable(name string) bool {
	// First check if it's in the PATH
	if _, err := exec.LookPath(name); err == nil {
		return true
	}

	var commonPaths []string

	switch {
	case runtime.GOOS == "windows":
		switch name {
		case "vlc":
			commonPaths = []string{
				`C:\Program Files\VideoLAN\VLC\vlc.exe`,
				`C:\Program Files (x86)\VideoLAN\VLC\vlc.exe`,
			}
		case "mpv":
			commonPaths = []string{
				`C:\Program Files\mpv\mpv.exe`,
				`C:\mpv\mpv.exe`,
			}
		}

	case runtime.GOOS == "darwin":
		switch name {
		case "vlc":
			commonPaths = []string{
				"/Applications/VLC.app/Contents/MacOS/VLC",
			}
		case "mpv":
			commonPaths = []string{
				"/Applications/mpv.app/Contents/MacOS/mpv",
				"/opt/homebrew/bin/mpv",
				"/usr/local/bin/mpv",
			}
		case "iina":
			commonPaths = []string{
				"/Applications/IINA.app/Contents/MacOS/iina-cli",
			}
		}
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}

	return false
}
