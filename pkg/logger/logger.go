package logger

import (
	"fmt"
	"log"
	"os"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	ERROR
)

type Logger struct {
	level       Level
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

func New(level string) *Logger {
	l := &Logger{
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	switch level {
	case "debug":
		l.level = DEBUG
	case "error":
		l.level = ERROR
	default:
		l.level = INFO
	}

	return l
}

func (l *Logger) log(level Level, logger *log.Logger, format string, v ...interface{}) {
	if level >= l.level {
		logger.Output(3, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.log(INFO, l.infoLogger, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ERROR, l.errorLogger, format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DEBUG, l.debugLogger, format, v...)
}