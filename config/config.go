package config

import (
	"errors"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/awslabs/aws-sdk-go/aws"

	"github.com/uchimanajet7/rds-try/logger"
	"github.com/uchimanajet7/rds-try/utils"
)

type Config struct {
	Aws AWSConfig
	Out OutConfig
	Rds map[string]RDSConfig
	Log LogConfig
}

type AWSConfig struct {
	Accesskey string `toml:"access_key"`
	SecretKey string `toml:"secret_key"`
}

type OutConfig struct {
	Root string `toml:"root"`
	File bool   `toml:"file"`
	Bom  bool   `toml:"bom"`
}

type LogConfig struct {
	Root    string `toml:"root"`
	Verbose bool   `toml:"verbose"`
	Json    bool   `toml:"json"`
}

type RDSConfig struct {
	MultiAz bool   `toml:"multi_az"`
	DBId    string `toml:"db_id"`
	Region  string `toml:"region"`
	User    string `toml:"user"`
	Pass    string `toml:"pass"`
	Type    string `toml:"type"`
}

const config_file = "rds-try.conf"

var log = logger.GetLogger("config")

var ErrRdsSectionNotFound = errors.New("[rds.*] section not found in file")

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

func GetDefaultPath() string {
	return path.Join(utils.GetHomeDir(), config_file)
}

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
