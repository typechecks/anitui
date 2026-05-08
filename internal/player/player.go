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
	if runtime.GOOS == "windows" {
		commonPaths := []string{
			`C:\Program Files\` + name + `\`,
			`C:\Program Files (x86)\` + name + `\`,
		}
		for _, p := range commonPaths {
			if _, err := os.Stat(p + name + ".exe"); err == nil {
				return true
			}
		}
	}

	_, err := exec.LookPath(name)
	return err == nil
}
