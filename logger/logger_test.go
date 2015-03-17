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
	logger.SetJsonLogFormat()

	// apply only to file output
	if !reflect.DeepEqual(logger.loggers[file_name_text].Formatter, &logrus.JSONFormatter{}) {
		t.Errorf("config data not match: %+v/%+v", logger.loggers[file_name_text].Formatter, logrus.JSONFormatter{})
	}
}

func TestSetFileOutPut(t *testing.T) {
	temp_file, err := ioutil.TempFile("", "logger-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}

	logger := GetLogger("logger-test")
	logger.SetFileOutPut(temp_file)
	logger.Infof("log file out put test infof")

	temp_file.Sync()
	fi, _ := temp_file.Stat()
	temp_file.Close()
	defer os.Remove(temp_file.Name())

	if fi.Size() <= 0 {
		t.Errorf("size of the log file is zero: %d", fi.Size())
	}
}

func TestErrorf(t *testing.T) {
	temp_file, err := ioutil.TempFile("", "logger-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}

	logger := GetLogger("logger-test")
	logger.SetFileOutPut(temp_file)
	logger.SetLogLevelInfo()
	logger.Errorf("log file out put test errorf")

	temp_file.Sync()
	fi, _ := temp_file.Stat()
	temp_file.Close()
	defer os.Remove(temp_file.Name())

	if fi.Size() <= 0 {
		t.Errorf("size of the log file is zero: %d", fi.Size())
	}
}

func TestDebugf(t *testing.T) {
	temp_file, err := ioutil.TempFile("", "logger-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}

	logger := GetLogger("logger-test")
	logger.SetFileOutPut(temp_file)
	logger.SetLogLevelDebug()
	logger.Debugf("log file out put test debugf")

	temp_file.Sync()
	fi, _ := temp_file.Stat()
	temp_file.Close()
	defer os.Remove(temp_file.Name())

	if fi.Size() <= 0 {
		t.Errorf("size of the log file is zero: %d", fi.Size())
	}
}

func TestInfof(t *testing.T) {
	temp_file, err := ioutil.TempFile("", "logger-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}

	logger := GetLogger("logger-test")
	logger.SetFileOutPut(temp_file)
	logger.SetLogLevelInfo()
	logger.Infof("log file out put test infof")

	temp_file.Sync()
	fi, _ := temp_file.Stat()
	temp_file.Close()
	defer os.Remove(temp_file.Name())

	if fi.Size() <= 0 {
		t.Errorf("size of the log file is zero: %d", fi.Size())
	}
}
