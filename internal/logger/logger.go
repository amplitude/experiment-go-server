package logger

import (
	"fmt"
	"log"
)

type Log struct {
	logger  *log.Logger
	isDebug bool
}

func New(debug bool) *Log {
	return &Log{
		logger:  log.Default(),
		isDebug: debug,
	}
}

func (l *Log) Debug(format string, args ...interface{}) {
	if l.isDebug {
		format = fmt.Sprintf("DEBUG - %v\n", format)
		l.logger.Printf(format, args...)
	}
}

func (l *Log) Error(format string, args ...interface{}) {
	format = fmt.Sprintf("ERROR - %v\n", format)
	l.logger.Printf(format, args...)
}
