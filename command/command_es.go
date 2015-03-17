package command

import (
	"errors"
	"flag"
	"fmt"

	"github.com/awslabs/aws-sdk-go/gen/rds"

	"github.com/uchimanajet7/rds-try/query"
	"github.com/uchimanajet7/rds-try/utils"
)

var ErrDBInstancetTimeOut = errors.New("DB Instance is time out")

type EsCommand struct {
	*Command
	OptQuery string
	OptType  string
	OptSnap  bool
}

func (c *EsCommand) Help() string {
	// to-do: removal of the fixed value
	help_text := fmt.Sprintf("\nUsage: %s es [options]\n\n", utils.GetAppName())
	help_text += "Options:\n"
	help_text += "  -q, --query  specify an alternate query file\n"
	help_text += "  -s, --snap   create snapshot before restore\n"
	help_text += "  -t, --type   specify an alternate db instance class\n"

	return help_text
}

func (c *EsCommand) Synopsis() string {
	// why ES ?
	// execute and store = ES
	return "restore db and get results by execute sql"
}

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
	query_file := query.GetDefaultPath()
	if c.OptQuery != "" {
		query_file = c.OptQuery
	}
	queries, err := query.LoadQuery(query_file)
	if err != nil {
		return err
	}
	log.Debugf("%+v", queries)

	// option create snapshot
	// or
	// get latest db snap shot
	var snap_shot *rds.DBSnapshot
	if c.OptSnap {
		snap_shot, err = c.CreateDBSnapshot(c.RDSConfig.DBId)
		if err != nil {
			return err
		}

		// wait for available
		wait_chan := c.WaitForStatusAvailable(snap_shot)
		if !<-wait_chan {
			return ErrDBInstancetTimeOut
		}
	} else {
		snap_shot, err = c.DescribeLatestDBSnapshot(c.RDSConfig.DBId)
		if err != nil {
			return err
		}
	}

	// get now active db info
	// to-do: can not run if the running instance does not exist
	act_db, err := c.DescribeDBInstance(c.RDSConfig.DBId)
	if err != nil {
		return err
	}

	// "DBInstanceClass" is determined in the following order
	// 1. argument value
	// 2. config file type
	// 3. running DB Instance Class
	rest_type := *act_db.DBInstanceClass
	if c.RDSConfig.Type != "" {
		rest_type = c.OptType
	}
	if c.OptType != "" {
		rest_type = c.OptType
	}
	rest_name := utils.GetFormatedDBDisplayName(c.RDSConfig.DBId)
	rest_args := &RestoreDBInstanceFromDBSnapshotArgs{
		DBInstanceClass: rest_type,
		DBIdentifier:    rest_name,
		MultiAZ:         c.RDSConfig.MultiAz,
		Snapshot:        snap_shot,
		Instance:        act_db,
	}
	rest_db, err := c.RestoreDBInstanceFromDBSnapshot(rest_args)
	if err != nil {
		return err
	}
	log.Infof("%+v", *rest_args)

	// wait for available
	wait_chan := c.WaitForStatusAvailable(rest_db)
	if !<-wait_chan {
		return ErrDBInstancetTimeOut
	}

	// DB is restored in the default state
	// So, I do modify
	rest_db, err = c.ModifyDBInstance(rest_name, act_db)
	if err != nil {
		return err
	}

	// wait for available
	wait_chan = c.WaitForStatusAvailable(rest_db)
	if !<-wait_chan {
		return ErrDBInstancetTimeOut
	}

	// enable the setting by performing reboot
	rest_db, err = c.RebootDBInstance(rest_name)
	if err != nil {
		return err
	}

	// wait for available
	wait_chan = c.WaitForStatusAvailable(rest_db)
	if !<-wait_chan {
		return ErrDBInstancetTimeOut
	}

	// get db info
	rest_db, err = c.DescribeDBInstance(rest_name)
	if err != nil {
		return err
	}

	// setting check
	if c.CheckPendingStatus(rest_db) {
		log.Infof("restart second time! because change has not been applied")

		// once again reboot
		rest_db, err = c.RebootDBInstance(rest_name)
		if err != nil {
			return err
		}

		// wait for available
		wait_chan = c.WaitForStatusAvailable(rest_db)
		if !<-wait_chan {
			return ErrDBInstancetTimeOut
		}
	}

	// run queries
	times, err := c.ExecuteSQL(
		&ExecuteSQLArgs{
			Engine:   *rest_db.Engine,
			Endpoint: rest_db.Endpoint,
			Queries:  queries.Query,
		})
	if err != nil {
		return err
	}

	// show total time
	var total float64
	total_text := "\nruntime result:\n"
	for i, time := range times {
		total += time.Seconds()
		total_text += fmt.Sprintf("  query name   : %s\n  query runtime: %s\n\n", queries.Query[i].Name, time.String())
	}

	hour := int(total) / 3600
	minute := (int(total) - hour*3600) / 60
	second := total - float64(hour*3600) - float64(minute*60)

	total_text += "--------------------------------\n"
	time_text := fmt.Sprintf("  total runtime: %.3f sec\n", second)
	if minute > 0 {
		time_text = fmt.Sprintf("  total runtime: %d m %.3f sec\n", minute, second)
	}
	if hour > 0 {
		time_text = fmt.Sprintf("  total runtime: %d h %d m %.3f sec\n", hour, minute, second)
	}
	total_text += time_text
	fmt.Println(total_text)

	return nil
}
