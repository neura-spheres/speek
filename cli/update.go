package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const repo = "neura-spheres/speek"

// Update checks for a newer release and replaces the running binary if one is found.
func Update(currentVersion string) int {
	fmt.Println("Checking for updates...")

	latest, downloadURL, err := fetchLatestRelease()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not reach GitHub: %v\n", err)
		return 1
	}

	// Normalize: strip leading 'v' for comparison
	current := strings.TrimPrefix(currentVersion, "v")
	latestClean := strings.TrimPrefix(latest, "v")

	if current == latestClean {
		fmt.Printf("Already up to date (Speek %s)\n", currentVersion)
		return 0
	}

	fmt.Printf("New version available: %s → %s\n", currentVersion, latest)
	fmt.Println("Downloading...")

	// Find the path of the running binary
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not locate current binary: %v\n", err)
		return 1
	}

	// Download new binary to a temp file next to the current one
	tmp := exe + ".new"
	if err := downloadFile(downloadURL, tmp); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		return 1
	}

	// Make it executable on Unix
	if runtime.GOOS != "windows" {
		os.Chmod(tmp, 0755)
	}

	// Replace the current binary
	if err := os.Rename(tmp, exe); err != nil {
		// On Windows the running binary is locked — write a helper batch script
		bat := exe + "_update.bat"
		script := fmt.Sprintf("@echo off\ntimeout /t 1 >nul\nmove /y \"%s\" \"%s\"\ndel \"%s\"\n", tmp, exe, bat)
		os.WriteFile(bat, []byte(script), 0644)
		fmt.Println("Update downloaded. Run the following to finish (close this terminal first):")
		fmt.Printf("  %s\n", bat)
		return 0
	}

	fmt.Printf("Speek updated to %s\n", latest)

	// Also update the VS Code extension
	fmt.Println("Updating VS Code extension...")
	InstallVSCode(nil, true)

	return 0
}

func fetchLatestRelease() (tag string, downloadURL string, err error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	// Pick the right asset for this OS/arch
	want := assetName()
	for _, a := range release.Assets {
		if a.Name == want {
			return release.TagName, a.BrowserDownloadURL, nil
		}
	}
	return "", "", fmt.Errorf("no asset named %q in release %s", want, release.TagName)
}

func assetName() string {
	arch := "amd64"
	switch runtime.GOARCH {
	case "arm64":
		arch = "arm64"
	}
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("speek-windows-%s.exe", arch)
	case "darwin":
		return fmt.Sprintf("speek-macos-%s", arch)
	default:
		return fmt.Sprintf("speek-linux-%s", arch)
	}
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
