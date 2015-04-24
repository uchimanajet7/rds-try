package command

import (
	"errors"
	"flag"
	"fmt"

	"github.com/awslabs/aws-sdk-go/service/rds"

	"github.com/uchimanajet7/rds-try/query"
	"github.com/uchimanajet7/rds-try/utils"
)

// EsCommand struct is the *Command and OptQuery and OptType and OptSnap variable
type EsCommand struct {
	*Command
	OptQuery string
	OptType  string
	OptSnap  bool
}

// ErrDBInstancetTimeOut is the "DB Instance is time out" error
var ErrDBInstancetTimeOut = errors.New("DB Instance is time out")

// Help is the show help text
func (c *EsCommand) Help() string {
	// to-do: removal of the fixed value
	helpText := fmt.Sprintf("\nUsage: %s es [options]\n\n", utils.GetAppName())
	helpText += "Options:\n"
	helpText += "  -q, --query  specify an alternate query file\n"
	helpText += "  -s, --snap   create snapshot before restore\n"
	helpText += "  -t, --type   specify an alternate db instance class\n"

	return helpText
}

// Synopsis is the show short help text
func (c *EsCommand) Synopsis() string {
	// why ES ?
	// execute and store = ES
	return "restore db and get results by execute sql"
}

// Run is the start command
func (c *EsCommand) Run(args []string) int {
	log.Infof("start command : es")

	// reset flag
	fs := flag.NewFlagSet("es", flag.ExitOnError)

	// register flag name
	fs.StringVar(&c.OptQuery, "query", "", "specify an alternate query file")
	fs.StringVar(&c.OptQuery, "q", "", "specify an alternate query file")
	fs.StringVar(&c.OptType, "type", "", "specify an alternate db instance class")
	fs.StringVar(&c.OptType, "t", "", "specify an alternate db instance class")
	fs.BoolVar(&c.OptSnap, "snap", false, "create snapshot before restore")
	fs.BoolVar(&c.OptSnap, "s", false, "create snapshot before restore")

	fs.Usage = func() { fmt.Println(c.Help()) }
	err := fs.Parse(args)
	if err != nil {
		log.Errorf("%s", err.Error())
		return 1
	}

	err = c.runDetails(fs)
	if err != nil {
		log.Errorf("%s", err.Error())
		return 1
	}

	log.Infof("end command : es")

	return 0
}

func (c *EsCommand) runDetails(f *flag.FlagSet) error {
	// load query
	queryFile := query.GetDefaultPath()
	if c.OptQuery != "" {
		queryFile = c.OptQuery
	}
	queries, err := query.LoadQuery(queryFile)
	if err != nil {
		return err
	}
	log.Debugf("%+v", queries)

	// option create snapshot
	// or
	// get latest db snap shot
	var snapShot *rds.DBSnapshot
	if c.OptSnap {
		snapShot, err = c.CreateDBSnapshot(c.RDSConfig.DBId)
		if err != nil {
			return err
		}

		// wait for available
		waitChan := c.WaitForStatusAvailable(snapShot)
		if !<-waitChan {
			return ErrDBInstancetTimeOut
		}
	} else {
		snapShot, err = c.DescribeLatestDBSnapshot(c.RDSConfig.DBId)
		if err != nil {
			return err
		}
	}

	// get now active db info
	// to-do: can not run if the running instance does not exist
	actDB, err := c.DescribeDBInstance(c.RDSConfig.DBId)
	if err != nil {
		return err
	}

	// "DBInstanceClass" is determined in the following order
	// 1. argument value
	// 2. config file type
	// 3. running DB Instance Class
	restType := *actDB.DBInstanceClass
	if c.RDSConfig.Type != "" {
		restType = c.OptType
	}
	if c.OptType != "" {
		restType = c.OptType
	}
	restName := utils.GetFormatedDBDisplayName(c.RDSConfig.DBId)
	restArgs := &RestoreDBInstanceFromDBSnapshotArgs{
		DBInstanceClass: restType,
		DBIdentifier:    restName,
		MultiAZ:         c.RDSConfig.MultiAz,
		Snapshot:        snapShot,
		Instance:        actDB,
	}
	restDB, err := c.RestoreDBInstanceFromDBSnapshot(restArgs)
	if err != nil {
		return err
	}
	log.Infof("%+v", *restArgs)

	// wait for available
	waitChan := c.WaitForStatusAvailable(restDB)
	if !<-waitChan {
		return ErrDBInstancetTimeOut
	}

	// DB is restored in the default state
	// So, I do modify
	restDB, err = c.ModifyDBInstance(restName, actDB)
	if err != nil {
		return err
	}

	// wait for available
	waitChan = c.WaitForStatusAvailable(restDB)
	if !<-waitChan {
		return ErrDBInstancetTimeOut
	}

	// enable the setting by performing reboot
	restDB, err = c.RebootDBInstance(restName)
	if err != nil {
		return err
	}

	// wait for available
	waitChan = c.WaitForStatusAvailable(restDB)
	if !<-waitChan {
		return ErrDBInstancetTimeOut
	}

	// get db info
	restDB, err = c.DescribeDBInstance(restName)
	if err != nil {
		return err
	}

	// setting check
	var count = 1
	for c.CheckPendingStatus(restDB) {
		// max count
		if count > 6 {
			return ErrDBInstancetTimeOut
		}

		count++
		log.Infof("restart %d times! because change has not been applied", count)

		// once again reboot
		restDB, err = c.RebootDBInstance(restName)
		if err != nil {
			return err
		}

		// wait for available
		waitChan = c.WaitForStatusAvailable(restDB)
		if !<-waitChan {
			return ErrDBInstancetTimeOut
		}

		// get db info
		restDB, err = c.DescribeDBInstance(restName)
		if err != nil {
			return err
		}
	}

	// run queries
	times, err := c.ExecuteSQL(
		&ExecuteSQLArgs{
			Engine:   *restDB.Engine,
			Endpoint: restDB.Endpoint,
			Queries:  queries.Query,
		})
	if err != nil {
		return err
	}

	// show total time
	var total float64
	totalText := "\nruntime result:\n"
	for i, time := range times {
		total += time.Seconds()
		totalText += fmt.Sprintf("  query name   : %s\n  query runtime: %s\n\n", queries.Query[i].Name, time.String())
	}

	hour := int(total) / 3600
	minute := (int(total) - hour*3600) / 60
	second := total - float64(hour*3600) - float64(minute*60)

	totalText += "--------------------------------\n"
	timeText := fmt.Sprintf("  total runtime: %.3f sec\n", second)
	if minute > 0 {
		timeText = fmt.Sprintf("  total runtime: %d m %.3f sec\n", minute, second)
	}
	if hour > 0 {
		timeText = fmt.Sprintf("  total runtime: %d h %d m %.3f sec\n", hour, minute, second)
	}
	totalText += timeText
	fmt.Println(totalText)

	return nil
}
