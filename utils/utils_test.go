package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetHomeDir(t *testing.T) {
	if GetHomeDir() == "" {
		t.Error("user home dir not found")
	}
}

func TestGetUserName(t *testing.T) {
	if GetUserName() == "" {
		t.Error("user name not found")
	}
}

func TestGetAppName(t *testing.T) {
	if app_name != GetAppName() {
		t.Error("app name not match")
	}
}

func TestGetFormatedTime(t *testing.T) {
	// check format "2015-01-20-18-03-35"
	strs := strings.Split(GetFormatedTime(), "-")

	if len(strs) != 6 {
		t.Errorf("time format not match : len(%d)", len(strs))
	}
}

func TestGetPrefix(t *testing.T) {
	if fmt.Sprintf("%s-v", GetAppName()) != GetPrefix() {
		t.Error("prefix not match")
	}
}

func TestGetFormatedAppName(t *testing.T) {
	// check format "rds-try-v0-0-1"
	strs := strings.Split(GetFormatedAppName(), "-")

	if len(strs) != 5 {
		t.Errorf("app name format not match : len(%d)", len(strs))
	}
}

func TestGetFormatedDBDisplayName(t *testing.T) {
	// check format "rds-try-v0-0-1-2015-01-20-18-03-35-dbIdentifier"
	id := "gotest"
	strs := strings.Split(GetFormatedDBDisplayName(id), "-")

	if len(strs) != 12 {
		t.Errorf("db display name format not match : len(%d)", len(strs))
	}
	if id != strs[len(strs)-1] {
		t.Errorf("db id not match : id(%s)", strs[len(strs)-1])
	}
}

func TestGetFormatedFileDisplayName(t *testing.T) {
	// check format "rds-try-2015-01-20"
	strs := strings.Split(GetFormatedFileDisplayName(), "-")

	if len(strs) != 5 {
		t.Errorf("file display name format not match : len(%d)", len(strs))
	}
}
