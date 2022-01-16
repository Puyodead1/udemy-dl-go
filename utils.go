package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/bzip2"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

var CourseURLRegex = regexp.MustCompile(`(?i)(?://(?P<portal_name>.+?).udemy.com/(?:course(/draft)*/)?(?P<name_or_id>[a-zA-Z0-9_-]+))`)

// checks if a file at path exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func MoveFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		in.Close()
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	in.Close()
	if err != nil {
		return err
	}

	err = out.Sync()
	if err != nil {
		return err
	}

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return err
	}

	err = os.Remove(src)
	if err != nil {
		return err
	}
	return nil
}

// returns portal and course name / id
func ExtractCourseNameAndPortal(url string) (*string, *string) {
	match := CourseURLRegex.FindStringSubmatch(url)
	if match != nil {
		portal := match[1]
		course := match[3]
		return &portal, &course
	}
	return nil, nil
}

func LocateBinary(name string) (string, error) {
	fpath, error := exec.LookPath(name)
	if error != nil {
		return "", error
	}

	return fpath, nil
}

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

func DownloadFileNew(url string, dest string) error {
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	headRes, err := UpdaterClient.Head(url)
	if err != nil {
		return err
	}
	defer headRes.Body.Close()

	if err != nil {
		return err
	}

	// done := make(chan int64)

	bar := progressbar.DefaultBytes(
		headRes.ContentLength,
		"Downloading",
	)

	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	io.Copy(io.MultiWriter(f, bar), resp.Body)

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

type transport struct {
	underlyingTransport http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Origin", "www.udemy.com")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	return t.underlyingTransport.RoundTrip(req)
}

func checkSystem() {
	logger.Debugf("Detected %s architecture running %s\n", runtime.GOARCH, runtime.GOOS)
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
