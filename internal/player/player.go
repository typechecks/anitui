package player

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var playerPriority = []string{"mpv", "iina", "vlc", "haruna"}

func DetectPlayer() string {
	if envPlayer := os.Getenv("ANITUI_PLAYER"); envPlayer != "" {
		return envPlayer
	}

	for _, player := range playerPriority {
		if path := findPlayer(player); path != "" {
			return path
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

	base := strings.ToLower(fileName(player))
	switch {
	case strings.Contains(base, "mpv"):
		args = []string{
			"--really-quiet",
			"--no-terminal",
			"--force-window=yes",
			"--http-header-fields=Referer: https://allmanga.to",
			url,
		}
	case strings.Contains(base, "iina"):
		args = []string{"--no-stdin", "--keep-running", url}
	case strings.Contains(base, "vlc"):
		args = []string{"--quiet", "--play-and-exit", "--http-referrer=https://allmanga.to", url}
	case strings.Contains(base, "haruna"):
		args = []string{url}
	default:
		args = []string{url}
	}

	cmd := exec.Command(player, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
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

func findPlayer(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}

	for _, p := range defaultPaths(name) {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func defaultPaths(name string) []string {
	switch runtime.GOOS {
	case "windows":
		return windowsPaths(name)
	case "darwin":
		return darwinPaths(name)
	default:
		return linuxPaths(name)
	}
}

func windowsPaths(name string) []string {
	switch name {
	case "mpv":
		return []string{
			`C:\Program Files\mpv\mpv.exe`,
			`C:\Program Files (x86)\mpv\mpv.exe`,
			`C:\mpv\mpv.exe`,
		}
	case "vlc":
		return []string{
			`C:\Program Files\VideoLAN\VLC\vlc.exe`,
			`C:\Program Files (x86)\VideoLAN\VLC\vlc.exe`,
		}
	}
	return nil
}

func darwinPaths(name string) []string {
	switch name {
	case "mpv":
		return []string{
			"/Applications/mpv.app/Contents/MacOS/mpv",
			"/opt/homebrew/bin/mpv",
			"/usr/local/bin/mpv",
		}
	case "iina":
		return []string{
			"/Applications/IINA.app/Contents/MacOS/iina-cli",
		}
	case "vlc":
		return []string{
			"/Applications/VLC.app/Contents/MacOS/VLC",
		}
	}
	return nil
}

func linuxPaths(name string) []string {
	switch name {
	case "mpv":
		return []string{
			"/usr/bin/mpv",
			"/usr/local/bin/mpv",
			"/snap/bin/mpv",
		}
	case "vlc":
		return []string{
			"/usr/bin/vlc",
			"/usr/local/bin/vlc",
			"/snap/bin/vlc",
		}
	case "haruna":
		return []string{
			"/usr/bin/haruna",
			"/usr/local/bin/haruna",
			"/snap/bin/haruna",
			"/var/lib/flatpak/exports/bin/haruna",
		}
	}
	return nil
}

func fileName(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	if i := strings.LastIndex(path, `\`); i >= 0 {
		return path[i+1:]
	}
	return path
}
