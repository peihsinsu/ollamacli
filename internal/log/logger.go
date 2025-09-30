package log

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	SetLevel(levelStr string, verbose bool)
}

type logger struct {
	level   Level
	verbose bool
	std     *log.Logger
	err     *log.Logger
}

func New(levelStr string, verbose bool) Logger {
	l := &logger{
		verbose: verbose,
		std:     log.New(os.Stdout, "", 0),
		err:     log.New(os.Stderr, "", 0),
	}

	l.SetLevel(levelStr, verbose)
	return l
}

func (l *logger) SetLevel(levelStr string, verbose bool) {
	l.verbose = verbose

	switch strings.ToLower(levelStr) {
	case "debug":
		l.level = LevelDebug
	case "info":
		l.level = LevelInfo
	case "warn", "warning":
		l.level = LevelWarn
	case "error":
		l.level = LevelError
	default:
		l.level = LevelInfo
	}

	if verbose {
		l.level = LevelDebug
	}
}

func (l *logger) Debug(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.log(LevelDebug, format, args...)
	}
}

func (l *logger) Info(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.log(LevelInfo, format, args...)
	}
}

func (l *logger) Warn(format string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.log(LevelWarn, format, args...)
	}
}

func (l *logger) Error(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.logError(LevelError, format, args...)
	}
}

func (l *logger) log(level Level, format string, args ...interface{}) {
	if l.verbose {
		msg := fmt.Sprintf("[%s] %s", levelNames[level], fmt.Sprintf(format, args...))
		l.std.Println(msg)
	} else if level >= LevelInfo {
		l.std.Printf(format+"\n", args...)
	}
}

func (l *logger) logError(level Level, format string, args ...interface{}) {
	if l.verbose {
		msg := fmt.Sprintf("[%s] %s", levelNames[level], fmt.Sprintf(format, args...))
		l.err.Println(msg)
	} else {
		l.err.Printf(format+"\n", args...)
	}
}