package config

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/uchimanajet7/rds-try/utils"
)

func TestLoadConfig(t *testing.T) {
	aws := AWSConfig{
		Accesskey: "CF2DC307CC89F49F68F365235AA54BAB2DDD02DA",
		SecretKey: "ca485e5b709eae0ceacd68b61cdb28119f13942c71bbb68066c8a3cb45185a39",
	}

	testName := utils.GetAppName() + "-test"
	tempDir, _ := ioutil.TempDir("", testName)
	out := OutConfig{
		Root: tempDir,
		File: true,
		Bom:  true,
	}

	log := LogConfig{
		Root:    tempDir,
		Verbose: true,
		JSON:    true,
	}

	rds := RDSConfig{
		MultiAz: true,
		DBId:    utils.GetFormatedDBDisplayName(testName),
		Region:  "us-west-2",
		User:    "test-admin",
		Pass:    "pass-pass",
		Type:    "db.m3.medium",
	}
	rdsMap := map[string]RDSConfig{
		"default": rds,
	}

	config := &Config{
		Aws: aws,
		Out: out,
		Rds: rdsMap,
		Log: log,
	}
	tempFile, err := ioutil.TempFile(tempDir, utils.GetAppName()+"-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}
	if err := toml.NewEncoder(tempFile).Encode(config); err != nil {
		t.Errorf("failed to create the toml file: %s", err.Error())
	}
	tempFile.Sync()
	tempFile.Close()
	defer os.RemoveAll(tempDir)

	conf, err := LoadConfig(tempFile.Name())
	if err != nil {
		t.Errorf("config file load error: %s", err.Error())
	}
	if !reflect.DeepEqual(config, conf) {
		t.Errorf("config data not match: %+v/%+v", config, conf)
	}
}

func TestGetDefaultPath(t *testing.T) {
	if path.Join(utils.GetHomeDir(), configFile) != GetDefaultPath() {
		t.Error("default path not match")
	}
}
