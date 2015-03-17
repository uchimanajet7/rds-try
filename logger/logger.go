package logger

import (
	"io"
	"runtime"
	"sync"

	"github.com/Sirupsen/logrus"
)

type Logger struct {
	loggers map[string]*logrus.Logger
	module  string
}

const file_name_text = "rt_file"
const cli_name_text = "rt_cli"

var loggers = map[string]*logrus.Logger{
	file_name_text: logrus.New(),
	cli_name_text:  logrus.New(),
}

var once = new(sync.Once)

func GetLogger(name string) *Logger {
	logger := &Logger{}
	logger.loggers = loggers
	logger.module = name

	// run once
	once.Do(func() {
		for key, logger := range loggers {
			logger.Level = logrus.InfoLevel

			switch key {
			case cli_name_text:
				logger.Formatter = &logrus.TextFormatter{DisableColors: false}
				if runtime.GOOS == "windows" {
					logger.Formatter = &logrus.TextFormatter{DisableColors: true}
				}
			case file_name_text:
				logger.Formatter = &logrus.TextFormatter{DisableColors: true}
			}
		}
	})

	return logger
}

func (l *Logger) SetLogLevelDebug() {
	for _, logger := range l.loggers {
		logger.Level = logrus.DebugLevel
	}
}

func (l *Logger) SetLogLevelInfo() {
	for _, logger := range l.loggers {
		logger.Level = logrus.InfoLevel
	}
}

func (l *Logger) SetJsonLogFormat() {
	// setting target log file only
	for key, logger := range l.loggers {
		switch key {
		case file_name_text:
			logger.Formatter = &logrus.JSONFormatter{}
		}
	}
}

func (l *Logger) SetFileOutPut(out io.Writer) {
	// setting target log file only
	for key, logger := range l.loggers {
		switch key {
		case file_name_text:
			logger.Out = out
		}
	}
}

func (l *Logger) GetLogLevel() string {
	return l.loggers[file_name_text].Level.String()
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	for _, logger := range l.loggers {
		logger.WithField("module", l.module).Errorf(format, args...)
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	for _, logger := range l.loggers {
		logger.WithField("module", l.module).Debugf(format, args...)
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	for _, logger := range l.loggers {
		logger.WithField("module", l.module).Infof(format, args...)
	}
}
