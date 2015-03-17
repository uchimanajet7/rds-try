package query

import (
	"errors"
	"path"

	"github.com/BurntSushi/toml"

	"github.com/uchimanajet7/rds-try/logger"
	"github.com/uchimanajet7/rds-try/utils"
)

type Queries struct {
	Query []Query
}

type Query struct {
	Name string `toml:"name"`
	Sql  string `toml:"sql"`
}

const query_file = "rds-try.query"

var log = logger.GetLogger("query")

var ErrQueryNotFound = errors.New("SQL item not found in query file")

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

func GetDefaultPath() string {
	return path.Join(utils.GetHomeDir(), query_file)
}
