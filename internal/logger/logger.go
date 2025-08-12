package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	Bold   = "\033[1m"
	Italic = "\033[3m"
)

// Logger represents a colored logger instance
type Logger struct {
	mu           sync.Mutex
	output       io.Writer
	level        LogLevel
	enableColors bool
	prefix       string
}

// Default logger instance
var defaultLogger *Logger

func init() {
	defaultLogger = New(os.Stderr, INFO, true)
}

// New creates a new Logger instance
func New(output io.Writer, level LogLevel, enableColors bool) *Logger {
	// Auto-detect if we should enable colors based on terminal support
	if enableColors {
		enableColors = isTerminal(output)
	}

	return &Logger{
		output:       output,
		level:        level,
		enableColors: enableColors,
	}
}

// isTerminal checks if the output is a terminal
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		// Simple check for common terminal file descriptors
		return f == os.Stdout || f == os.Stderr
	}
	return false
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetColors enables or disables colored output
func (l *Logger) SetColors(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enableColors = enabled && isTerminal(l.output)
}

// SetPrefix sets a prefix for all log messages
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// levelToString converts log level to string representation
func (l *Logger) levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// levelToColor returns the ANSI color code for a log level
func (l *Logger) levelToColor(level LogLevel) string {
	if !l.enableColors {
		return ""
	}

	switch level {
	case DEBUG:
		return Green
	case INFO:
		return Blue
	case WARN:
		return Yellow
	case ERROR:
		return Red
	default:
		return Reset
	}
}

// log is the internal logging function
func (l *Logger) log(level LogLevel, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if we should log this level
	if level < l.level {
		return
	}

	// Format the message
	message := fmt.Sprintf(format, args...)

	// Get current time
	timestamp := time.Now().Format("2006/01/02 15:04:05")

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	var caller string
	if ok {
		// Extract just the filename from the full path
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			caller = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
		}
	}

	// Build the log message
	levelStr := l.levelToString(level)
	color := l.levelToColor(level)
	reset := ""
	if l.enableColors {
		reset = Reset
	}

	var logLine string
	if caller != "" {
		italic := ""
		italicReset := ""

		if level == DEBUG {
			italic = Italic
			italicReset = Reset
		}

		logLine = fmt.Sprintf("%s%s[%s]%s %s %s [%s] %s\n",
			color, Bold, levelStr, reset,
			timestamp,
			l.prefix,
			caller,
			italic+message+italicReset)

	} else {
		logLine = fmt.Sprintf("%s%s[%s]%s %s %s %s\n",
			color, Bold, levelStr, reset,
			timestamp,
			l.prefix,
			message)
	}

	l.output.Write([]byte(logLine))
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...any) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...any) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...any) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...any) {
	l.log(ERROR, format, args...)
}

// Printf provides compatibility with standard log.Printf
// It logs at INFO level by default
func (l *Logger) Printf(format string, args ...any) {
	l.log(INFO, format, args...)
}

// InitializeFromStandardLog configures the logger to replace standard log package
func InitializeFromStandardLog() {
	// Set the standard log package to use our logger
	log.SetOutput(&loggerWriter{defaultLogger})
	log.SetFlags(0) // Disable standard log formatting since we handle it
}

// loggerWriter implements io.Writer to bridge with standard log package
type loggerWriter struct {
	logger *Logger
}

func (w *loggerWriter) Write(p []byte) (n int, err error) {
	// Remove trailing newline if present since our logger adds it
	message := strings.TrimSuffix(string(p), "\n")
	w.logger.Info("%s", message)
	return len(p), nil
}
