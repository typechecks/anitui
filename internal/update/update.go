package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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
const downloadTimeout = 10 * time.Minute

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

func fetchRelease() (*release, error) {
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
	return &rel, nil
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
	rel, err := fetchRelease()
	if err != nil {
		return err
	}

	assetName := assetNameForPlatform()
	var target *asset
	for i := range rel.Assets {
		if rel.Assets[i].Name == assetName {
			target = &rel.Assets[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("no release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp("", "anitui-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	fmt.Printf("\r\033[KDownloading v%s\n", version)
	err = downloadWithProgress(tmpFile, target.BrowserDownloadURL, target.Size)
	tmpFile.Close()
	if err != nil {
		return err
	}

	newExe := exe + ".new"
	err = extractBinary(tmpPath, newExe)
	if err != nil {
		return fmt.Errorf("extracting binary: %w", err)
	}
	defer os.Remove(newExe)

	if runtime.GOOS != "windows" {
		if info, err := os.Stat(exe); err == nil {
			os.Chmod(newExe, info.Mode())
		} else {
			os.Chmod(newExe, 0755)
		}
	}

	oldExe := exe + ".old"
	os.Remove(oldExe)

	if err := os.Rename(exe, oldExe); err != nil {
		return fmt.Errorf("renaming current binary: %w", err)
	}

	if err := os.Rename(newExe, exe); err != nil {
		os.Rename(oldExe, exe)
		return fmt.Errorf("installing new binary: %w", err)
	}

	return nil
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

func assetNameForPlatform() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	switch osName {
	case "windows":
		return fmt.Sprintf("anitui_windows_%s.zip", arch)
	case "darwin":
		return fmt.Sprintf("anitui_darwin_%s.tar.gz", arch)
	default:
		return fmt.Sprintf("anitui_%s_%s.tar.gz", osName, arch)
	}
}

func downloadWithProgress(w io.Writer, url string, totalSize int64) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "anitui")

	client := &http.Client{Timeout: downloadTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	size := resp.ContentLength
	if size <= 0 {
		size = totalSize
	}

	var downloaded int64
	buf := make([]byte, 32*1024)
	lastUpdate := time.Now()

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)
			if time.Since(lastUpdate) > 80*time.Millisecond || downloaded >= size {
				renderProgress(downloaded, size)
				lastUpdate = time.Now()
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	renderProgress(downloaded, size)
	fmt.Print("\n")
	return nil
}

func renderProgress(downloaded, total int64) {
	const barWidth = 30
	if total <= 0 {
		fmt.Printf("\r  %s", formatBytes(downloaded))
		return
	}
	pct := float64(downloaded) / float64(total)
	filled := int(pct * barWidth)
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	fmt.Printf("\r  %s %3.0f%% (%s/%s)",
		bar, pct*100,
		formatBytes(downloaded),
		formatBytes(total),
	)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func extractBinary(archivePath, destPath string) error {
	if strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
		return extractFromZip(archivePath, destPath)
	}
	return extractFromTarGz(archivePath, destPath)
}

func extractFromTarGz(tgzPath, destPath string) error {
	f, err := os.Open(tgzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		out, err := os.Create(destPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	}
	return fmt.Errorf("binary not found in tar.gz archive")
}

func extractFromZip(zipPath, destPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return err
		}
		out.Close()
		rc.Close()
		return nil
	}
	return fmt.Errorf("binary not found in zip archive")
}
