package main

import (
	"flag"
	"os"
)

var version string = "DEVELOPMENT"
var debug bool = false

func main() {
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

	if version == "DEVELOPMENT" {
		debug = true
	}

	versionPtr := flag.Bool("version", false, "Print the program version")
	// skipUpdatePtr := flag.Bool("skip-update", false, "Skip update check")
	bearerPtr := flag.String("bearer", "", "Bearer token for authentication")
	courseUrlPtr := flag.String("course", "", "Course URL")
	debugPtr := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if *debugPtr {
		debug = true
	}

	if *versionPtr {
		Infof("Running version: %s", version)
		os.Exit(0)
	}

	if *bearerPtr == "" {
		Critical("A bearer token is required!")
	}

	if *courseUrlPtr == "" {
		Critical("A Course URL is required!")
	}

	ffmpegStatus, aria2Status, ytdlpStatus, shakaStatus, err := RunDependencyCheck()

	if err != nil {
		Criticalf("Dependency Check Error: %s", err)
	}

	// Print the status of all the checks
	if ffmpegStatus {
		Logf(SUCCESS, "FFMPEG: %t", ffmpegStatus)
	} else {
		Logf(ERROR, "FFMPEG: %t", ffmpegStatus)
	}

	if aria2Status {
		Logf(SUCCESS, "ARIA2: %t", aria2Status)
	} else {
		Logf(ERROR, "ARIA2: %t", aria2Status)
	}

	if ytdlpStatus {
		Logf(SUCCESS, "YTDLP: %t", ytdlpStatus)
	} else {
		Logf(ERROR, "YTDLP: %t", ytdlpStatus)
	}

	if shakaStatus {
		Logf(SUCCESS, "SHAKA: %t", shakaStatus)
	} else {
		Logf(ERROR, "SHAKA: %t", shakaStatus)
	}

	if !ffmpegStatus || !aria2Status || !ytdlpStatus || !shakaStatus {
		Critical("One or more dependencies are missing!")
	}
}
