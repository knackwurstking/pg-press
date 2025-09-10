// Package logger provides structured logging functionality for the pgpress application.
// It supports different log levels, colored output, and component-specific loggers.
package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// AppLogger is the global logger instance for the application
var AppLogger *Logger

var (
	logOutput        io.Writer = os.Stderr
	componentLoggers           = make(map[string]*Logger)
	componentMutex   sync.RWMutex
)

// Initialize sets up the application logger with appropriate defaults
func Initialize() {
	// Create logger with sensible defaults
	AppLogger = New(logOutput, WARN, true)

	// Configure based on environment
	configureFromEnvironment()

	// Replace standard log package
	InitializeFromStandardLog()
}

// SetOutput sets the output for all loggers.
func SetOutput(output io.Writer) {
	logOutput = output
	AppLogger.SetOutput(output)

	componentMutex.Lock()
	defer componentMutex.Unlock()
	for _, l := range componentLoggers {
		l.SetOutput(output)
	}
}

// NewComponentLogger creates a new logger for a specific component
func NewComponentLogger(component string) *Logger {
	logger := New(logOutput, AppLogger.level, AppLogger.colorPreference)
	logger.SetPrefix(fmt.Sprintf("[%s]", component))
	return logger
}

// GetComponentLogger returns a logger for the specified component, creating it if necessary
func GetComponentLogger(component string) *Logger {
	componentMutex.RLock()
	if logger, exists := componentLoggers[component]; exists {
		componentMutex.RUnlock()
		return logger
	}
	componentMutex.RUnlock()

	componentMutex.Lock()
	defer componentMutex.Unlock()

	// Double-check after acquiring write lock
	if logger, exists := componentLoggers[component]; exists {
		return logger
	}

	logger := NewComponentLogger(component)
	componentLoggers[component] = logger
	return logger
}

// init function to automatically initialize the logger when the package is imported
func init() {
	Initialize()
}

// configureFromEnvironment sets logger configuration based on environment variables
func configureFromEnvironment() {
	// Check LOG_LEVEL environment variable
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch strings.ToUpper(level) {
		case "DEBUG":
			AppLogger.SetLevel(DEBUG)
		case "INFO":
			AppLogger.SetLevel(INFO)
		case "WARN", "WARNING":
			AppLogger.SetLevel(WARN)
		case "ERROR":
			AppLogger.SetLevel(ERROR)
		}
	}

	// Check NO_COLOR environment variable (respects https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		AppLogger.SetColors(false)
	}

	// Check FORCE_COLOR environment variable
	if os.Getenv("FORCE_COLOR") != "" {
		AppLogger.ForceColors(true)
	}

	// Set application prefix
	AppLogger.SetPrefix("")
}
