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

// NewComponentLogger creates a new logger for a specific component
func NewComponentLogger(component string) *Logger {
	logger := New(os.Stderr, AppLogger.level, AppLogger.enableColors)
	logger.SetPrefix(fmt.Sprintf("[%s] ", component))
	return logger
}

// Middleware returns a logger configured for middleware operations
func Middleware() *Logger {
	return NewComponentLogger("Middleware")
}

// Server returns a logger configured for server operations
func Server() *Logger {
	return NewComponentLogger("Server")
}

func WSFeedHandler() *Logger {
	return NewComponentLogger("WS Feed Handler")
}

func WSFeedConnection() *Logger {
	return NewComponentLogger("WS Feed Connection")
}

func DBAttachments() *Logger {
	return NewComponentLogger("DB Attachments")
}

func DBCookies() *Logger {
	return NewComponentLogger("DB Cookies")
}

func DBFeeds() *Logger {
	return NewComponentLogger("DB Feeds")
}

func DBTools() *Logger {
	return NewComponentLogger("DB Tools")
}

func DBToolCyclesHelper() *Logger {
	return NewComponentLogger("DB ToolCyclesHelper")
}

func DBUsers() *Logger {
	return NewComponentLogger("DB Users")
}

func DBTroubleReportsHelper() *Logger {
	return NewComponentLogger("DB TroubleReportsHelper")
}

func DBMetalSheets() *Logger {
	return NewComponentLogger("DB MetalSheets")
}

func DBNotes() *Logger {
	return NewComponentLogger("DB Notes")
}

func DBTroubleReports() *Logger {
	return NewComponentLogger("DB TroubleReports")
}

func DBToolsHelper() *Logger {
	return NewComponentLogger("DB ToolsHelper")
}

func HandlerAuth() *Logger {
	return NewComponentLogger("Handler Auth")
}

func HandlerFeed() *Logger {
	return NewComponentLogger("Handler Feed")
}

func HandlerHome() *Logger {
	return NewComponentLogger("Handler Home")
}

func HandlerProfile() *Logger {
	return NewComponentLogger("Handler Profile")
}

func HandlerTools() *Logger {
	return NewComponentLogger("Handler Tools")
}

func HandlerTroubleReports() *Logger {
	return NewComponentLogger("Handler TroubleReports")
}

func HTMXHandlerFeed() *Logger {
	return NewComponentLogger("HTMX Handler Feed")
}

func HTMXHandlerNav() *Logger {
	return NewComponentLogger("HTMX Handler Nav")
}

func HTMXHandlerProfile() *Logger {
	return NewComponentLogger("HTMX Handler Profile")
}

func HTMXHandlerTools() *Logger {
	return NewComponentLogger("HTMX Handler Tools")
}

func HTMXHandlerTroubleReports() *Logger {
	return NewComponentLogger("HTMX Handler TroubleReports")
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
		AppLogger.SetColors(true)
	}

	// Set application prefix
	AppLogger.SetPrefix("")
}
