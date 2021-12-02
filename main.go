package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("udemy-dl")
var loggerBackend = logging.NewLogBackend(os.Stderr, "", 0)
var loggerFormat = logging.MustStringFormatter(
	`%{color}%{time:15:04:05} %{level:.4s} ▶ [%{shortfunc}] %{color:reset} %{message}`,
)
var backendFormatter = logging.NewBackendFormatter(loggerBackend, loggerFormat)

var httpClient = &http.Client{Timeout: time.Second * 10}

func getReleaseJson(url string) (Release, error) {
	target := Release{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return target, err
	}

	res, getErr := httpClient.Do(req)
	if getErr != nil {
		return target, getErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return target, readErr
	}

	jsonErr := json.Unmarshal(body, &target)
	if jsonErr != nil {
		return target, jsonErr
	}

	return target, nil
}

func getTagJson(url string) ([]Tag, error) {
	target := []Tag{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return target, err
	}

	res, getErr := httpClient.Do(req)
	if getErr != nil {
		return target, getErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return target, readErr
	}

	jsonErr := json.Unmarshal(body, &target)
	if jsonErr != nil {
		return target, jsonErr
	}

	return target, nil
}

func getMacFFFMPEGInfo(url string) (FFMPEGMacInfo, error) {
	target := FFMPEGMacInfo{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return target, err
	}

	res, getErr := httpClient.Do(req)
	if getErr != nil {
		return target, getErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return target, readErr
	}

	jsonErr := json.Unmarshal(body, &target)
	if jsonErr != nil {
		return target, jsonErr
	}

	return target, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func PrintDownloadPercent(done chan int64, path string, total int64) {
	var stop bool = false
	file, err := os.Open(path)
	if err != nil {
		logger.Fatal(err)
	}
	defer file.Close()
	for {
		select {
		case <-done:
			stop = true
		default:
			fi, err := file.Stat()
			if err != nil {
				logger.Fatal(err)
			}

			size := fi.Size()
			if size == 0 {
				size = 1
			}

			var percent float64 = float64(size) / float64(total) * 100
			fmt.Printf("%.0f", percent)
			fmt.Println("%")
		}

		if stop {
			break
		}
		time.Sleep(time.Second)
	}
}

func DownloadFile(url string, dest string) error {

	file := path.Base(url)

	logger.Infof("Downloading file %s\n", file)

	var pathh bytes.Buffer
	pathh.WriteString(dest)
	pathh.WriteString("/")
	pathh.WriteString(file)

	start := time.Now()

	out, err := os.Create(pathh.String())

	if err != nil {
		logger.Debugf(pathh.String())
		return err
	}

	defer out.Close()

	headResp, err := http.Head(url)

	if err != nil {
		return err
	}

	defer headResp.Body.Close()

	size, err := strconv.Atoi(headResp.Header.Get("Content-Length"))

	if err != nil {
		return err
	}

	done := make(chan int64)

	go PrintDownloadPercent(done, pathh.String(), int64(size))

	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)

	if err != nil {
		return err
	}

	done <- n

	elapsed := time.Since(start)
	logger.Infof("Download completed in %s", elapsed)
	return nil
}

func makeDirectoryIfNotExists(fpath string) error {
	if !exists(fpath) {
		err := os.MkdirAll(fpath, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func getLatestGithubRelease(url string) (Release, error) {
	// get latest release
	data, err1 := getReleaseJson(url)

	// check for error
	if err1 != nil {
		return data, err1
	}
	return data, nil
}

func getLatestGithubTag(url string) ([]Tag, error) {
	// get latest release
	data, err1 := getTagJson(url)

	// check for error
	if err1 != nil {
		return data, err1
	}
	return data, nil
}

func getLatestMacFFMPEGVersion(url string) (FFMPEGMacInfo, error) {
	// get latest release
	data, err1 := getMacFFFMPEGInfo(url)

	// check for error
	if err1 != nil {
		return data, err1
	}
	return data, nil
}

/*
* Source: https://golangcode.com/unzip-files-in-go/
 */
func unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := path.Join(dest, f.Name)

		// Check for ZipSlip.
		if !strings.HasPrefix(fpath, path.Clean(dest)) {
			return filenames, fmt.Errorf("%s: illegal file path", dest)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(dest, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(path.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func handleBinaryUpdate(depName string, latestRelease Release, asset Asset, depFolder string, versionFilePath string) bool {
	archiveFilePath := path.Join(depFolder, asset.Name)

	// update binary
	if !exists(archiveFilePath) {
		logger.Debugf("Downloading %s to %s\n", depName, archiveFilePath)

		// download binary
		err := DownloadFile(asset.Browser_Download_URL, depFolder)
		if err != nil {
			logger.Errorf("Error downloading %s: %s\n", depName, err)
			return false
		}
	} else {
		logger.Debugf("Found existing %s archive: %s\n", depName, archiveFilePath)
	}

	logger.Debugf("Unzipping %s", depName)

	// unzip binary
	files, err := unzip(archiveFilePath, depFolder)
	if err != nil {
		logger.Errorf("Error unzipping %s archive: %s\n", depName, err)
		return false
	}

	logger.Infof("%s unzipped, Moving binaries", depName)

	// move executable files to bin folder
	for _, v := range files {
		if strings.HasSuffix(v, "aria2c.exe") || strings.HasSuffix(v, "ffmpeg.exe") {
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
	err2 := ioutil.WriteFile(versionFilePath, []byte(latestRelease.TagName), 0644)
	// check for error
	if err2 != nil {
		logger.Errorf("Error writing %s version file: %s\n", depName, err2)
		return false
	}

	return true
}

func handleMacFFMPEGBinaryUpdate(latestVersion FFMPEGMacInfo, depFolder string, versionFilePath string) bool {
	downloadURL := latestVersion.Download.ZIP.URL
	archiveFilePath := path.Join(depFolder, path.Base(downloadURL))

	// update binary
	if !exists(archiveFilePath) {
		logger.Debugf("Downloading FFMPEG to %s\n", archiveFilePath)

		// download binary
		err := DownloadFile(downloadURL, depFolder)
		if err != nil {
			logger.Errorf("Error downloading FFMPEG: %s\n", err)
			return false
		}
	} else {
		logger.Debugf("Found existing FFMPEG archive: %s\n", archiveFilePath)
	}

	logger.Debug("Unzipping FFMPEG")

	// unzip binary
	files, err := unzip(archiveFilePath, depFolder)
	if err != nil {
		logger.Errorf("Error unzipping FFMPEG archive: %s\n", err)
		return false
	}

	logger.Info("FFMPEG unzipped, Moving binaries")

	// move executable files to bin folder
	for _, v := range files {
		if strings.HasSuffix(v, ".exe") {
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

func handleBento4BinaryUpdate(depFolder string, versionFilePath string, version string, tagName string) bool {
	var osString string
	if runtime.GOOS == "windows" {
		osString = "x86_64-microsoft-win32"
	} else if runtime.GOOS == "linux" {
		osString = "x86_64-unknown-linux"
	} else if runtime.GOOS == "darwin" {
		osString = "universal-apple-macosx"
	}

	archiveFilename := fmt.Sprintf("Bento4-SDK-%s.%s.zip", version, osString)
	archiveFilePath := path.Join(depFolder, archiveFilename)

	downloadURL := fmt.Sprintf("https://www.bok.net/Bento4/binaries/%s", archiveFilename)

	// update binary
	if !exists(archiveFilePath) {
		logger.Debugf("Downloading bento4 to %s\n", archiveFilePath)

		// download binary
		err := DownloadFile(downloadURL, depFolder)
		if err != nil {
			logger.Errorf("Error downloading bento4: %s\n", err)
			return false
		}
	} else {
		logger.Debugf("Found existing bento4 archive: %s\n", archiveFilePath)
	}

	logger.Info("Unzipping bento4")

	// unzip binary
	files, err := unzip(archiveFilePath, depFolder)
	if err != nil {
		logger.Errorf("Error unzipping bento4 archive: %s\n", err)
		return false
	}

	logger.Info("Bento4 unzipped, Moving binaries")

	// move executable files to bin folder
	for _, v := range files {
		if strings.HasSuffix(v, "mp4decrypt.exe") {
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
	err2 := ioutil.WriteFile(versionFilePath, []byte(tagName), 0644)
	// check for error
	if err2 != nil {
		logger.Errorf("Error writing bento4 version file: %s\n", err2)
		return false
	}

	return true
}

func checkFFMPEGWindows() bool {
	depFolder := path.Join("bin", "ffmpeg")

	err := makeDirectoryIfNotExists(depFolder)
	if err != nil {
		logger.Fatalf("Error creating directory %s: %s", depFolder, err)
		return false
	}

	versionFilePath := path.Join(depFolder, "ffmpeg.version")
	latestRelease, err := getLatestGithubRelease("https://api.github.com/repos/GyanD/codexffmpeg/releases/latest")

	// check for error
	if err != nil {
		logger.Errorf("Error getting latest ffmpeg release: %s\n", err)
		return false
	}

	// the zip file
	asset := latestRelease.Assets[1]

	// check if version file exists
	if exists(versionFilePath) {
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
			return handleBinaryUpdate("ffmpeg", latestRelease, asset, depFolder, versionFilePath)
		} else {
			// ffmpeg is up to date
			logger.Info("FFMPEG is up to date")
			return true
		}
	} else {
		// version file does not exist
		logger.Notice("FFMPEG version file does not exist")
		return handleBinaryUpdate("ffmpeg", latestRelease, asset, depFolder, versionFilePath)
	}
}

func checkFFMPEGMac() bool {
	depFolder := path.Join("bin", "ffmpeg")

	err := makeDirectoryIfNotExists(depFolder)
	if err != nil {
		logger.Fatalf("Error creating directory %s: %s", depFolder, err)
		return false
	}

	versionFilePath := path.Join(depFolder, "ffmpeg.version")
	latestVersion, err := getLatestMacFFMPEGVersion("https://evermeet.cx/ffmpeg/info/ffmpeg/snapshot")

	// check for error
	if err != nil {
		logger.Errorf("Error getting latest ffmpeg release: %s\n", err)
		return false
	}

	// check if version file exists
	if exists(versionFilePath) {
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
			return handleMacFFMPEGBinaryUpdate(latestVersion, depFolder, versionFilePath)
		} else {
			// ffmpeg is up to date
			logger.Info("FFMPEG is up to date")
			return true
		}
	} else {
		// version file does not exist
		logger.Notice("FFMPEG version file does not exist")
		return handleMacFFMPEGBinaryUpdate(latestVersion, depFolder, versionFilePath)
	}
}

func checkAria2() bool {
	depFolder := path.Join("bin", "aria2")

	err := makeDirectoryIfNotExists(depFolder)
	if err != nil {
		logger.Fatalf("Error creating directory %s: %s", depFolder, err)
		return false
	}

	versionFilePath := path.Join(depFolder, "aria2.version")
	latestRelease, err := getLatestGithubRelease("https://api.github.com/repos/aria2/aria2/releases/latest")

	// check for error
	if err != nil {
		logger.Errorf("Error getting latest aria2 release: %s\n", err)
		return false
	}

	// the zip file for windows 64 bit
	asset := latestRelease.Assets[2]

	// check if version file exists
	if exists(versionFilePath) {
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
			return handleBinaryUpdate("aria2", latestRelease, asset, depFolder, versionFilePath)
		} else {
			// aria2 is up to date
			logger.Info("Aria2 is up to date")
			return true
		}
	} else {
		// version file does not exist
		logger.Notice("Aria2 version file does not exist")
		return handleBinaryUpdate("aria2", latestRelease, asset, depFolder, versionFilePath)
	}
}

func checkBento4() bool {
	depFolder := path.Join("bin", "bento4")

	err := makeDirectoryIfNotExists(depFolder)
	if err != nil {
		logger.Fatalf("Error creating directory %s: %s", depFolder, err)
		return false
	}

	versionFilePath := path.Join(depFolder, "bento4.version")
	tags, err := getLatestGithubTag("https://api.github.com/repos/axiomatic-systems/Bento4/tags")

	latestTag := tags[0]

	// check for error
	if err != nil {
		logger.Errorf("Error getting latest bento4 tag: %s\n", err)
		return false
	}

	versionString1 := strings.Split(latestTag.Name, "v")[1]
	versionString2 := strings.ReplaceAll(versionString1, ".", "-")

	// check if version file exists
	if exists(versionFilePath) {
		// read version from file
		currentVersion, err := ioutil.ReadFile(versionFilePath)
		// check for error
		if err != nil {
			logger.Errorf("Error reading bento4 version file: %s\n", err)
			return false
		}

		// compare version to latest release version
		if strings.Compare(string(currentVersion), latestTag.Name) == -1 {
			logger.Warningf("Bento4 is out of date, current version: " + string(currentVersion) + "; latest version: " + latestTag.Name + "\n")
			return handleBento4BinaryUpdate(depFolder, versionFilePath, versionString2, latestTag.Name)
		} else {
			// Bento4 is up to date
			logger.Info("Bento4 is up to date")
			return true
		}
	} else {
		// version file does not exist
		logger.Notice("Bento4 version file does not exist")
		return handleBento4BinaryUpdate(depFolder, versionFilePath, versionString2, latestTag.Name)
	}
}

func checkSystem() {
	if runtime.GOOS == "windows" {
		if runtime.GOARCH != "amd64" && runtime.GOARCH != "386" {
			logger.Fatalf("Unsupported windows architecture: %s", runtime.GOARCH)
			os.Exit(1)
		}
	} else if runtime.GOOS == "linux" {
		if runtime.GOARCH != "amd64" && runtime.GOARCH != "386" && runtime.GOARCH != "arm" && runtime.GOARCH != "arm64" {
			logger.Fatalf("Unsupported linux architecture: %s", runtime.GOARCH)
			os.Exit(1)
		}
	} else if runtime.GOOS == "darwin" {
		if runtime.GOARCH != "amd64" {
			logger.Fatalf("Unsupported darwin architecture: %s", runtime.GOARCH)
			os.Exit(1)
		}
	} else {
		logger.Fatalf("Unsupported operating system: %s", runtime.GOOS)
		os.Exit(1)
	}
}

func main() {
	logging.SetBackend(backendFormatter)

	logger.Info("Checking system...")
	checkSystem()
	logger.Info("System check passed")

	logger.Info("Checking dependencies...")

	var ffmpegCheckStatus bool

	if runtime.GOOS == "windows" {
		ffmpegCheckStatus = checkFFMPEGWindows()
	} else if runtime.GOOS == "linux" {
		_, error := exec.LookPath("ffmpeg")
		if error != nil {
			logger.Fatal("Please install FFMPEG using your system package manager: https://ffmpeg.org/download.html#build-linux")
			os.Exit(1)
		}
		ffmpegCheckStatus = true
	} else if runtime.GOOS == "darwin" {
		ffmpegCheckStatus = checkFFMPEGMac()
	}

	aria2CheckStatus := checkAria2()
	bento4CheckStatus := checkBento4()

	if !ffmpegCheckStatus {
		logger.Fatal("FFMPEG check failed")
		os.Exit(1)
	}

	if !aria2CheckStatus {
		logger.Fatal("Aria2 check failed")
		os.Exit(1)
	}

	if !bento4CheckStatus {
		logger.Fatal("Bento4 check failed")
		os.Exit(1)
	}
}
