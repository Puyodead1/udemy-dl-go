package main

import (
	"fmt"
	"os"
)

func FFMPEGCheck() (bool, error) {
	Info("Checking FFMPEG...")

	// check if ffmpeg is installed externally
	exists := CommandExists("ffmpeg")
	if exists {
		Success("FFMPEG appears to be installed already, probably via a package manager.")
		return true, nil
	}

	var err error

	// Ensure directory exists
	err = EnsureDirExist(FFMPEG_BIN_DIRECTORY)
	if err != nil {
		return false, fmt.Errorf("Error creating ffmpeg directory: %s", err)
	}

	// get the latest version of ffmpeg
	latestVersion, err := GetLatestFFMPEGVersion()
	if err != nil {
		return false, fmt.Errorf("Error getting latest ffmpeg version: %s", err)
	}

	// check for version file
	versionFileExists := VersionFileExists(FFMPEG_BIN_DIRECTORY)

	// existing install
	if versionFileExists {
		// read version from file
		currentVersion, err := ReadVersionFile(FFMPEG_BIN_DIRECTORY)
		if err != nil {
			return false, fmt.Errorf("Failed to read FFMPEG version file: %s", err)
		}

		// compare versions
		if IsOutdated(currentVersion, latestVersion) {
			// outdated
			Warningf("FFMPEG is outdated, current version: %s, latest version: %s", currentVersion, latestVersion)
			// remove old directory
			err = os.RemoveAll(FFMPEG_BIN_DIRECTORY)
			if err != nil {
				return false, err
			}
			// remake the directory
			err = EnsureDirExist(FFMPEG_BIN_DIRECTORY)
			if err != nil {
				return false, fmt.Errorf("Error creating ffmpeg directory: %s", err)
			}
			// download and extract
			err = DownloadFFMPEG(latestVersion, FFMPEG_BIN_DIRECTORY)
			if err != nil {
				return false, err
			}
		} else {
			// up to date
			Successf("FFMPEG is up to date, current version: %s, latest version: %s", currentVersion, latestVersion)
		}
	} else {
		// no existing install
		Warning("FFMPEG not found, downloading...")
		err = DownloadFFMPEG(latestVersion, FFMPEG_BIN_DIRECTORY)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func GetLatestShakaPackagerVersion() (string, int64, error) {
	release, err := GetLatestRelease("shaka-project", "shaka-packager")
	if err != nil {
		// Errorf("Error getting latest shaka release: %s", err)
		return "", -1, err
	}

	return release.GetTagName(), release.GetID(), nil
}

func ShakaPackager() (bool, error) {
	// versionString, versionID, err := GetLatestShakaPackagerVersion()
	// if err != nil {
	// 	// Errorf("Error getting latest shaka release: %s", err)
	// 	return false, errors.New(fmt.Sprintf("Error getting latest shaka release: %s", err))
	// }

	// Info("Latest Shaka Release: ", versionString)

	// assets, err := GetReleaseAssets("shaka-project", "shaka-packager", versionID)
	// if err != nil {
	// 	return false, errors.New(fmt.Sprintf("Error getting shaka release assets: %s", err))
	// }

	// Info("Shaka Release Assets: ", assets)

	return true, nil
}

// -----------------------------

// Main dependency checking function, runs checks for each dependency depending on the current OS
func RunDependencyCheck() (bool, bool, bool, bool, error) {
	Info("Starting dependency check")

	var err error
	var ffmpegStatus bool = false
	var aria2Status bool = false
	var ytdlpStatus bool = false
	var shakaStatus bool = false

	// ensure the main bin directory exists
	// ensure bin dir exists
	err = EnsureDirExist("bin")
	if err != nil {
		return ffmpegStatus, aria2Status, ytdlpStatus, shakaStatus, fmt.Errorf("Error creating ffmpeg directory: %s", err)
	}

	// tries to make bin directory whether it exists or not
	err = EnsureDirExist("bin")
	if err != nil {
		return ffmpegStatus, aria2Status, ytdlpStatus, shakaStatus, fmt.Errorf("Error creating bin directory: %s", err)
	}

	// TODO: add other dependencies (aria2, yt-dlp, shaka-packager)
	ffmpegStatus, err = FFMPEGCheck()

	if err != nil {
		return ffmpegStatus, aria2Status, ytdlpStatus, shakaStatus, err
	}

	return ffmpegStatus, aria2Status, ytdlpStatus, shakaStatus, nil
}
