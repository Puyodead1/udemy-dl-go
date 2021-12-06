package main

import (
	"flag"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

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
var mp4decryptExecutablePath = path.Join("bin", "bento4", "mp4decrypt")
var ytdlpExecutablePath = path.Join("bin", "ytdlp", "yt-dlp")

var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")

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

	// since ffmpeg gets installed on linux via package manager (or another external method), we dont need a full path to the executable
	if runtime.GOOS == "linux" {
		ffmpegExecutablePath = "ffmpeg"
	}

	logger.Info("Checking system...")
	checkSystem()
	logger.Info("System check passed")

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

	_, mp4decryptError := exec.Command(mp4decryptExecutablePath).CombinedOutput()
	// hack to check if mp4decrypt is installed, mp4decrypt prints to stderr by default
	if mp4decryptError != nil && !strings.Contains(mp4decryptError.Error(), "exit status") {
		panic(mp4decryptError)
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

	// TODO: get course information
}
