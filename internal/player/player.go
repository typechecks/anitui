package player

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var playerPriority = []string{"mpv", "vlc", "haruna"}

func DetectPlayer() string {
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
		return fmt.Errorf("no supported video player found. Please install mpv, vlc, or haruna")
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

	if runtime.GOOS == "windows" {
		var commonPaths []string
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

		for _, p := range commonPaths {
			if _, err := os.Stat(p); err == nil {
				return true
			}
		}
	}

	return false
}
