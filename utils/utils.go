package utils

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/uchimanajet7/rds-try/logger"
)

const appName = "rds-try"
const appVersion = "v0.0.3"

var log = logger.GetLogger("utils")

// GetHomeDir is the function to acquire the user's home directory.
func GetHomeDir() string {
	// OS under execution is Windows
	if runtime.GOOS == "windows" {
		u, err := user.Current()
		if err != nil {
			log.Errorf("%s", err.Error())
			return ""
		}

		return u.HomeDir
	}

	// OS under execution is others
	home := os.Getenv("HOME")
	if home == "" {
		log.Errorf("Environment variable HOME not set")
		return ""
	}
	log.Debugf("home dir: %s", home)

	return home
}

// GetUserName is the function to acquire the user name.
func GetUserName() string {
	var user = ""

	if runtime.GOOS == "windows" {
		// OS under execution is Windows
		user = os.Getenv("USERNAME")
	} else {
		// OS under execution is others
		user = os.Getenv("USER")
	}

	if user == "" {
		log.Errorf("Environment variable USER or USERNAME not set")
		return ""
	}
	log.Debugf("os user: %s", user)

	return user
}

// GetAppName is the function to acquire the application name.
// return format: "rds-try"
func GetAppName() string {
	return appName
}

// GetAppVersion is the function to acquire the application version.
// return format: "v0.0.1"
func GetAppVersion() string {
	return appVersion
}

// GetFormatedTime is the function to acquire the edited "now date time".
// return format: "2015-01-20-18-03-35"
func GetFormatedTime() string {
	return time.Now().Format("2006-01-02-15-04-05")
}

// GetPrefix is the function to acquire the application name prefix.
// return format: "rds-try-v"
func GetPrefix() string {
	return fmt.Sprintf("%s-v", GetAppName())
}

// GetFormatedAppName is the function to acquire the edited application name and version.
// return format: "rds-try-v0-0-1"
func GetFormatedAppName() string {
	versionText := strings.Replace(GetAppVersion(), ".", "-", -1)
	return fmt.Sprintf("%s-%s", GetAppName(), versionText)
}

// GetFormatedDBDisplayName is the function to acquire the edited application name and version and "now date time".
// return format: "rds-try-v0-0-1-2015-01-20-18-03-35-dbIdentifier"
func GetFormatedDBDisplayName(dbIdentifier string) string {
	return fmt.Sprintf("%s-%s-%s", GetFormatedAppName(), GetFormatedTime(), dbIdentifier)
}

// GetFormatedFileDisplayName is the function to acquire the edited application name and "now date time".
// return format: "rds-try-2015-01-20"
func GetFormatedFileDisplayName() string {
	return fmt.Sprintf("%s-%s", GetAppName(), time.Now().Format("2006-01-02"))
}
