package logger

import (
	"fmt"
	"log"
	"os"
)

type Level int

const (
	Verbose Level = iota
	Debug
	Error
)

type Log struct {
	logger *log.Logger
	level  Level
}

func New(debug bool) *Log {
	var level Level
	if debug {
		level = Debug
	} else {
		level = Error
	}
	return &Log{
		logger: log.New(os.Stderr, "", log.LstdFlags),
		level:  level,
	}
}

func (l *Log) Verbose(format string, args ...interface{}) {
	if l.level <= Verbose {
		format = fmt.Sprintf("DEBUG - %v\n", format)
		l.logger.Printf(format, args...)
	}
}

func (l *Log) Debug(format string, args ...interface{}) {
	if l.level <= Debug {
		format = fmt.Sprintf("DEBUG - %v\n", format)
		l.logger.Printf(format, args...)
	}
}

func (l *Log) Error(format string, args ...interface{}) {
	if l.level <= Error {
		format = fmt.Sprintf("ERROR - %v\n", format)
		l.logger.Printf(format, args...)
	}
}
