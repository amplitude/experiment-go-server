package logger

import (
	"testing"
)

// mockLoggerProvider is a mock implementation for testing
type mockLoggerProvider struct {
	verboseCalls []logCall
	debugCalls   []logCall
	infoCalls    []logCall
	warnCalls    []logCall
	errorCalls   []logCall
}

type logCall struct {
	format string
	args   []interface{}
}

func newMockLoggerProvider() *mockLoggerProvider {
	return &mockLoggerProvider{
		verboseCalls: make([]logCall, 0),
		debugCalls:   make([]logCall, 0),
		infoCalls:    make([]logCall, 0),
		warnCalls:    make([]logCall, 0),
		errorCalls:   make([]logCall, 0),
	}
}

func (m *mockLoggerProvider) Verbose(format string, args ...interface{}) {
	m.verboseCalls = append(m.verboseCalls, logCall{format: format, args: args})
}

func (m *mockLoggerProvider) Debug(format string, args ...interface{}) {
	m.debugCalls = append(m.debugCalls, logCall{format: format, args: args})
}

func (m *mockLoggerProvider) Info(format string, args ...interface{}) {
	m.infoCalls = append(m.infoCalls, logCall{format: format, args: args})
}

func (m *mockLoggerProvider) Warn(format string, args ...interface{}) {
	m.warnCalls = append(m.warnCalls, logCall{format: format, args: args})
}

func (m *mockLoggerProvider) Error(format string, args ...interface{}) {
	m.errorCalls = append(m.errorCalls, logCall{format: format, args: args})
}

// TestNew tests the Logger constructor
func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		provider LoggerProvider
	}{
		{
			name:     "Create logger with Verbose level",
			level:    Verbose,
			provider: newMockLoggerProvider(),
		},
		{
			name:     "Create logger with Debug level",
			level:    Debug,
			provider: newMockLoggerProvider(),
		},
		{
			name:     "Create logger with Info level",
			level:    Info,
			provider: newMockLoggerProvider(),
		},
		{
			name:     "Create logger with Warn level",
			level:    Warn,
			provider: newMockLoggerProvider(),
		},
		{
			name:     "Create logger with Error level",
			level:    Error,
			provider: newMockLoggerProvider(),
		},
		{
			name:     "Create logger with Disable level",
			level:    Disable,
			provider: newMockLoggerProvider(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level, tt.provider)

			if logger == nil {
				t.Fatal("Expected non-nil logger")
			}

			if logger.level != tt.level {
				t.Errorf("Expected level %v, got %v", tt.level, logger.level)
			}

			if logger.loggerProvider != tt.provider {
				t.Error("Expected provider to be set correctly")
			}
		})
	}
}

// TestShouldLog tests the level filtering logic
func TestShouldLog(t *testing.T) {
	tests := []struct {
		name         string
		loggerLevel  LogLevel
		messageLevel LogLevel
		shouldLog    bool
	}{
		// Verbose level (1) - logs everything except Disable
		{"Verbose logger logs Verbose", Verbose, Verbose, true},
		{"Verbose logger logs Debug", Verbose, Debug, true},
		{"Verbose logger logs Info", Verbose, Info, true},
		{"Verbose logger logs Warn", Verbose, Warn, true},
		{"Verbose logger logs Error", Verbose, Error, true},

		// Debug level (2) - logs Debug and higher
		{"Debug logger blocks Verbose", Debug, Verbose, false},
		{"Debug logger logs Debug", Debug, Debug, true},
		{"Debug logger logs Info", Debug, Info, true},
		{"Debug logger logs Warn", Debug, Warn, true},
		{"Debug logger logs Error", Debug, Error, true},

		// Info level (3) - logs Info and higher
		{"Info logger blocks Verbose", Info, Verbose, false},
		{"Info logger blocks Debug", Info, Debug, false},
		{"Info logger logs Info", Info, Info, true},
		{"Info logger logs Warn", Info, Warn, true},
		{"Info logger logs Error", Info, Error, true},

		// Warn level (4) - logs Warn and higher
		{"Warn logger blocks Verbose", Warn, Verbose, false},
		{"Warn logger blocks Debug", Warn, Debug, false},
		{"Warn logger blocks Info", Warn, Info, false},
		{"Warn logger logs Warn", Warn, Warn, true},
		{"Warn logger logs Error", Warn, Error, true},

		// Error level (5) - logs only Error
		{"Error logger blocks Verbose", Error, Verbose, false},
		{"Error logger blocks Debug", Error, Debug, false},
		{"Error logger blocks Info", Error, Info, false},
		{"Error logger blocks Warn", Error, Warn, false},
		{"Error logger logs Error", Error, Error, true},

		// Disable level (6) - logs nothing
		{"Disable logger blocks Verbose", Disable, Verbose, false},
		{"Disable logger blocks Debug", Disable, Debug, false},
		{"Disable logger blocks Info", Disable, Info, false},
		{"Disable logger blocks Warn", Disable, Warn, false},
		{"Disable logger blocks Error", Disable, Error, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.loggerLevel, newMockLoggerProvider())
			result := logger.shouldLog(tt.messageLevel)

			if result != tt.shouldLog {
				t.Errorf("Expected shouldLog(%v) with logger level %v to be %v, got %v",
					tt.messageLevel, tt.loggerLevel, tt.shouldLog, result)
			}
		})
	}
}

// TestLoggerMultipleMessages tests calling multiple log methods
func TestLoggerMultipleMessages(t *testing.T) {
	mock := newMockLoggerProvider()
	logger := New(Debug, mock)

	logger.Debug("debug 1")
	logger.Debug("debug 2")
	logger.Info("info 1")
	logger.Warn("warn 1")
	logger.Error("error 1")

	if len(mock.debugCalls) != 2 {
		t.Errorf("Expected 2 Debug calls, got %d", len(mock.debugCalls))
	}
	if len(mock.infoCalls) != 1 {
		t.Errorf("Expected 1 Info call, got %d", len(mock.infoCalls))
	}
	if len(mock.warnCalls) != 1 {
		t.Errorf("Expected 1 Warn call, got %d", len(mock.warnCalls))
	}
	if len(mock.errorCalls) != 1 {
		t.Errorf("Expected 1 Error call, got %d", len(mock.errorCalls))
	}
}
