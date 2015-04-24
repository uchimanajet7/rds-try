package logger

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestGetLogger(t *testing.T) {
	logger := GetLogger("logger-test")

	if logger.module != "logger-test" {
		t.Errorf("module name not match: %s", logger.module)
	}
	if logger.GetLogLevel() != "info" {
		t.Errorf("initial log level not match: %s", logger.GetLogLevel())
	}
}

func TestSetLogLevelDebug(t *testing.T) {
	logger := GetLogger("logger-test")
	logger.SetLogLevelDebug()

	if logger.GetLogLevel() != "debug" {
		t.Errorf("log level not match: %s", logger.GetLogLevel())
	}
}

func TestSetLogLevelInfo(t *testing.T) {
	logger := GetLogger("logger-test")
	logger.SetLogLevelInfo()

	if logger.GetLogLevel() != "info" {
		t.Errorf("log level not match: %s", logger.GetLogLevel())
	}
}

func TestSetJsonLogFormat(t *testing.T) {
	logger := GetLogger("logger-test")
	logger.SetJSONLogFormat()

	// apply only to file output
	if !reflect.DeepEqual(logger.loggers[fileNameText].Formatter, &logrus.JSONFormatter{}) {
		t.Errorf("config data not match: %+v/%+v", logger.loggers[fileNameText].Formatter, logrus.JSONFormatter{})
	}
}

func TestSetFileOutPut(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "logger-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}

	logger := GetLogger("logger-test")
	logger.SetFileOutPut(tempFile)
	logger.Infof("log file out put test infof")

	tempFile.Sync()
	fi, _ := tempFile.Stat()
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	if fi.Size() <= 0 {
		t.Errorf("size of the log file is zero: %d", fi.Size())
	}
}

func TestErrorf(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "logger-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}

	logger := GetLogger("logger-test")
	logger.SetFileOutPut(tempFile)
	logger.SetLogLevelInfo()
	logger.Errorf("log file out put test errorf")

	tempFile.Sync()
	fi, _ := tempFile.Stat()
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	if fi.Size() <= 0 {
		t.Errorf("size of the log file is zero: %d", fi.Size())
	}
}

func TestDebugf(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "logger-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}

	logger := GetLogger("logger-test")
	logger.SetFileOutPut(tempFile)
	logger.SetLogLevelDebug()
	logger.Debugf("log file out put test debugf")

	tempFile.Sync()
	fi, _ := tempFile.Stat()
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	if fi.Size() <= 0 {
		t.Errorf("size of the log file is zero: %d", fi.Size())
	}
}

func TestInfof(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "logger-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}

	logger := GetLogger("logger-test")
	logger.SetFileOutPut(tempFile)
	logger.SetLogLevelInfo()
	logger.Infof("log file out put test infof")

	tempFile.Sync()
	fi, _ := tempFile.Stat()
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	if fi.Size() <= 0 {
		t.Errorf("size of the log file is zero: %d", fi.Size())
	}
}
