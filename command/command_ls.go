package command

import (
	"flag"
	"fmt"

	"github.com/uchimanajet7/rds-try/utils"
)

// LsCommand struct is the *Command and OptSnap variable
type LsCommand struct {
	*Command
	OptSnap bool
}

// Help is the show help text
func (c *LsCommand) Help() string {
	// to-do: removal of the fixed value
	helpText := fmt.Sprintf("\nUsage: %s ls [options]\n\n", utils.GetAppName())
	helpText += "Options:\n"
	helpText += "  -s, --snap  include own db snapshots to list\n"

	return helpText
}

// Synopsis is the show short help text
func (c *LsCommand) Synopsis() string {
	return "list up own db instances and snapshots"
}

// Run is the start command
func (c *LsCommand) Run(args []string) int {
	log.Infof("start command : ls")

	// reset flag
	fs := flag.NewFlagSet("ls", flag.ExitOnError)

	// register flag name
	fs.BoolVar(&c.OptSnap, "snap", false, "include own db snapshots to list")
	fs.BoolVar(&c.OptSnap, "s", false, "include own db snapshots to list")

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

	log.Infof("end command : ls")

	return 0
}

func (c *LsCommand) runDetails(f *flag.FlagSet) error {
	// to get list created in this tool
	dbList, err := c.DescribeDBInstancesByTags()
	if err != nil {
		return err
	}

	// show db list
	if len(dbList) <= 0 {
		fmt.Printf("\ndb instance list not exist\n")
	} else {
		fmt.Printf("\nlist of own db instance\n")
		for i, db := range dbList {
			fmt.Printf("  [% d] DB Instance: %s\n", i+1, *db.DBInstanceIdentifier)
		}
	}
	// blank new line
	fmt.Println("")

	if c.OptSnap {
		// to get list created in this tool
		snapList, err := c.DescribeDBSnapshotsByTags()
		if err != nil {
			return err
		}

		// show snapshot list
		if len(snapList) <= 0 {
			fmt.Printf("db snapshot list not exist\n")
		} else {
			fmt.Printf("list of own db snapshot\n")
			for i, snap := range snapList {
				fmt.Printf("  [% d] DB Snapshot: %s\n", i+1, *snap.DBSnapshotIdentifier)
			}
		}
		// blank new line
		fmt.Println("")
	}

	return nil
}
