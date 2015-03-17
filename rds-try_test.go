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

	test_name := utils.GetAppName() + "-test"
	temp_dir, _ := ioutil.TempDir("", test_name)
	out := config.OutConfig{
		Root: temp_dir,
		File: true,
		Bom:  true,
	}

	log := config.LogConfig{
		Root:    temp_dir,
		Verbose: true,
		Json:    true,
	}

	rds := config.RDSConfig{
		MultiAz: true,
		DBId:    utils.GetFormatedDBDisplayName(test_name),
		Region:  "us-west-2",
		User:    "test-admin",
		Pass:    "pass-pass",
		Type:    "db.m3.medium",
	}
	rds_map := map[string]config.RDSConfig{
		"default2": rds,
	}

	config := &config.Config{
		Aws: aws,
		Out: out,
		Rds: rds_map,
		Log: log,
	}
	temp_file, err := ioutil.TempFile(temp_dir, utils.GetAppName()+"-test")
	if err != nil {
		return nil, temp_dir, err
	}
	if err := toml.NewEncoder(temp_file).Encode(config); err != nil {
		return nil, temp_dir, err
	}
	temp_file.Sync()
	temp_file.Close()

	os.Args = []string{utils.GetAppName(), "-c=" + temp_file.Name(), "-n=default", "ls"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	return config, temp_dir, nil
}

func TestResolveArgs(t *testing.T) {
	test_conf, temp_dir, err := getTestConfig()
	defer os.RemoveAll(temp_dir)
	if err != nil {
		t.Errorf("config file for the test was not created: %s", err.Error())
	}

	// resolve command line
	conf, ex_code := resolveArgs()
	if ex_code != 0 || conf == nil {
		t.Errorf("config file load error")
	}
	if !reflect.DeepEqual(test_conf, conf) {
		t.Errorf("config data not match: %+v/%+v", test_conf, conf)
	}
}

func TestSetLogOptions(t *testing.T) {
	test_conf, temp_dir, err := getTestConfig()
	defer os.RemoveAll(temp_dir)
	if err != nil {
		t.Errorf("config file for the test was not created: %s", err.Error())
	}

	// resolve command line
	conf, ex_code := resolveArgs()
	if ex_code != 0 || conf == nil {
		t.Errorf("config file load error")
	}
	if !reflect.DeepEqual(test_conf, conf) {
		t.Errorf("config data not match: %+v/%+v", test_conf, conf)
	}

	// log setting
	log_file, ex_code := setLogOptions(conf)
	defer log_file.Close()
	if ex_code != 0 || log_file == nil {
		t.Errorf("log setting error")
	}
}

func TestGetCommandStruct(t *testing.T) {
	test_conf, temp_dir, err := getTestConfig()
	defer os.RemoveAll(temp_dir)
	if err != nil {
		t.Errorf("config file for the test was not created: %s", err.Error())
	}

	// resolve command line
	conf, ex_code := resolveArgs()
	if ex_code != 0 || conf == nil {
		t.Errorf("config file load error")
	}
	if !reflect.DeepEqual(test_conf, conf) {
		t.Errorf("config data not match: %+v/%+v", test_conf, conf)
	}

	// get base command
	// Known error message:[The security token included in the request is invalid.]
	t.Log("Known error message:[The security token included in the request is invalid.]")
	fmt.Println(" ### Known error message:[The security token included in the request is invalid.]")
	cmd_st, ex_code := getCommandStruct(conf)
	if ex_code != 0 || cmd_st == nil {
		return
	}
	if !reflect.DeepEqual(cmd_st.RDSConfig, conf.Rds) {
		t.Errorf("RDS config data not match: %+v/%+v", cmd_st.RDSConfig, conf.Rds)
	}
}
