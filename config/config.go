package config

import (
	"errors"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/awslabs/aws-sdk-go/aws"

	"github.com/uchimanajet7/rds-try/logger"
	"github.com/uchimanajet7/rds-try/utils"
)

// Config struct is Aws AWSConfig and Out OutConfig and Rds map and Log LogConfig variable
type Config struct {
	Aws AWSConfig
	Out OutConfig
	Rds map[string]RDSConfig
	Log LogConfig
}

// AWSConfig struct is Accesskey and SecretKey variable
type AWSConfig struct {
	Accesskey string `toml:"access_key"`
	SecretKey string `toml:"secret_key"`
}

// OutConfig struct is Root and File and Bom variable
type OutConfig struct {
	Root string `toml:"root"`
	File bool   `toml:"file"`
	Bom  bool   `toml:"bom"`
}

// LogConfig struct is Root and Verbose and JSON variable
type LogConfig struct {
	Root    string `toml:"root"`
	Verbose bool   `toml:"verbose"`
	JSON    bool   `toml:"json"`
}

// RDSConfig struct is MultiAz and DBId and Region and User and Pass and Type variable
type RDSConfig struct {
	MultiAz bool   `toml:"multi_az"`
	DBId    string `toml:"db_id"`
	Region  string `toml:"region"`
	User    string `toml:"user"`
	Pass    string `toml:"pass"`
	Type    string `toml:"type"`
}

const configFile = "rds-try.conf"

var log = logger.GetLogger("config")

// ErrRdsSectionNotFound is the config file format error
var ErrRdsSectionNotFound = errors.New("[rds.*] section not found in file")

// LoadConfig is the contents are loaded from "rds-try.conf" file.
func LoadConfig(file string) (*Config, error) {
	config := &Config{}

	if _, err := toml.DecodeFile(file, &config); err != nil {
		log.Errorf("%s", err.Error())
		return config, err
	}

	if len(config.Rds) <= 0 {
		log.Errorf("%s", ErrRdsSectionNotFound.Error())
		return nil, ErrRdsSectionNotFound
	}
	log.Debugf("Config: %+v", config)

	return config, nil
}

// GetDefaultPath is return default conf file path.
func GetDefaultPath() string {
	return path.Join(utils.GetHomeDir(), configFile)
}

// GetAWSCreds is the call aws-sdk-go "DetectCreds"
//
// evaluated in the following order
// 1. input variable
// 2. Environment variable
// 3. /.aws/credentials
// 4. IAM Role
// see also
// aws - GoDoc
// http://godoc.org/github.com/awslabs/aws-sdk-go/aws#DetectCreds
func (c *Config) GetAWSCreds() aws.CredentialsProvider {
	return aws.DetectCreds(c.Aws.Accesskey, c.Aws.SecretKey, "")
}
