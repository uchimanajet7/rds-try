package command

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/awslabs/aws-sdk-go/service/rds"

	"github.com/uchimanajet7/rds-try/utils"
)

type RmCommand struct {
	*Command
	OptSnap  bool
	OptForce bool
	OptItem  string
}

var ErrInterruptedAskDelete = errors.New("OS Interrupted Ask Delete")

func (c *RmCommand) Help() string {
	// to-do: removal of the fixed value
	help_text := fmt.Sprintf("\nUsage: %s rm [options]\n\n", utils.GetAppName())
	help_text += "Options:\n"
	help_text += "  -s, --snap   list up own db snapshots\n"
	help_text += "  -f, --force  forced delete without confirmation\n"

	return help_text
}

func (c *RmCommand) Synopsis() string {
	return "delete your created db instances and snapshots"
}

func (c *RmCommand) Run(args []string) int {
	log.Infof("start command : rm")

	// reset flag
	fs := flag.NewFlagSet("rm", flag.ExitOnError)

	// register flag name
	fs.BoolVar(&c.OptSnap, "snap", false, "include own db snapshots to delete")
	fs.BoolVar(&c.OptSnap, "s", false, "include own db snapshots to delete")
	fs.BoolVar(&c.OptForce, "force", false, "forced delete without confirmation")
	fs.BoolVar(&c.OptForce, "f", false, "forced delete without confirmation")

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
	log.Infof("end command : rm")

	return 0
}

func (c *RmCommand) runDetails(f *flag.FlagSet) error {
	// to get list created in this tool
	db_list, err := c.DescribeDBInstancesByTags()
	if err != nil {
		return err
	}
	ask_flg := true

	// show db list
	if len(db_list) <= 0 {
		ask_flg = false
		fmt.Printf("\ndb instance list not exist\n")
	} else {
		fmt.Printf("\nlist of own db instance\n")
		for i, db := range db_list {
			fmt.Printf("  [% d] DB Instance: %s\n", i+1, *db.DBInstanceIdentifier)
		}
	}
	// blank new line
	fmt.Println("")

	var snap_list []*rds.DBSnapshot
	if c.OptSnap {
		// to get list created in this tool
		snap_list, err = c.DescribeDBSnapshotsByTags()
		if err != nil {
			return err
		}

		// show snapshot list
		if len(snap_list) <= 0 {
			ask_flg = false
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

	// list does not exist
	if !ask_flg {
		return nil
	}

	// confirm delete
	var ask_resp string
	if c.OptForce {
		ask_resp = "yes"
	} else {
		ask_resp, err = askQuestion("you want to delete all of those? [y/n]:")
		if err != nil {
			log.Errorf("%s", err.Error())
			return err
		}
	}
	// blank new line
	fmt.Println("")

	switch ask_resp {
	case "y", "Y", "yes", "YES", "Yes":
		// delete db instance
		err = c.DeleteDBResources(db_list)
		if err != nil {
			return err
		}
		// delete db snapshot
		if c.OptSnap {
			err = c.DeleteDBResources(snap_list)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// this method copied
// see also
// https://github.com/mitchellh/cli
func askQuestion(query string) (string, error) {
	// show query string
	fmt.Printf("%s ", query)

	// Register for interrupts so that we can catch it and immediately
	// return...
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	// Ask for input in a go-routine so that we can ignore it.
	errCh := make(chan error, 1)
	lineCh := make(chan string, 1)
	go func() {
		r := bufio.NewReader(os.Stdin)
		line, err := r.ReadString('\n')
		if err != nil {
			errCh <- err
			return
		}

		lineCh <- strings.TrimRight(line, "\r\n")
	}()

	select {
	case err := <-errCh:
		return "", err
	case line := <-lineCh:
		return line, nil
	case <-sigCh:
		// Print a newline so that any further output starts properly
		// on a new line.
		fmt.Println("")

		return "", ErrInterruptedAskDelete
	}
}
