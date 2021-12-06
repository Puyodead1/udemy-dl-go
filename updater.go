package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/bzip2"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var UpdaterClient = &http.Client{Timeout: time.Second * 10}

func GetReleaseJson(url string) (Release, error) {
	target := Release{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return target, err

	}

	if GITHUB_TOKEN != "" {
		req.Header.Add("Authentication", fmt.Sprintf("token %s", GITHUB_TOKEN))
	}

	res, getErr := UpdaterClient.Do(req)
	if getErr != nil {
		return target, getErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return target, fmt.Errorf("%s: %s", res.Status, body)
	}

	if readErr != nil {
		return target, readErr
	}

	jsonErr := json.Unmarshal(body, &target)
	if jsonErr != nil {
		return target, jsonErr
	}

	return target, nil
}

func GetTagJson(url string) ([]Tag, error) {
	target := []Tag{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return target, err
	}

	if GITHUB_TOKEN != "" {
		req.Header.Add("Authentication", fmt.Sprintf("token %s", GITHUB_TOKEN))
	}

	res, getErr := UpdaterClient.Do(req)
	if getErr != nil {
		return target, getErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return target, fmt.Errorf("%s: %s", res.Status, body)
	}

	if readErr != nil {
		return target, readErr
	}

	jsonErr := json.Unmarshal(body, &target)
	if jsonErr != nil {
		return target, jsonErr
	}

	return target, nil
}

func GetMacFFFMPEGInfo(url string) (FFMPEGMacInfo, error) {
	target := FFMPEGMacInfo{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return target, err
	}

	res, getErr := UpdaterClient.Do(req)
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

func MakeDirectoryIfNotExists(fpath string) error {
	if !Exists(fpath) {
		err := os.MkdirAll(fpath, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetLatestGithubRelease(url string) (Release, error) {
	// get latest release
	data, err1 := GetReleaseJson(url)

	// check for error
	if err1 != nil {
		return data, err1
	}
	return data, nil
}

func GetLatestGithubTag(url string) ([]Tag, error) {
	// get latest release
	data, err1 := GetTagJson(url)

	// check for error
	if err1 != nil {
		return data, err1
	}
	return data, nil
}

func GetLatestMacFFMPEGVersion(url string) (FFMPEGMacInfo, error) {
	// get latest release
	data, err1 := GetMacFFFMPEGInfo(url)

	// check for error
	if err1 != nil {
		return data, err1
	}
	return data, nil
}

func DecompressBZIP(src string, dest string) ([]string, error) {

	var filenames []string

	f, err := os.OpenFile(src, 0, 0)
	if err != nil {
		return filenames, err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	cr := bzip2.NewReader(br)
	tarReader := tar.NewReader(cr)

	for {
		f, err := tarReader.Next()

		if err == io.EOF {
			break
		}

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

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, tarReader)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()

		if err != nil {
			return filenames, err
		}
	}

	return filenames, nil
}

/*
* Source: https://golangcode.com/Unzip-files-in-go/
 */
func Unzip(src string, dest string) ([]string, error) {

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

func HandleBinaryUpdate(depName string, latestRelease Release, asset Asset, depFolder string, versionFilePath string) bool {
	archiveFilePath := path.Join(depFolder, asset.Name)

	// update binary
	if !Exists(archiveFilePath) {
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

	if depName != "ytdlp" {
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

func HandleBento4BinaryUpdate(depFolder string, versionFilePath string, version string, tagName string) bool {
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
	if !Exists(archiveFilePath) {
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

	logger.Debug("Unzipping bento4")

	// unzip binary
	files, err := Unzip(archiveFilePath, depFolder)
	if err != nil {
		logger.Errorf("Error unzipping bento4 archive: %s\n", err)
		return false
	}

	logger.Debug("Bento4 unzipped, Moving binaries")

	// move executable files to bin folder
	for _, v := range files {
		if strings.HasSuffix(v, "mp4decrypt") || strings.HasSuffix(v, "mp4decrypt.exe") {
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

func CheckBento4() bool {
	depFolder := path.Join("bin", "bento4")

	err := MakeDirectoryIfNotExists(depFolder)
	if err != nil {
		logger.Fatalf("Error creating directory %s: %s", depFolder, err)
		return false
	}

	versionFilePath := path.Join(depFolder, ".version")
	tags, err := GetLatestGithubTag("https://api.github.com/repos/axiomatic-systems/Bento4/tags")

	// check for error
	if err != nil {
		logger.Errorf("Error getting latest bento4 tag: %s\n", err)
		return false
	}

	latestTag := tags[0]

	versionString1 := strings.Split(latestTag.Name, "v")[1]
	versionString2 := strings.ReplaceAll(versionString1, ".", "-")

	// check if version file exists
	if Exists(versionFilePath) {
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
			return HandleBento4BinaryUpdate(depFolder, versionFilePath, versionString2, latestTag.Name)
		} else {
			// Bento4 is up to date
			logger.Notice("Bento4 is up to date")
			return true
		}
	} else {
		// version file does not exist
		logger.Notice("Bento4 version file does not exist")
		return HandleBento4BinaryUpdate(depFolder, versionFilePath, versionString2, latestTag.Name)
	}
}

func Updater() {
	var aria2CheckStatus bool
	var bento4CheckStatus bool
	var ffmpegCheckStatus bool
	var ytdlpCheckStatus bool

	// aria2
	if runtime.GOOS != "linux" {
		aria2CheckStatus = CheckAria2()
	} else {
		// This is handled in main.go since its installed externally, if we got here it means it was found
		aria2CheckStatus = true
	}

	// bento4
	bento4CheckStatus = CheckBento4()

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

	if !bento4CheckStatus {
		logger.Fatal("Bento4 check failed")
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
