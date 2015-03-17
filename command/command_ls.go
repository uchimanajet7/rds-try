package command

import (
	"flag"
	"fmt"

	"github.com/uchimanajet7/rds-try/utils"
)

type LsCommand struct {
	*Command
	OptSnap bool
}

func (c *LsCommand) Help() string {
	// to-do: removal of the fixed value
	help_text := fmt.Sprintf("\nUsage: %s ls [options]\n\n", utils.GetAppName())
	help_text += "Options:\n"
	help_text += "  -s, --snap  include own db snapshots to list\n"

	return help_text
}

func (c *LsCommand) Synopsis() string {
	return "list up own db instances and snapshots"
}

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
	db_list, err := c.DescribeDBInstancesByTags()
	if err != nil {
		return err
	}

	// show db list
	if len(db_list) <= 0 {
		fmt.Printf("\ndb instance list not exist\n")
	} else {
		fmt.Printf("\nlist of own db instance\n")
		for i, db := range db_list {
			fmt.Printf("  [% d] DB Instance: %s\n", i+1, *db.DBInstanceIdentifier)
		}
	}
	// blank new line
	fmt.Println("")

	if c.OptSnap {
		// to get list created in this tool
		snap_list, err := c.DescribeDBSnapshotsByTags()
		if err != nil {
			return err
		}

		// show snapshot list
		if len(snap_list) <= 0 {
			fmt.Printf("db snapshot list not exist\n")
		} else {
			fmt.Printf("list of own db snapshot\n")
			for i, snap := range snap_list {
				fmt.Printf("  [% d] DB Snapshot: %s\n", i+1, *snap.DBSnapshotIdentifier)
			}
		}
		// blank new line
		fmt.Println("")
	}

	return nil
}
