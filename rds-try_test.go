package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/uchimanajet7/rds-try/config"
	"github.com/uchimanajet7/rds-try/utils"
)

// need to run the caller always "defer os.RemoveAll(temp_dir)"
func getTestConfig() (*config.Config, string, error) {
	aws := config.AWSConfig{
		Accesskey: "CF2DC307CC89F49F68F365235AA54BAB2DDD02DA",
		SecretKey: "ca485e5b709eae0ceacd68b61cdb28119f13942c71bbb68066c8a3cb45185a39",
	}

	testName := utils.GetAppName() + "-test"
	tempDir, _ := ioutil.TempDir("", testName)
	out := config.OutConfig{
		Root: tempDir,
		File: true,
		Bom:  true,
	}

	log := config.LogConfig{
		Root:    tempDir,
		Verbose: true,
		JSON:    true,
	}

	rds := config.RDSConfig{
		MultiAz: true,
		DBId:    utils.GetFormatedDBDisplayName(testName),
		Region:  "us-west-2",
		User:    "test-admin",
		Pass:    "pass-pass",
		Type:    "db.m3.medium",
	}
	rdsMap := map[string]config.RDSConfig{
		"default2": rds,
	}

	config := &config.Config{
		Aws: aws,
		Out: out,
		Rds: rdsMap,
		Log: log,
	}
	tempFile, err := ioutil.TempFile(tempDir, utils.GetAppName()+"-test")
	if err != nil {
		return nil, tempDir, err
	}
	if err := toml.NewEncoder(tempFile).Encode(config); err != nil {
		return nil, tempDir, err
	}
	tempFile.Sync()
	tempFile.Close()

	os.Args = []string{utils.GetAppName(), "-c=" + tempFile.Name(), "-n=default", "ls"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	return config, tempDir, nil
}

func TestResolveArgs(t *testing.T) {
	testConfig, tempDir, err := getTestConfig()
	defer os.RemoveAll(tempDir)
	if err != nil {
		t.Errorf("config file for the test was not created: %s", err.Error())
	}

	// resolve command line
	conf, exCode := resolveArgs()
	if exCode != 0 || conf == nil {
		t.Errorf("config file load error")
	}
	if !reflect.DeepEqual(testConfig, conf) {
		t.Errorf("config data not match: %+v/%+v", testConfig, conf)
	}
}

func TestSetLogOptions(t *testing.T) {
	testConfig, tempDir, err := getTestConfig()
	defer os.RemoveAll(tempDir)
	if err != nil {
		t.Errorf("config file for the test was not created: %s", err.Error())
	}

	// resolve command line
	conf, exCode := resolveArgs()
	if exCode != 0 || conf == nil {
		t.Errorf("config file load error")
	}
	if !reflect.DeepEqual(testConfig, conf) {
		t.Errorf("config data not match: %+v/%+v", testConfig, conf)
	}

	// log setting
	logFile, exCode := setLogOptions(conf)
	defer logFile.Close()
	if exCode != 0 || logFile == nil {
		t.Errorf("log setting error")
	}
}

func TestGetCommandStruct(t *testing.T) {
	testConfig, tempDir, err := getTestConfig()
	defer os.RemoveAll(tempDir)
	if err != nil {
		t.Errorf("config file for the test was not created: %s", err.Error())
	}

	// resolve command line
	conf, exCode := resolveArgs()
	if exCode != 0 || conf == nil {
		t.Errorf("config file load error")
	}
	if !reflect.DeepEqual(testConfig, conf) {
		t.Errorf("config data not match: %+v/%+v", testConfig, conf)
	}

	// get base command
	// Known error message:[The security token included in the request is invalid.]
	t.Log("Known error message:[The security token included in the request is invalid.]")
	fmt.Println(" ### Known error message:[The security token included in the request is invalid.]")
	commandStruct, exCode := getCommandStruct(conf)
	if exCode != 0 || commandStruct == nil {
		return
	}
	if !reflect.DeepEqual(commandStruct.RDSConfig, conf.Rds) {
		t.Errorf("RDS config data not match: %+v/%+v", commandStruct.RDSConfig, conf.Rds)
	}
}
