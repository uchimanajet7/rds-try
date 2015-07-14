package query

import (
	"errors"
	"path"

	"github.com/BurntSushi/toml"

	"github.com/uchimanajet7/rds-try/logger"
	"github.com/uchimanajet7/rds-try/utils"
)

// Queries struct have Query array variable
type Queries struct {
	Query []Query
}

// Query struct have Name and Sql variable
type Query struct {
	Name string `toml:"name"`
	SQL  string `toml:"sql"`
}

const queryFile = "rds-try.query"

var log = logger.GetLogger("query")

// ErrQueryNotFound is "SQL item not found in query file" error.
var ErrQueryNotFound = errors.New("SQL item not found in query file")

// LoadQuery is the contents are loaded from "rds-try.query" file.
func LoadQuery(file string) (*Queries, error) {
	queries := &Queries{}

	if _, err := toml.DecodeFile(file, &queries); err != nil {
		log.Errorf("%s", err.Error())
		return queries, err
	}

	// check require values
	if len(queries.Query) <= 0 {
		log.Errorf("%s", ErrQueryNotFound.Error())
		return nil, ErrQueryNotFound
	}
	log.Debugf("Queries: %+v", queries)

	return queries, nil
}

// GetDefaultPath is return default query file path.
func GetDefaultPath() string {
	return path.Join(utils.GetHomeDir(), queryFile)
}
