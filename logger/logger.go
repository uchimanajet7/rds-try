package logger

import (
	"io"
	"runtime"
	"sync"

	"github.com/Sirupsen/logrus"
)

// Logger struct is loggers map and module variable
type Logger struct {
	loggers map[string]*logrus.Logger
	module  string
}

const fileNameText = "rt_file"
const cliNameText = "rt_cli"

var loggers = map[string]*logrus.Logger{
	fileNameText: logrus.New(),
	cliNameText:  logrus.New(),
}

var once = new(sync.Once)

// GetLogger is object for log printout is acquired.
func GetLogger(name string) *Logger {
	logger := &Logger{}
	logger.loggers = loggers
	logger.module = name

	// run once
	once.Do(func() {
		for key, logger := range loggers {
			logger.Level = logrus.InfoLevel

			switch key {
			case cliNameText:
				logger.Formatter = &logrus.TextFormatter{DisableColors: false}
				if runtime.GOOS == "windows" {
					logger.Formatter = &logrus.TextFormatter{DisableColors: true}
				}
			case fileNameText:
				logger.Formatter = &logrus.TextFormatter{DisableColors: true}
			}
		}
	})

	return logger
}

// SetLogLevelDebug is the output level of log is changed to debug.
func (l *Logger) SetLogLevelDebug() {
	for _, logger := range l.loggers {
		logger.Level = logrus.DebugLevel
	}
}

// SetLogLevelInfo is the output level of log is changed to info.
func (l *Logger) SetLogLevelInfo() {
	for _, logger := range l.loggers {
		logger.Level = logrus.InfoLevel
	}
}

// SetJSONLogFormat is the output log format "JSON"
func (l *Logger) SetJSONLogFormat() {
	// setting target log file only
	for key, logger := range l.loggers {
		switch key {
		case fileNameText:
			logger.Formatter = &logrus.JSONFormatter{}
		}
	}
}

// SetFileOutPut is the output to file
func (l *Logger) SetFileOutPut(out io.Writer) {
	// setting target log file only
	for key, logger := range l.loggers {
		switch key {
		case fileNameText:
			logger.Out = out
		}
	}
}

// GetLogLevel is the return current log level state
func (l *Logger) GetLogLevel() string {
	return l.loggers[fileNameText].Level.String()
}

// Errorf is the output error level log text
func (l *Logger) Errorf(format string, args ...interface{}) {
	for _, logger := range l.loggers {
		logger.WithField("module", l.module).Errorf(format, args...)
	}
}

// Debugf is the output debug level log text
func (l *Logger) Debugf(format string, args ...interface{}) {
	for _, logger := range l.loggers {
		logger.WithField("module", l.module).Debugf(format, args...)
	}
}

// Infof is the output info level log text
func (l *Logger) Infof(format string, args ...interface{}) {
	for _, logger := range l.loggers {
		logger.WithField("module", l.module).Infof(format, args...)
	}
}
