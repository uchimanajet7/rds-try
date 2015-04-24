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

	tempFile, err := ioutil.TempFile("", utils.GetAppName()+"-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}
	if err := toml.NewEncoder(tempFile).Encode(queries); err != nil {
		t.Errorf("failed to create the toml file: %s", err.Error())
	}
	tempFile.Sync()
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	query, err := LoadQuery(tempFile.Name())

	if err != nil {
		t.Errorf("query file load error: %s", err.Error())
	}
	if !reflect.DeepEqual(queries, query) {
		t.Errorf("query data not match: %+v/%+v", queries, query)
	}
}

func TestGetDefaultPath(t *testing.T) {
	if path.Join(utils.GetHomeDir(), queryFile) != GetDefaultPath() {
		t.Error("default path not match")
	}
}
