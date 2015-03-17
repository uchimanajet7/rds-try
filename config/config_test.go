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

	test_name := utils.GetAppName() + "-test"
	temp_dir, _ := ioutil.TempDir("", test_name)
	out := OutConfig{
		Root: temp_dir,
		File: true,
		Bom:  true,
	}

	log := LogConfig{
		Root:    temp_dir,
		Verbose: true,
		Json:    true,
	}

	rds := RDSConfig{
		MultiAz: true,
		DBId:    utils.GetFormatedDBDisplayName(test_name),
		Region:  "us-west-2",
		User:    "test-admin",
		Pass:    "pass-pass",
		Type:    "db.m3.medium",
	}
	rds_map := map[string]RDSConfig{
		"default": rds,
	}

	config := &Config{
		Aws: aws,
		Out: out,
		Rds: rds_map,
		Log: log,
	}
	temp_file, err := ioutil.TempFile(temp_dir, utils.GetAppName()+"-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}
	if err := toml.NewEncoder(temp_file).Encode(config); err != nil {
		t.Errorf("failed to create the toml file: %s", err.Error())
	}
	temp_file.Sync()
	temp_file.Close()
	defer os.RemoveAll(temp_dir)

	conf, err := LoadConfig(temp_file.Name())
	if err != nil {
		t.Errorf("config file load error: %s", err.Error())
	}
	if !reflect.DeepEqual(config, conf) {
		t.Errorf("config data not match: %+v/%+v", config, conf)
	}
}

func TestGetDefaultPath(t *testing.T) {
	if path.Join(utils.GetHomeDir(), config_file) != GetDefaultPath() {
		t.Error("default path not match")
	}
}
