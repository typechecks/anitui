package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const repoAPI = "https://api.github.com/repos/typechecks/anitui/releases/latest"
const installScriptURL = "https://raw.githubusercontent.com/typechecks/anitui/main/scripts/install.sh"
const installScriptPS1URL = "https://raw.githubusercontent.com/typechecks/anitui/main/scripts/install.ps1"

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

func Cleanup() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	os.Remove(exe + ".old")
}

var cachedRelease *release

func fetchRelease() (*release, error) {
	if cachedRelease != nil {
		return cachedRelease, nil
	}

	req, err := http.NewRequest("GET", repoAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "anitui")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	cachedRelease = &rel
	return cachedRelease, nil
}

func Check(currentVersion string) (string, error) {
	rel, err := fetchRelease()
	if err != nil {
		return "", err
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	if compareVersions(latest, currentVersion) > 0 {
		return latest, nil
	}
	return "", nil
}

func compareVersions(a, b string) int {
	if a == b {
		return 0
	}
	if a == "dev" {
		return -1
	}
	if b == "dev" {
		return 1
	}

	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aNum, bNum int
		if i < len(aParts) {
			aNum, _ = strconv.Atoi(strings.TrimSpace(aParts[i]))
		}
		if i < len(bParts) {
			bNum, _ = strconv.Atoi(strings.TrimSpace(bParts[i]))
		}
		if aNum > bNum {
			return 1
		}
		if aNum < bNum {
			return -1
		}
	}
	return 0
}

func IsWritable() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	dir := filepath.Dir(exe)
	testFile := filepath.Join(dir, ".anitui-write-test")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)
	return true
}

func Apply(version string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exe)

	if runtime.GOOS == "windows" {
		return applyWindows(exeDir)
	}
	return applyUnix(exeDir)
}

func applyUnix(exeDir string) error {
	tmpScript, err := os.CreateTemp("", "anitui-install-")
	if err != nil {
		return err
	}
	tmpPath := tmpScript.Name()
	defer os.Remove(tmpPath)

	req, err := http.NewRequest("GET", installScriptURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "anitui")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download install script: status %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmpScript, resp.Body); err != nil {
		tmpScript.Close()
		return err
	}
	tmpScript.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	cmd := exec.Command("sh", tmpPath, "--dir", exeDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func applyWindows(exeDir string) error {
	tmpScript, err := os.CreateTemp("", "anitui-install-*.ps1")
	if err != nil {
		return err
	}
	tmpPath := tmpScript.Name()
	defer os.Remove(tmpPath)

	req, err := http.NewRequest("GET", installScriptPS1URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "anitui")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download install script: status %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmpScript, resp.Body); err != nil {
		tmpScript.Close()
		return err
	}
	tmpScript.Close()

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", tmpPath, "-InstallDir", exeDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func Relaunch() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	os.Exit(0)
}
