// Package logger provides structured logging functionality for the pgpress application.
// It supports different log levels, colored output, and component-specific loggers.
package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
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

var (
	componentLoggers = make(map[string]*Logger)
	componentMutex   sync.RWMutex
)

// NewComponentLogger creates a new logger for a specific component
func NewComponentLogger(component string) *Logger {
	logger := New(os.Stderr, AppLogger.level, AppLogger.enableColors)
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

// Backward compatibility functions with original names
func Middleware() *Logger                { return GetComponentLogger("Middleware") }
func Server() *Logger                    { return GetComponentLogger("Server") }
func WSFeedHandler() *Logger             { return GetComponentLogger("WS Feed Handler") }
func WSFeedConnection() *Logger          { return GetComponentLogger("WS Feed Connection") }
func DBAttachments() *Logger             { return GetComponentLogger("DB Attachments") }
func DBCookies() *Logger                 { return GetComponentLogger("DB Cookies") }
func DBFeeds() *Logger                   { return GetComponentLogger("DB Feeds") }
func DBTools() *Logger                   { return GetComponentLogger("DB Tools") }
func DBUsers() *Logger                   { return GetComponentLogger("DB Users") }
func DBMetalSheets() *Logger             { return GetComponentLogger("DB MetalSheets") }
func DBNotes() *Logger                   { return GetComponentLogger("DB Notes") }
func DBTroubleReports() *Logger          { return GetComponentLogger("DB TroubleReports") }
func DBPressCycles() *Logger             { return GetComponentLogger("DB Service PressCycles") }
func DBToolRegenerations() *Logger       { return GetComponentLogger("DB ToolRegenerations") }
func HandlerAuth() *Logger               { return GetComponentLogger("Handler Auth") }
func HandlerFeed() *Logger               { return GetComponentLogger("Handler Feed") }
func HandlerHome() *Logger               { return GetComponentLogger("Handler Home") }
func HandlerProfile() *Logger            { return GetComponentLogger("Handler Profile") }
func HandlerTools() *Logger              { return GetComponentLogger("Handler Tools") }
func HandlerTroubleReports() *Logger     { return GetComponentLogger("Handler TroubleReports") }
func HTMXHandlerFeed() *Logger           { return GetComponentLogger("HTMX Handler Feed") }
func HTMXHandlerNav() *Logger            { return GetComponentLogger("HTMX Handler Nav") }
func HTMXHandlerProfile() *Logger        { return GetComponentLogger("HTMX Handler Profile") }
func HTMXHandlerTools() *Logger          { return GetComponentLogger("HTMX Handler Tools") }
func HTMXHandlerTroubleReports() *Logger { return GetComponentLogger("HTMX Handler TroubleReports") }

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
