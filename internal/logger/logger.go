package logger

import (
	"fmt"
	"io"
	"log"
	"os"
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
	Reset     = "\033[0m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Purple    = "\033[35m"
	Cyan      = "\033[36m"
	Gray      = "\033[37m"
	Bold      = "\033[1m"
	Italic    = "\033[3m"
	BgBlue    = "\033[44m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgRed     = "\033[41m"
	BgCyan    = "\033[46m"
	Underline = "\033[4m"
)

// Logger represents a colored logger instance
type Logger struct {
	mu              sync.Mutex
	output          io.Writer
	level           LogLevel
	enableColors    bool
	colorPreference bool
	prefix          string
}

// Default logger instance
var defaultLogger *Logger

func init() {
	defaultLogger = New(os.Stderr, INFO, true)
}

// New creates a new Logger instance
func New(output io.Writer, level LogLevel, enableColors bool) *Logger {
	l := &Logger{
		output:          output,
		level:           level,
		colorPreference: enableColors,
	}
	l.SetColors(enableColors)
	return l
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

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
	l.enableColors = l.colorPreference && isTerminal(l.output)
}

// SetColors enables or disables colored output
func (l *Logger) SetColors(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.colorPreference = enabled
	l.enableColors = enabled && isTerminal(l.output)
}

// ForceColors enables or disables colored output, bypassing terminal detection
func (l *Logger) ForceColors(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.colorPreference = enabled
	l.enableColors = enabled
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
	switch level {
	case DEBUG:
		return Green + Italic
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
	// Fast path: check level without lock first
	l.mu.Lock()
	currentLevel := l.level
	enableColors := l.enableColors
	prefix := l.prefix
	output := l.output
	l.mu.Unlock()

	// Check if we should log this level
	if level < currentLevel {
		return
	}

	// Format the message
	message := fmt.Sprintf(format, args...)

	// Get current time
	timestamp := time.Now().Format("2006/01/02 15:04:05")

	// Build the log message
	levelStr := l.levelToString(level)
	color := ""
	reset := ""
	if enableColors {
		color = l.levelToColor(level)
		reset = Reset
	}

	logLine := fmt.Sprintf("%s[%-5s]%s %s %s %s\n",
		color+Bold, levelStr, reset,
		timestamp,
		prefix,
		color+message+reset)

	output.Write([]byte(logLine))
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...any) {
	// Early exit for performance if level too low
	l.mu.Lock()
	currentLevel := l.level
	l.mu.Unlock()
	if DEBUG < currentLevel {
		return
	}
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...any) {
	// Early exit for performance if level too low
	l.mu.Lock()
	currentLevel := l.level
	l.mu.Unlock()
	if INFO < currentLevel {
		return
	}
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...any) {
	// Early exit for performance if level too low
	l.mu.Lock()
	currentLevel := l.level
	l.mu.Unlock()
	if WARN < currentLevel {
		return
	}
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...any) {
	// Early exit for performance if level too low
	l.mu.Lock()
	currentLevel := l.level
	l.mu.Unlock()
	if ERROR < currentLevel {
		return
	}
	l.log(ERROR, format, args...)
}

// HTTPRequest logs an HTTP request with special highlighting
func (l *Logger) HTTPRequest(status int, method, uri, remoteIP string, latency time.Duration, userInfo string) {
	// Determine log level based on status code
	var level LogLevel
	if status >= 500 {
		level = ERROR
	} else if status >= 400 {
		level = WARN
	} else {
		level = INFO
	}

	// Early exit for performance if level too low
	l.mu.Lock()
	currentLevel := l.level
	enableColors := l.enableColors
	prefix := l.prefix
	output := l.output
	l.mu.Unlock()

	if level < currentLevel {
		return
	}

	// Get current time
	timestamp := time.Now().Format("2006/01/02 15:04:05")

	// Build the special HTTP log message with enhanced formatting
	var logLine string
	if enableColors {
		// Choose colors and symbols based on status code
		var statusColor, methodColor, httpIcon string

		if status >= 500 {
			statusColor = BgRed + Bold
			httpIcon = "🔥"
		} else if status >= 400 {
			statusColor = BgYellow + Bold
			httpIcon = "⚠️ "
		} else if status >= 300 {
			statusColor = BgCyan + Bold
			httpIcon = "↩️ "
		} else {
			statusColor = BgGreen + Bold
			httpIcon = "✅"
		}

		// Method color coding
		switch method {
		case "GET":
			methodColor = Blue + Bold
		case "POST":
			methodColor = Green + Bold
		case "PUT":
			methodColor = Yellow + Bold
		case "DELETE":
			methodColor = Red + Bold
		case "PATCH":
			methodColor = Purple + Bold
		default:
			methodColor = Cyan + Bold
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s %s ", httpIcon, timestamp))
		if prefix != "" {
			sb.WriteString(fmt.Sprintf("%s ", prefix))
		}
		sb.WriteString(fmt.Sprintf("%s%3d%s ", statusColor, status, Reset))
		sb.WriteString(fmt.Sprintf("%s%-7s%s ", methodColor, method, Reset))
		sb.WriteString(fmt.Sprintf("%s%s%s ", Underline+Cyan, uri, Reset))
		sb.WriteString(fmt.Sprintf("(%s%s%s) ", Gray, remoteIP, Reset))
		sb.WriteString(fmt.Sprintf("%s%v%s", Bold+Purple, latency, Reset))
		if userInfo != "" {
			sb.WriteString(userInfo)
		}
		sb.WriteString("\n")
		logLine = sb.String()
	} else {
		// Plain text version for when colors are disabled
		levelStr := l.levelToString(level)
		logLine = fmt.Sprintf("[%-5s] %s %s %d %-7s %s (%s) %v%s\n",
			levelStr, timestamp, prefix, status, method, uri, remoteIP, latency, userInfo)
	}

	output.Write([]byte(logLine))
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
