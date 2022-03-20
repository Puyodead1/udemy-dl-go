package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

type ColorFunction func(format string, a ...interface{}) string

type LogLevel struct {
	LevelName string
	Function  ColorFunction
}

const (
	SUCCESS  = 0
	INFO     = 1
	ERROR    = 2
	DEBUG    = 3
	WARNING  = 4
	NOTICE   = 5
	CRITICAL = 6
)

func GetLogLevel(level int) LogLevel {
	switch level {
	case SUCCESS:
		return LogLevel{LevelName: "SUCCESS", Function: color.HiGreenString}
	case INFO:
		return LogLevel{LevelName: "INFO", Function: color.HiWhiteString}
	case ERROR:
		return LogLevel{LevelName: "ERROR", Function: color.HiRedString}
	case DEBUG:
		return LogLevel{LevelName: "DEBUG", Function: color.HiBlueString}
	case WARNING:
		return LogLevel{LevelName: "WARNING", Function: color.HiYellowString}
	case NOTICE:
		return LogLevel{LevelName: "NOTICE", Function: color.HiCyanString}
	case CRITICAL:
		return LogLevel{LevelName: "CRITICAL", Function: color.RedString}
	default:
		return LogLevel{LevelName: "INFO", Function: color.HiWhiteString}
	}
}

func Success(message string) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "SUCCESS", color.HiGreenString(message)))
}

func Successf(format string, args ...interface{}) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "SUCCESS", color.HiGreenString(fmt.Sprintf(format, args...))))
}

func Info(message string) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "INFO", color.HiWhiteString(message)))
}

func Infof(format string, args ...interface{}) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "INFO", color.HiWhiteString(fmt.Sprintf(format, args...))))
}

func Error(message string) {
	fmt.Println(fmt.Errorf("%s %s ▶ %s", time.Now().Format("03:04:05"), "ERROR", color.HiRedString(message)))
}

func Errorf(format string, args ...interface{}) {
	fmt.Println(fmt.Errorf("%s %s ▶ %s", time.Now().Format("03:04:05"), "ERROR", color.HiRedString(fmt.Sprintf(format, args...))))
}

func Debug(message string) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "DEBUG", color.HiBlueString(message)))
}

func Debugf(format string, args ...interface{}) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "DEBUG", color.HiBlueString(fmt.Sprintf(format, args...))))
}

func Warning(message string) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "WARNING", color.HiYellowString(message)))
}

func Warningf(format string, args ...interface{}) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "WARNING", color.HiYellowString(fmt.Sprintf(format, args...))))
}

func Notice(message string) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "NOTICE", color.HiCyanString(message)))
}

func Noticef(format string, args ...interface{}) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "NOTICE", color.HiCyanString(fmt.Sprintf(format, args...))))
}

func Critical(message string) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "CRITICAL", color.RedString(message)))
	os.Exit(1)
}

func Criticalf(format string, args ...interface{}) {
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), "CRITICAL", color.RedString(fmt.Sprintf(format, args...))))
	os.Exit(1)
}

func Log(level int, message string) {
	loglevel := GetLogLevel(level)
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), loglevel.LevelName, loglevel.Function(message)))
}

func Logf(level int, format string, args ...interface{}) {
	loglevel := GetLogLevel(level)
	fmt.Println(fmt.Sprintf("%s %s ▶ %s", time.Now().Format("03:04:05"), loglevel.LevelName, loglevel.Function(fmt.Sprintf(format, args...))))
}
