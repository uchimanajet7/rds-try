package query

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/uchimanajet7/rds-try/utils"
)

func TestLoadQuery(t *testing.T) {
	queries := &Queries{}

	for i := 0; i < 10; i++ {
		query := Query{
			Name: fmt.Sprintf("name_%d", i+1),
			Sql:  fmt.Sprintf("sql_%d", i+1),
		}
		queries.Query = append(queries.Query, query)
	}

	temp_file, err := ioutil.TempFile("", utils.GetAppName()+"-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}
	if err := toml.NewEncoder(temp_file).Encode(queries); err != nil {
		t.Errorf("failed to create the toml file: %s", err.Error())
	}
	temp_file.Sync()
	temp_file.Close()
	defer os.Remove(temp_file.Name())

	query, err := LoadQuery(temp_file.Name())

	if err != nil {
		t.Errorf("query file load error: %s", err.Error())
	}
	if !reflect.DeepEqual(queries, query) {
		t.Errorf("query data not match: %+v/%+v", queries, query)
	}
}

func TestGetDefaultPath(t *testing.T) {
	if path.Join(utils.GetHomeDir(), query_file) != GetDefaultPath() {
		t.Error("default path not match")
	}
}
