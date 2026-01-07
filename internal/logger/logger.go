package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	logFile     *os.File
	verbose     bool
)

const LogFileName = "cliprouter/cliprouter.log"

// InitLogger initializes the logging system
func InitLogger(isVerbose bool) error {
	verbose = isVerbose

	// Get log file path
	logPath := filepath.Join(xdg.DataHome, LogFileName)

	// Create directory if it doesn't exist
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file in append mode
	var err error
	logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Setup loggers
	var writers io.Writer
	if verbose {
		// Write to both file and stderr when verbose
		writers = io.MultiWriter(logFile, os.Stderr)
	} else {
		// Write only to file
		writers = logFile
	}

	infoLogger = log.New(writers, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(writers, "ERROR: ", log.Ldate|log.Ltime)
	debugLogger = log.New(writers, "DEBUG: ", log.Ldate|log.Ltime)

	return nil
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// LogInfo logs an informational message
func LogInfo(format string, v ...interface{}) {
	if infoLogger != nil {
		infoLogger.Printf(format, v...)
	}
}

// LogError logs an error message
func LogError(format string, v ...interface{}) {
	if errorLogger != nil {
		errorLogger.Printf(format, v...)
	}
}

// LogDebug logs a debug message
func LogDebug(format string, v ...interface{}) {
	if debugLogger != nil {
		debugLogger.Printf(format, v...)
	}
}
