package wsocket

import "log"

type Logger interface {
	Printf(format string, v ...interface{})
}

type defaultLogger struct{}

func (l *defaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

type NoLogger struct{}

func (l *NoLogger) Printf(format string, v ...interface{}) {}
