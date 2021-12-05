package main

import (
	"flag"
	"os"
	"os/exec"
	"path"
	"runtime"

	"github.com/op/go-logging"
)

var version string = "DEVELOPMENT"
var logger = logging.MustGetLogger("udemy-dl")
var loggerBackend = logging.NewLogBackend(os.Stderr, "", 0)
var loggerFormat = logging.MustStringFormatter(
	`%{color}%{time:15:04:05} %{level:.4s} ▶ [%{shortfunc}] %{color:reset} %{message}`,
)
var backendFormatter = logging.NewBackendFormatter(loggerBackend, loggerFormat)

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
	flag.Parse()

	if *versionPtr {
		logger.Infof("Running version: %s", version)
		os.Exit(0)
	}

	ffmpegExecutablePath := path.Join("bin", "ffmpeg", "ffmpeg")
	aria2cExecutablePath := path.Join("bin", "aria2", "aria2c")
	// since ffmpeg gets installed on linux via package manager (or another external method), we dont need a full path to the executable
	if runtime.GOOS == "linux" {
		ffmpegExecutablePath = "ffmpeg"
	}

	mp4decryptExecutablePath := path.Join("bin", "bento4", "mp4decrypt")

	logger.Info("Checking system...")
	checkSystem()
	logger.Info("System check passed")

	logger.Info("Checking dependencies...")
	Updater()

	a, error := exec.Command(ffmpegExecutablePath, "-version").CombinedOutput()
	if error != nil {
		panic(error)
	}

	logger.Debug(string(a) + "\n")

	b, error1 := exec.Command(aria2cExecutablePath, "--version").CombinedOutput()
	if error1 != nil {
		panic(error)
	}

	logger.Debug(string(b) + "\n")

	c, error2 := exec.Command(mp4decryptExecutablePath).CombinedOutput()
	if error2 != nil && error2.Error() != "exit status 1" { // hack for testing, mp4decrypt doesnt have a version argument and prints to stderr by default
		panic(error)
	}

	logger.Debug(string(c) + "\n")
}
