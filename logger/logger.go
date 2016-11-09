package logger

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

const (
	Debug int = iota
	Info
	Warning
	Error
	Critical
)

type Logger struct {
	context context.Context
}

func NewLogger(context context.Context) *Logger {
	return &Logger{context}
}

func (logger Logger) Log(level int, messages ...interface{}) {
	logger.LogF(level, "%v", messages)
}
func (logger Logger) LogF(level int, format string, messages ...interface{}) {
	switch level {
	case Debug:
		log.Debugf(logger.context, format, messages...)
	case Info:
		log.Infof(logger.context, format, messages...)
	case Warning:
		log.Warningf(logger.context, format, messages...)
	case Error:
		log.Errorf(logger.context, format, messages...)
	case Critical:
		log.Criticalf(logger.context, format, messages...)
	}
}
