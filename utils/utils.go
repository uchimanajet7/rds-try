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

const app_name = "rds-try"
const app_version = "v0.0.1"

var log = logger.GetLogger("utils")

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

// format "rds-try"
func GetAppName() string {
	return app_name
}

// format "v0.0.1"
func GetAppVersion() string {
	return app_version
}

// format "2015-01-20-18-03-35"
func GetFormatedTime() string {
	return time.Now().Format("2006-01-02-15-04-05")
}

// format "rds-try-v"
func GetPrefix() string {
	return fmt.Sprintf("%s-v", GetAppName())
}

// format "rds-try-v0-0-1"
func GetFormatedAppName() string {
	ver_name := strings.Replace(GetAppVersion(), ".", "-", -1)
	return fmt.Sprintf("%s-%s", GetAppName(), ver_name)
}

// format "rds-try-v0-0-1-2015-01-20-18-03-35-dbIdentifier"
func GetFormatedDBDisplayName(dbIdentifier string) string {
	return fmt.Sprintf("%s-%s-%s", GetFormatedAppName(), GetFormatedTime(), dbIdentifier)
}

// format "rds-try-2015-01-20"
func GetFormatedFileDisplayName() string {
	return fmt.Sprintf("%s-%s", GetAppName(), time.Now().Format("2006-01-02"))
}
