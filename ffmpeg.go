package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
)

type FFMPEGMacVersionDownloadInfo struct {
	Url  string `json:"url"`
	Size int    `json:"size"`
	Sig  string `json:"sig"`
}

type FFMPEGMacVersionDownload struct {
	SZ  FFMPEGMacVersionDownloadInfo `json:"7z"`
	Zip FFMPEGMacVersionDownloadInfo `json:"zip"`
}

type FFMPEGMacVersion struct {
	Name     string                   `json:"name"`
	Type     string                   `json:"type"`
	Version  string                   `json:"version"`
	Size     int                      `json:"size"`
	Download FFMPEGMacVersionDownload `json:"download"`
}

// Gets the latest version number of FFMPEG for Windows
func GetLatestWinFFMPEGVersion() (string, error) {
	//
	return GetText(FFMPEG_WIN_LATEST_VERSION_URL)
}

// Gets the latest version information of FFMPEG for Mac
func GetLatestMacFFMPEGVersion() (string, error) {
	var err error
	release := FFMPEGMacVersion{}
	data, err := GetBytes(FFMPEG_MAC_INFO_URL)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(data, &version)
	if err != nil {
		return "", err
	}

	return release.Version, nil
}

// Gets the latest version of FFMPEG for Linux
func GetLatestLinuxFFMPEGVersion() (string, error) {
	// TODO: Implement
	return "", nil
}

func IsOutdated(currentVersion, latestVersion string) bool {
	return latestVersion > currentVersion
}

func DownloadFFMPEGWindows(version, dir string) error {
	var err error

	filename := "ffmpeg-essentials_build.7z"
	archivePath := filepath.Join(dir, filename)

	// Download FFMPEG Archive
	url := fmt.Sprintf(FFMPEG_WIN_URL, version)
	logger.Debugf("Downloading ffmpeg from: %s", url)
	err = DownloadFile(url, archivePath)
	if err != nil {
		return fmt.Errorf("Error downloading ffmpeg: %s", err)
	}

	logger.Debug("Writing FFMPEG Version file...")
	err = WriteVersionFile(dir, version)
	if err != nil {
		return fmt.Errorf("Error writing ffmpeg version file: %s", err)
	}

	// Extract the FFMPEG Archive
	logger.Debugf("Unzipping ffmpeg to %s...", dir)
	err = DecompressWFilter(archivePath, dir, fmt.Sprintf("ffmpeg-%s-essentials_build/", version), []string{"bin/ffmpeg.exe"})
	if err != nil {
		return fmt.Errorf("Error unzipping ffmpeg: %s", err)
	}

	return nil
}

func DownloadFFMPEGMac(version, dir string) error {
	var err error

	filename := "ffmpeg.7z"
	archivePath := filepath.Join(dir, filename)

	// Download FFMPEG Archive
	url := fmt.Sprintf(FFMPEG_MAC_VERSION_INFO_URL, version)
	logger.Debugf("Downloading ffmpeg from: %s", url)
	err = DownloadFile(url, archivePath)
	if err != nil {
		return fmt.Errorf("Error downloading ffmpeg: %s", err)
	}

	logger.Debug("Writing FFMPEG Version file...")
	err = WriteVersionFile(dir, version)
	if err != nil {
		return fmt.Errorf("Error writing ffmpeg version file: %s", err)
	}

	// Extract the FFMPEG Archive
	logger.Debugf("Unzipping ffmpeg to %s...", dir)
	err = DecompressWFilter(archivePath, dir, fmt.Sprintf("ffmpeg-%s/", version), []string{"ffmpeg"})
	if err != nil {
		return fmt.Errorf("Error unzipping ffmpeg: %s", err)
	}

	return nil
}

func DownloadFFMPEGLinux(version, dir string) error {
	// TODO: At some point we should look into completing this
	return fmt.Errorf("It looks like you're running linux, this script does not support installing FFMPEG binaries for linux yet :(\nPlease install FFMPEG via your systems package manager and re-run the script")
}

// Function to get the latest version of FFMPEG for current platform
func GetLatestFFMPEGVersion() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return GetLatestWinFFMPEGVersion()
	case "darwin":
		return GetLatestMacFFMPEGVersion()
	case "linux":
		return GetLatestLinuxFFMPEGVersion()
	}

	return "", fmt.Errorf("Unsupported OS: %s", runtime.GOOS)
}

// Function to download the latest version of FFMPEG for current platform
func DownloadFFMPEG(version, dir string) error {
	switch runtime.GOOS {
	case "windows":
		return DownloadFFMPEGWindows(version, dir)
	case "darwin":
		return DownloadFFMPEGMac(version, dir)
	case "linux":
		return DownloadFFMPEGLinux(version, dir)
	}

	return fmt.Errorf("Unsupported OS: %s", runtime.GOOS)
}
