package logger

import (
	"fmt"
	"log"
	"os"
)

type LogLevel int

const (
	Unknown LogLevel = iota
	Verbose
	Debug
	Info
	Warn
	Error
	Disable
)

type LoggerProvider interface {
	Verbose(message string, args ...interface{})
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
}

type Logger struct {
	level  LogLevel
	loggerProvider LoggerProvider
}

func New(level LogLevel, loggerProvider LoggerProvider) *Logger {
	return &Logger{level: level, loggerProvider: loggerProvider}
}

func (l *Logger) Verbose(format string, args ...interface{}) {
	if l.shouldLog(Verbose) {
		l.loggerProvider.Verbose(format, args)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.shouldLog(Debug) {
		l.loggerProvider.Debug(format, args)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if l.shouldLog(Info) {
		l.loggerProvider.Info(format, args)
	}
}

func (l *Logger) Warn(format string, args ...interface{}) {
	if l.shouldLog(Warn) {
		l.loggerProvider.Warn(format, args)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if l.shouldLog(Error) {
		l.loggerProvider.Error(format, args)
	}
}

func (l *Logger) shouldLog(level LogLevel) bool {
	return l.level <= level
}

type defaultLoggerProvider struct {
	logger *log.Logger
} 

func NewDefault() LoggerProvider {
	return &defaultLoggerProvider{logger: log.New(os.Stderr, "", log.LstdFlags)} 
}

func (l *defaultLoggerProvider) Verbose(format string, args ...interface{}) {
	format = fmt.Sprintf("VERBOSE - %v\n", format)
	l.logger.Printf(format, args...)
}

func (l *defaultLoggerProvider) Debug(format string, args ...interface{}) {
	format = fmt.Sprintf("DEBUG - %v\n", format)
	l.logger.Printf(format, args...)
}

func (l *defaultLoggerProvider) Info(format string, args ...interface{}) {
	format = fmt.Sprintf("INFO - %v\n", format)
	l.logger.Printf(format, args...)
}

func (l *defaultLoggerProvider) Warn(format string, args ...interface{}) {
	format = fmt.Sprintf("WARN - %v\n", format)
	l.logger.Printf(format, args...)
}

func (l *defaultLoggerProvider) Error(format string, args ...interface{}) {
	format = fmt.Sprintf("ERROR - %v\n", format)
	l.logger.Printf(format, args...)
}

