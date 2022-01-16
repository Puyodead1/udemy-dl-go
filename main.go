package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/op/go-logging"
)

var version string = "DEVELOPMENT"
var logger = logging.MustGetLogger("udemy-dl")
var loggerBackend = logging.NewLogBackend(os.Stderr, "", 0)
var loggerFormat = logging.MustStringFormatter(
	`%{color}%{time:15:04:05} %{level:.4s} ▶ [%{shortfunc}] %{color:reset} %{message}`,
)
var backendFormatter = logging.NewBackendFormatter(loggerBackend, loggerFormat)

var ffmpegExecutablePath = path.Join("bin", "ffmpeg", "ffmpeg")
var aria2cExecutablePath = path.Join("bin", "aria2", "aria2c")

var shakaPackageExecutablePath = path.Join("bin", "shaka-packager", "shaka-packager")
var ytdlpExecutablePath = path.Join("bin", "ytdlp", "yt-dlp")

var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")

var UdemyHTTPClient = &http.Client{Timeout: time.Second * 10, Transport: &transport{underlyingTransport: http.DefaultTransport}}

func login() {
	res, err := UdemyHTTPClient.Get(LOGIN_URL)
	if err != nil {
		logger.Fatalf("Unable to get login page: %s", err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		logger.Debug(string(data))
		logger.Fatalf("Login page returned status code %d", res.StatusCode)
	}

	// load the html
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		logger.Fatalf("Unable to parse login page: %s", err)
	}

	// find csrf token
	doc.Find("input[name='csrfmiddlewaretoken']").Each(func(i int, s *goquery.Selection) {
		value := s.Text()
		logger.Debug("Found csrf token: %s", value)
	})
}

func main() {
	logging.SetBackend(backendFormatter)

	versionPtr := flag.Bool("version", false, "Print version")
	bearerPtr := flag.String("bearer", "", "Bearer token for authentication")
	courseUrlPtr := flag.String("course", "", "Course URL")
	githubTokenPtr := flag.String("github-token", "", "Github token for authentication")
	skipUpdatePtr := flag.Bool("skip-update", false, "Skip update check")
	flag.Parse()

	if *versionPtr {
		logger.Infof("Running version: %s", version)
		os.Exit(0)
	}

	if *bearerPtr == "" {
		logger.Fatalf("A bearer token is required!")
		os.Exit(1)
	}

	if *courseUrlPtr == "" {
		logger.Fatalf("A Course URL is required!")
		os.Exit(1)
	}

	if *githubTokenPtr != "" {
		GITHUB_TOKEN = *githubTokenPtr
	}

	if GITHUB_TOKEN != "" {
		logger.Notice("GitHub token found in environment or command line, using GitHub authentication")
	}

	// ffmpeg and aria2c are installed externally, so we need to check if they are installed
	if runtime.GOOS == "linux" {
		aria2cPath, e1 := LocateBinary("aria2c")
		ffmpegPath, e2 := LocateBinary("ffmpeg")

		if e1 != nil {
			logger.Fatal("Please install aria2 using your system package manager: https://aria2.github.io/")
			os.Exit(1)
		}
		aria2cExecutablePath = aria2cPath

		if e2 != nil {
			logger.Fatal("Please install FFMPEG using your system package manager: https://ffmpeg.org/download.html#build-linux")
			os.Exit(1)
		}
		ffmpegExecutablePath = ffmpegPath
	}

	logger.Info("Checking system...")
	checkSystem()
	logger.Notice("System check passed")

	if !*skipUpdatePtr {
		logger.Info("Checking dependencies...")
		Updater()
	} else {
		logger.Notice("Skipping dependency check")
	}

	_, aria2Error := exec.Command(aria2cExecutablePath, "--version").CombinedOutput()
	if aria2Error != nil {
		panic(aria2Error)
	}

	_, shakaError := exec.Command(shakaPackageExecutablePath, "-version").CombinedOutput()
	if shakaError != nil {
		panic(shakaError)
	}

	_, ffmpegError := exec.Command(ffmpegExecutablePath, "-version").CombinedOutput()
	if ffmpegError != nil {
		panic(ffmpegError)
	}

	_, ytdlpError := exec.Command(ytdlpExecutablePath, "-v").CombinedOutput()

	// hack to check if yt-dlp is installed, yt-dlp prints to stderr by default
	if ytdlpError != nil && !strings.Contains(ytdlpError.Error(), "exit status") {
		panic(ytdlpError)
	}

	portal, courseName := ExtractCourseNameAndPortal(*courseUrlPtr)
	if portal == nil || courseName == nil {
		logger.Fatalf("Invalid course URL: %s", *courseUrlPtr)
		os.Exit(1)
	}

	logger.Infof("Course: %s", *courseName)

	// TODO: login
	// login()
	// TODO: get course information
}
