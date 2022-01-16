package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

var UpdaterClient = &http.Client{Timeout: time.Second * 10}

func HandleBinaryUpdate(depName string, latestRelease Release, asset Asset, depFolder string, versionFilePath string) bool {
	archiveFilePath := path.Join(depFolder, asset.Name)

	// update binary
	if !Exists(archiveFilePath) {
		logger.Debugf("Downloading %s to %s\n", depName, archiveFilePath)

		// download binary
		err := DownloadFileNew(asset.Browser_Download_URL, archiveFilePath)
		if err != nil {
			logger.Errorf("Error downloading %s: %s\n", depName, err)
			return false
		}
	} else {
		logger.Debugf("Found existing %s archive: %s\n", depName, archiveFilePath)
	}

	if depName != "ytdlp" && depName != "shaka-packager" {
		logger.Debugf("Unzipping %s", depName)

		// unzip binary
		var files []string
		var err error
		if strings.HasSuffix(asset.Name, ".bz2") {
			files, err = DecompressBZIP(archiveFilePath, depFolder)
		} else {
			files, err = Unzip(archiveFilePath, depFolder)
		}
		if err != nil {
			logger.Errorf("Error unzipping %s archive: %s\n", depName, err)
			return false
		}

		logger.Debugf("%s unzipped, Moving binaries", depName)

		// move executable files to bin folder
		for _, v := range files {
			if strings.HasSuffix(v, "bin/aria2c") || strings.HasSuffix(v, "ffmpeg") || strings.HasSuffix(v, "aria2c.exe") || strings.HasSuffix(v, "ffmpeg.exe") {
				logger.Debugf("Moving executable file %s to %s\n", v, depFolder)
				err := os.Rename(v, path.Join(depFolder, path.Base(v)))
				if err != nil {
					logger.Errorf("Error moving executable file %s to %s: %s\n", v, depFolder, err)
					return false
				}
			}
		}

		// basename := path.Base(archiveFilePath)
		tmpFolderPath := files[0]

		logger.Debugf("Removing temp folder: %s\n", tmpFolderPath)

		err1 := os.RemoveAll(tmpFolderPath)
		if err1 != nil {
			logger.Errorf("Error removing temp folder: %s\n", err1)
			return false
		}
	}

	// rename shaka-packger binary
	if depName == "shaka-packager" {
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			os.Rename(archiveFilePath, path.Join(depFolder, "shaka-packager"))
		} else if runtime.GOOS == "windows" {
			os.Rename(archiveFilePath, path.Join(depFolder, "shaka-packager.exe"))
		}
	}

	// write version to file
	err2 := ioutil.WriteFile(versionFilePath, []byte(latestRelease.TagName), 0644)
	// check for error
	if err2 != nil {
		logger.Errorf("Error writing %s version file: %s\n", depName, err2)
		return false
	}

	return true
}

func HandleDarwinFFMPEGBinaryUpdate(latestVersion FFMPEGMacInfo, depFolder string, versionFilePath string) bool {
	downloadURL := latestVersion.Download.ZIP.URL
	archiveFilePath := path.Join(depFolder, path.Base(downloadURL))

	// update binary
	if !Exists(archiveFilePath) {
		logger.Debugf("Downloading FFMPEG to %s\n", archiveFilePath)

		// download binary
		err := DownloadFileNew(downloadURL, depFolder)
		if err != nil {
			logger.Errorf("Error downloading FFMPEG: %s\n", err)
			return false
		}
	} else {
		logger.Debugf("Found existing FFMPEG archive: %s\n", archiveFilePath)
	}

	logger.Debug("Unzipping FFMPEG")

	// unzip binary
	files, err := Unzip(archiveFilePath, depFolder)
	if err != nil {
		logger.Errorf("Error unzipping FFMPEG archive: %s\n", err)
		return false
	}

	logger.Debug("FFMPEG unzipped, Moving binaries")

	// move executable files to bin folder
	for _, v := range files {
		if strings.HasSuffix(v, "ffmpeg") {
			logger.Debugf("Moving executable file %s to %s\n", v, depFolder)
			err := os.Rename(v, path.Join(depFolder, path.Base(v)))
			if err != nil {
				logger.Errorf("Error moving executable file %s to %s: %s\n", v, depFolder, err)
				return false
			}
		}
	}

	basename := path.Base(archiveFilePath)
	tmpFolderPath := path.Join(depFolder, strings.TrimSuffix(basename, path.Ext(basename)))

	logger.Debugf("Removing temp folder: %s\n", tmpFolderPath)

	err1 := os.RemoveAll(tmpFolderPath)
	if err1 != nil {
		logger.Errorf("Error removing temp folder: %s\n", err1)
		return false
	}

	// write version to file
	err2 := ioutil.WriteFile(versionFilePath, []byte(latestVersion.Version), 0644)
	// check for error
	if err2 != nil {
		logger.Errorf("Error writing FFMPEG version file: %s\n", err2)
		return false
	}

	return true
}

func CheckFFMPEGWindows() bool {
	depFolder := path.Join("bin", "ffmpeg")

	err := MakeDirectoryIfNotExists(depFolder)
	if err != nil {
		logger.Fatalf("Error creating directory %s: %s", depFolder, err)
		return false
	}

	versionFilePath := path.Join(depFolder, ".version")
	latestRelease, err := GetLatestGithubRelease("https://api.github.com/repos/GyanD/codexffmpeg/releases/latest")

	// check for error
	if err != nil {
		logger.Errorf("Error getting latest ffmpeg release: %s\n", err)
		return false
	}

	// the zip file
	asset := latestRelease.Assets[1]

	// check if version file exists
	if Exists(versionFilePath) {
		// read version from file
		version, err := ioutil.ReadFile(versionFilePath)
		// check for error
		if err != nil {
			logger.Errorf("Error reading ffmpeg version file: %s\n", err)
			return false
		}

		// compare version to latest release version
		if strings.Compare(string(version), latestRelease.TagName) == -1 {
			logger.Warningf("FFMPEG is out of date, current version: " + string(version) + "; latest version: " + latestRelease.TagName + "\n")
			return HandleBinaryUpdate("ffmpeg", latestRelease, asset, depFolder, versionFilePath)
		} else {
			// ffmpeg is up to date
			logger.Notice("FFMPEG is up to date")
			return true
		}
	} else {
		// version file does not exist
		logger.Notice("FFMPEG version file does not exist")
		return HandleBinaryUpdate("ffmpeg", latestRelease, asset, depFolder, versionFilePath)
	}
}

func CheckFFMPEGDarwin() bool {
	depFolder := path.Join("bin", "ffmpeg")

	err := MakeDirectoryIfNotExists(depFolder)
	if err != nil {
		logger.Fatalf("Error creating directory %s: %s", depFolder, err)
		return false
	}

	versionFilePath := path.Join(depFolder, ".version")
	latestVersion, err := GetLatestMacFFMPEGVersion("https://evermeet.cx/ffmpeg/info/ffmpeg/snapshot")

	// check for error
	if err != nil {
		logger.Errorf("Error getting latest ffmpeg release: %s\n", err)
		return false
	}

	// check if version file exists
	if Exists(versionFilePath) {
		// read version from file
		version, err := ioutil.ReadFile(versionFilePath)
		// check for error
		if err != nil {
			logger.Errorf("Error reading ffmpeg version file: %s\n", err)
			return false
		}

		// compare version to latest release version
		if strings.Compare(string(version), latestVersion.Version) == -1 {
			logger.Warningf("FFMPEG is out of date, current version: " + string(version) + "; latest version: " + latestVersion.Version + "\n")
			return HandleDarwinFFMPEGBinaryUpdate(latestVersion, depFolder, versionFilePath)
		} else {
			// ffmpeg is up to date
			logger.Notice("FFMPEG is up to date")
			return true
		}
	} else {
		// version file does not exist
		logger.Notice("FFMPEG version file does not exist")
		return HandleDarwinFFMPEGBinaryUpdate(latestVersion, depFolder, versionFilePath)
	}
}

func CheckAria2() bool {
	depFolder := path.Join("bin", "aria2")

	err := MakeDirectoryIfNotExists(depFolder)
	if err != nil {
		logger.Fatalf("Error creating directory %s: %s", depFolder, err)
		return false
	}

	versionFilePath := path.Join(depFolder, ".version")
	var apiURL string
	if runtime.GOOS != "darwin" {
		apiURL = "https://api.github.com/repos/aria2/aria2/releases/latest"
	} else {
		// the latest release (1.36) doesn't contain any binaries for macOS
		apiURL = "https://api.github.com/repos/aria2/aria2/releases/20496544"
	}
	latestRelease, err := GetLatestGithubRelease(apiURL)

	// check for error
	if err != nil {
		logger.Errorf("Error getting latest aria2 release: %s\n", err)
		return false
	}

	// the zip file for windows 64 bit
	var asset Asset
	if runtime.GOOS == "windows" {
		if runtime.GOARCH != "amd64" {
			// the zip file for windows 32 bit
			asset = latestRelease.Assets[1]
		} else {
			// the zip file for windows 64 bit
			asset = latestRelease.Assets[2]
		}
	} else if runtime.GOOS == "darwin" {
		asset = latestRelease.Assets[2]
	}

	// check if version file exists
	if Exists(versionFilePath) {
		// read version from file
		currentVersion, err := ioutil.ReadFile(versionFilePath)
		// check for error
		if err != nil {
			logger.Errorf("Error reading aria2 version file: %s\n", err)
			return false
		}

		// compare version to latest release version
		if strings.Compare(string(currentVersion), latestRelease.TagName) == -1 {
			logger.Warningf("Aria2 is out of date, current version: " + string(currentVersion) + "; latest version: " + latestRelease.TagName + "\n")
			return HandleBinaryUpdate("aria2", latestRelease, asset, depFolder, versionFilePath)
		} else {
			// aria2 is up to date
			logger.Notice("Aria2 is up to date")
			return true
		}
	} else {
		// version file does not exist
		logger.Notice("Aria2 version file does not exist")
		return HandleBinaryUpdate("aria2", latestRelease, asset, depFolder, versionFilePath)
	}
}

func CheckYtdlp() bool {
	depFolder := path.Join("bin", "ytdlp")

	err := MakeDirectoryIfNotExists(depFolder)
	if err != nil {
		logger.Fatalf("Error creating directory %s: %s", depFolder, err)
		return false
	}

	versionFilePath := path.Join(depFolder, ".version")
	latestRelease, err := GetLatestGithubRelease("https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest")

	// check for error
	if err != nil {
		logger.Errorf("Error getting latest yt-dlp release: %s\n", err)
		return false
	}

	// the zip file for windows 64 bit
	var asset Asset
	if runtime.GOOS == "linux" {
		asset = latestRelease.Assets[2]
	} else if runtime.GOOS == "windows" {
		asset = latestRelease.Assets[3]
	} else if runtime.GOOS == "darwin" {
		asset = latestRelease.Assets[5]
	}

	// check if version file exists
	if Exists(versionFilePath) {
		// read version from file
		currentVersion, err := ioutil.ReadFile(versionFilePath)
		// check for error
		if err != nil {
			logger.Errorf("Error reading yt-dlp version file: %s\n", err)
			return false
		}

		// compare version to latest release version
		if strings.Compare(string(currentVersion), latestRelease.TagName) == -1 {
			logger.Warningf("yt-dlp is out of date, current version: " + string(currentVersion) + "; latest version: " + latestRelease.TagName + "\n")
			return HandleBinaryUpdate("yt-dlp", latestRelease, asset, depFolder, versionFilePath)
		} else {
			// ytdlp is up to date
			logger.Notice("yt-dlp is up to date")
			return true
		}
	} else {
		// version file does not exist
		logger.Notice("yt-dlp version file does not exist")
		return HandleBinaryUpdate("ytdlp", latestRelease, asset, depFolder, versionFilePath)
	}
}

func CheckShakaPackager() bool {
	depFolder := path.Join("bin", "shaka-packager")

	err := MakeDirectoryIfNotExists(depFolder)
	if err != nil {
		logger.Fatalf("Error creating directory %s: %s", depFolder, err)
		return false
	}

	versionFilePath := path.Join(depFolder, ".version")
	latestRelease, err := GetLatestGithubRelease("https://api.github.com/repos/google/shaka-packager/releases/latest")

	// check for error
	if err != nil {
		logger.Errorf("Error getting latest shaka-packager release: %s\n", err)
		return false
	}

	var asset Asset
	if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" {
		asset = latestRelease.Assets[4]
	} else if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		asset = latestRelease.Assets[5]
	} else if runtime.GOOS == "windows" {
		asset = latestRelease.Assets[7]
	} else if runtime.GOOS == "darwin" {
		asset = latestRelease.Assets[6]
	}

	// check if version file exists
	if Exists(versionFilePath) {
		// read version from file
		currentVersion, err := ioutil.ReadFile(versionFilePath)
		// check for error
		if err != nil {
			logger.Errorf("Error reading shaka-packager version file: %s\n", err)
			return false
		}

		// compare version to latest release version
		if strings.Compare(string(currentVersion), latestRelease.TagName) == -1 {
			logger.Warningf("shaka-packager is out of date, current version: " + string(currentVersion) + "; latest version: " + latestRelease.TagName + "\n")
			return HandleBinaryUpdate("shaka-packager", latestRelease, asset, depFolder, versionFilePath)
		} else {
			// shaka-packager is up to date
			logger.Notice("shaka-packager is up to date")
			return true
		}
	} else {
		// version file does not exist
		logger.Notice("shaka-packager version file does not exist")
		return HandleBinaryUpdate("shaka-packager", latestRelease, asset, depFolder, versionFilePath)
	}
}

func Updater() {
	var aria2CheckStatus bool
	var shakaPackagerCheckStatus bool
	var ffmpegCheckStatus bool
	var ytdlpCheckStatus bool

	// aria2
	if runtime.GOOS != "linux" {
		aria2CheckStatus = CheckAria2()
	} else {
		// This is handled in main.go since its installed externally, if we got here it means it was found
		aria2CheckStatus = true
	}

	// shaka packager
	shakaPackagerCheckStatus = CheckShakaPackager()

	// ffmpeg
	if runtime.GOOS == "windows" {
		ffmpegCheckStatus = CheckFFMPEGWindows()
	} else if runtime.GOOS == "darwin" {
		ffmpegCheckStatus = CheckFFMPEGDarwin()
	} else if runtime.GOOS == "linux" {
		// This is handled in main.go since its installed externally, if we got here it means it was found
		ffmpegCheckStatus = true
	}

	// ytdlp
	ytdlpCheckStatus = CheckYtdlp()

	if !aria2CheckStatus {
		logger.Fatal("Aria2 check failed")
		os.Exit(1)
	}

	if !shakaPackagerCheckStatus {
		logger.Fatal("Shaka Packager check failed")
		os.Exit(1)
	}

	if !ffmpegCheckStatus {
		logger.Fatal("FFMPEG check failed")
		os.Exit(1)
	}

	if !ytdlpCheckStatus {
		logger.Fatal("yt-dlp check failed")
		os.Exit(1)
	}
}
