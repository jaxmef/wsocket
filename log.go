package wsocket

import "log"

type Logger interface {
	Printf(format string, v ...interface{})
}

func DefaultLogger() Logger {
	return &defaultLogger{}
}

type defaultLogger struct{}

func (l *defaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func NoLogger() Logger {
	return &noLogger{}
}

type noLogger struct{}

func (l *noLogger) Printf(format string, v ...interface{}) {}
