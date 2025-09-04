package logger

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// countingWriter counts the number of Write calls for concurrency testing
type countingWriter struct {
	count int64
}

func (c *countingWriter) Write(p []byte) (n int, err error) {
	atomic.AddInt64(&c.count, 1)
	return len(p), nil
}

func (c *countingWriter) Count() int64 {
	return atomic.LoadInt64(&c.count)
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, INFO, false)

	// Test that DEBUG messages are filtered out
	logger.Debug("debug message")
	if buf.Len() != 0 {
		t.Error("Debug message should not be logged when level is INFO")
	}

	// Test that INFO messages are logged
	logger.Info("info message")
	if buf.Len() == 0 {
		t.Error("Info message should be logged when level is INFO")
	}

	// Reset buffer
	buf.Reset()

	// Test that WARN messages are logged
	logger.Warn("warn message")
	if buf.Len() == 0 {
		t.Error("Warn message should be logged when level is INFO")
	}

	// Reset buffer
	buf.Reset()

	// Test that ERROR messages are logged
	logger.Error("error message")
	if buf.Len() == 0 {
		t.Error("Error message should be logged when level is INFO")
	}
}

func TestLoggerEarlyExit(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, ERROR, false)

	// These should exit early without expensive formatting
	logger.Debug("expensive %s formatting %d", strings.Repeat("test", 1000), 12345)
	logger.Info("expensive %s formatting %d", strings.Repeat("test", 1000), 12345)
	logger.Warn("expensive %s formatting %d", strings.Repeat("test", 1000), 12345)

	// Buffer should be empty since all are below ERROR level
	if buf.Len() != 0 {
		t.Error("No messages should be logged when level is ERROR")
	}

	// This should be logged
	logger.Error("error message")
	if buf.Len() == 0 {
		t.Error("Error message should be logged")
	}
}

func TestLoggerConcurrency(t *testing.T) {
	// Use a counting writer instead of parsing buffer output
	counter := &countingWriter{}
	logger := New(counter, DEBUG, false)

	var wg sync.WaitGroup
	numGoroutines := 100
	messagesPerGoroutine := 10

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info("goroutine %d message %d", id, j)
			}
		}(i)
	}

	wg.Wait()

	expectedWrites := int64(numGoroutines * messagesPerGoroutine)
	actualWrites := counter.Count()
	if actualWrites != expectedWrites {
		t.Errorf("Expected %d writes, got %d", expectedWrites, actualWrites)
	}
}

func TestLoggerColors(t *testing.T) {
	var buf bytes.Buffer

	// Test with colors enabled - force colors since we're not writing to a terminal
	logger := New(&buf, DEBUG, false)
	logger.ForceColors(true) // Force enable colors for testing
	logger.Info("test message")
	output := buf.String()

	// Should contain ANSI color codes when colors are enabled
	if !strings.Contains(output, Blue) {
		t.Error("Expected color codes in output when colors enabled")
	}

	// Test with colors disabled
	buf.Reset()
	logger.SetColors(false)
	logger.Info("test message")
	output = buf.String()

	// Should not contain ANSI color codes when colors are disabled
	if strings.Contains(output, Blue) {
		t.Error("Should not contain color codes when colors disabled")
	}
}

func TestLoggerPrefix(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, INFO, false)

	prefix := "[TEST] "
	logger.SetPrefix(prefix)
	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, prefix) {
		t.Errorf("Expected prefix '%s' in output: %s", prefix, output)
	}
}

func TestComponentLogger(t *testing.T) {
	// Test that component loggers are cached and reused
	logger1 := GetComponentLogger("TestComponent")
	logger2 := GetComponentLogger("TestComponent")

	if logger1 != logger2 {
		t.Error("Component loggers should be cached and reused")
	}

	// Test that different components get different loggers
	logger3 := GetComponentLogger("DifferentComponent")
	if logger1 == logger3 {
		t.Error("Different components should get different loggers")
	}
}

func TestBackwardCompatibilityFunctions(t *testing.T) {
	// Test that all backward compatibility functions return valid loggers
	loggers := []struct {
		name   string
		logger *Logger
	}{
		{"Middleware", Middleware()},
		{"Server", Server()},
		{"WSFeedHandler", WSFeedHandler()},
		{"DBAttachments", DBAttachments()},
		{"HandlerAuth", HandlerAuth()},
	}

	for _, l := range loggers {
		if l.logger == nil {
			t.Errorf("Logger function %s returned nil", l.name)
		}
	}

	// Test that calling the same function twice returns the same logger (cached)
	server1 := Server()
	server2 := Server()
	if server1 != server2 {
		t.Error("Server() should return the same cached logger instance")
	}
}

func BenchmarkLoggerFilteredOut(b *testing.B) {
	var buf bytes.Buffer
	logger := New(&buf, ERROR, false) // Set to ERROR level, so DEBUG won't be logged

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This should exit early due to level filtering
		logger.Debug("debug message %d with expensive formatting %s", i, strings.Repeat("test", 100))
	}
}

func BenchmarkLoggerNotFiltered(b *testing.B) {
	var buf bytes.Buffer
	logger := New(&buf, DEBUG, false) // Set to DEBUG level, so DEBUG will be logged

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will go through the full logging pipeline
		logger.Debug("debug message %d with expensive formatting %s", i, strings.Repeat("test", 100))
	}
}

func BenchmarkComponentLoggerCaching(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetComponentLogger("BenchmarkComponent")
	}
}

func BenchmarkLoggerConcurrency(b *testing.B) {
	counter := &countingWriter{}
	logger := New(counter, INFO, false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("concurrent logging test message")
		}
	})
}

func TestLoggerPerformanceWithHighVolume(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, INFO, false)

	start := time.Now()
	numMessages := 10000

	for i := 0; i < numMessages; i++ {
		logger.Info("message %d", i)
	}

	elapsed := time.Since(start)
	messagesPerSecond := float64(numMessages) / elapsed.Seconds()

	t.Logf("Logged %d messages in %v (%.0f msg/sec)", numMessages, elapsed, messagesPerSecond)

	// Basic performance check - should be able to log at least 10k messages per second
	if messagesPerSecond < 10000 {
		t.Errorf("Logger performance too slow: %.0f msg/sec (expected > 10000)", messagesPerSecond)
	}
}

func TestHTTPRequestLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, DEBUG, false)

	// Test different status codes
	testCases := []struct {
		status   int
		method   string
		uri      string
		remoteIP string
		latency  time.Duration
		userInfo string
		desc     string
	}{
		{200, "GET", "/api/users", "127.0.0.1", 15 * time.Millisecond, " user@example.com", "successful GET request"},
		{404, "GET", "/missing", "192.168.1.1", 5 * time.Millisecond, "", "not found error"},
		{500, "POST", "/api/data", "10.0.0.1", 100 * time.Millisecond, " admin@test.com", "server error"},
		{201, "POST", "/api/users", "127.0.0.1", 25 * time.Millisecond, "", "created resource"},
		{302, "GET", "/redirect", "127.0.0.1", 2 * time.Millisecond, "", "redirect response"},
	}

	for _, tc := range testCases {
		buf.Reset()
		logger.HTTPRequest(tc.status, tc.method, tc.uri, tc.remoteIP, tc.latency, tc.userInfo)

		output := buf.String()
		if output == "" {
			t.Errorf("HTTPRequest should produce output for %s", tc.desc)
			continue
		}

		// Check that essential information is present
		if !strings.Contains(output, tc.method) {
			t.Errorf("Output should contain method %s for %s", tc.method, tc.desc)
		}
		if !strings.Contains(output, tc.uri) {
			t.Errorf("Output should contain URI %s for %s", tc.uri, tc.desc)
		}
		if !strings.Contains(output, tc.remoteIP) {
			t.Errorf("Output should contain remote IP %s for %s", tc.remoteIP, tc.desc)
		}
		if !strings.Contains(output, "[HTTP]") {
			t.Errorf("Output should contain [HTTP] marker for %s", tc.desc)
		}
		if !strings.Contains(output, fmt.Sprintf("%d", tc.status)) {
			t.Errorf("Output should contain status code %d for %s", tc.status, tc.desc)
		}
		if tc.userInfo != "" && !strings.Contains(output, strings.TrimSpace(tc.userInfo)) {
			t.Errorf("Output should contain user info for %s", tc.desc)
		}
	}
}

func TestHTTPRequestLoggingWithColors(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, DEBUG, false)
	logger.ForceColors(true) // Force colors for testing

	// Test that colored output contains ANSI codes
	logger.HTTPRequest(200, "GET", "/test", "127.0.0.1", 10*time.Millisecond, "")
	output := buf.String()

	// Should contain some ANSI color codes when colors are enabled
	hasColors := strings.Contains(output, "\033[") || strings.Contains(output, "✅")
	if !hasColors {
		t.Error("Colored HTTP request output should contain ANSI codes or emojis")
	}

	// Test without colors
	buf.Reset()
	logger.SetColors(false)
	logger.HTTPRequest(200, "GET", "/test", "127.0.0.1", 10*time.Millisecond, "")
	output = buf.String()

	// Should not contain ANSI color codes when colors are disabled
	if strings.Contains(output, "\033[") {
		t.Error("Non-colored HTTP request output should not contain ANSI codes")
	}
}

func TestHTTPRequestLoggingLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, WARN, false) // Set to WARN level

	// INFO level request (200) should be filtered out
	logger.HTTPRequest(200, "GET", "/test", "127.0.0.1", 10*time.Millisecond, "")
	if buf.Len() != 0 {
		t.Error("200 status request should be filtered out when logger level is WARN")
	}

	// WARN level request (400) should be logged
	logger.HTTPRequest(400, "GET", "/bad-request", "127.0.0.1", 5*time.Millisecond, "")
	if buf.Len() == 0 {
		t.Error("400 status request should be logged when logger level is WARN")
	}

	buf.Reset()

	// ERROR level request (500) should be logged
	logger.HTTPRequest(500, "POST", "/error", "127.0.0.1", 50*time.Millisecond, "")
	if buf.Len() == 0 {
		t.Error("500 status request should be logged when logger level is WARN")
	}
}
