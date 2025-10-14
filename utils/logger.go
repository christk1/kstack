package utils

import (
	"log"
	"os"
)

// ANSI color codes (kept simple; disabled when colorEnabled is false)
const (
	colorReset  = "\033[0m"
	colorBlue   = "\033[34m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorGray   = "\033[90m"
)

// Logger provides a thin leveled logging wrapper over the standard library logger.
type Logger struct {
	verbose      bool
	colorEnabled bool
	logger       *log.Logger
}

// NewLogger returns a configured logger.
func NewLogger(verbose bool) *Logger {
	return &Logger{verbose: verbose, colorEnabled: true, logger: log.New(os.Stdout, "", log.LstdFlags)}
}

func (l *Logger) prefix(level string) string {
	if !l.colorEnabled {
		return level + ": "
	}
	switch level {
	case "INFO":
		return colorBlue + "INFO:" + colorReset + " "
	case "WARN":
		return colorYellow + "WARN:" + colorReset + " "
	case "ERROR":
		return colorRed + "ERROR:" + colorReset + " "
	case "DEBUG":
		return colorGray + "DEBUG:" + colorReset + " "
	default:
		return level + ": "
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.logger.Printf(l.prefix("INFO")+format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.logger.Printf(l.prefix("WARN")+format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.logger.Printf(l.prefix("ERROR")+format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if !l.verbose {
		return
	}
	l.logger.Printf(l.prefix("DEBUG")+format, v...)
}

// Package-level default logger convenience
var std = NewLogger(false)

// SetVerbose toggles verbosity on the package-level logger.
func SetVerbose(verbose bool) { std.verbose = verbose }

// SetColorEnabled toggles ANSI color output on the package-level logger.
func SetColorEnabled(enabled bool) { std.colorEnabled = enabled }

func Info(format string, v ...interface{})  { std.Info(format, v...) }
func Warn(format string, v ...interface{})  { std.Warn(format, v...) }
func Error(format string, v ...interface{}) { std.Error(format, v...) }
func Debug(format string, v ...interface{}) { std.Debug(format, v...) }
