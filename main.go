package main

import (
	"flag"
	"os"

	"github.com/op/go-logging"
)

var version string = "DEVELOPMENT"
var logger = logging.MustGetLogger("udemy-dl")
var loggerBackend = logging.NewLogBackend(os.Stderr, "", 0)
var loggerFormat = logging.MustStringFormatter(
	`%{color}%{time:15:04:05} %{level:.4s} ▶ [%{shortfunc}] %{color:reset} %{message}`,
)
var backendFormatter = logging.NewBackendFormatter(loggerBackend, loggerFormat)

func main() {
	logging.SetBackend(backendFormatter)

	versionPtr := flag.Bool("version", false, "Print the program version")
	// skipUpdatePtr := flag.Bool("skip-update", false, "Skip update check")
	bearerPtr := flag.String("bearer", "", "Bearer token for authentication")
	courseUrlPtr := flag.String("course", "", "Course URL")
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

	ffmpegStatus, aria2Status, ytdlpStatus, shakaStatus, err := RunDependencyCheck()

	if err != nil {
		logger.Errorf("Dependency Check Error: %s", err)
		os.Exit(1)
	}

	logger.Info("FFMEG: ", ffmpegStatus)
	logger.Info("ARIA2: ", aria2Status)
	logger.Info("YTDLP: ", ytdlpStatus)
	logger.Info("SHAKA: ", shakaStatus)

	// TODO: skip dependency check option
	// TODO: get course information
	// TODO: load from file argument
	// TODO: save to file argument
	// TODO: loading course information from file
	// TODO: save course information to file
	// TODO: get course content
	// TODO: process course content (this should be 'on the fly', so instead of pre-processing, just start downloading and fetch information for the lectures as we go)
	// TODO: info argument
	// TODO: mkv support
}
