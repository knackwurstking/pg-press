// Package logger provides structured logging functionality for the pgpress application.
// It supports different log levels, colored output, and component-specific loggers.
package logger

import (
	"fmt"
	"os"
	"strings"
)

// AppLogger is the global logger instance for the application
var AppLogger *Logger

// Initialize sets up the application logger with appropriate defaults
func Initialize() {
	// Create logger with sensible defaults
	AppLogger = New(os.Stderr, WARN, true)

	// Configure based on environment
	configureFromEnvironment()

	// Replace standard log package
	InitializeFromStandardLog()
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
		AppLogger.SetColors(true)
	}

	// Set application prefix
	AppLogger.SetPrefix("")
}

// Component-specific logger helpers

// NewComponentLogger creates a new logger for a specific component
func NewComponentLogger(component string) *Logger {
	logger := New(os.Stderr, AppLogger.level, AppLogger.enableColors)
	logger.SetPrefix(fmt.Sprintf("[%s] ", component))
	return logger
}

// Server returns a logger configured for server operations
func Server() *Logger {
	return NewComponentLogger("Server")
}

// WebSocket returns a logger configured for WebSocket operations
func WebSocket() *Logger {
	return NewComponentLogger("WebSocket")
}

// Middleware returns a logger configured for middleware operations
func Middleware() *Logger {
	return NewComponentLogger("Middleware")
}

// Auth returns a logger configured for authentication operations
func Auth() *Logger {
	return NewComponentLogger("Auth")
}

// Feed returns a logger configured for feed operations
func Feed() *Logger {
	return NewComponentLogger("Feed")
}

// TroubleReport returns a logger configured for trouble report operations
func TroubleReport() *Logger {
	return NewComponentLogger("TroubleReport")
}

// User returns a logger configured for user operations
func User() *Logger {
	return NewComponentLogger("User")
}

// Cookie returns a logger configured for cookie operations
func Cookie() *Logger {
	return NewComponentLogger("Cookie")
}

// init function to automatically initialize the logger when the package is imported
func init() {
	Initialize()
}
