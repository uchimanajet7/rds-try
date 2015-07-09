package command

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/uchimanajet7/rds-try/utils"
)

// RmCommand struct is the *Command and OptSnap and OptForce and OptItem variable
type RmCommand struct {
	*Command
	OptSnap  bool
	OptForce bool
	OptItem  string
}

// ErrInterruptedAskDelete is the "OS Interrupted Ask Delete" error
var ErrInterruptedAskDelete = errors.New("OS Interrupted Ask Delete")

// Help is the show help text
func (c *RmCommand) Help() string {
	// to-do: removal of the fixed value
	helpText := fmt.Sprintf("\nUsage: %s rm [options]\n\n", utils.GetAppName())
	helpText += "Options:\n"
	helpText += "  -s, --snap   list up own db snapshots\n"
	helpText += "  -f, --force  forced delete without confirmation\n"

	return helpText
}

// Synopsis is the show short help text
func (c *RmCommand) Synopsis() string {
	return "delete your created db instances and snapshots"
}

// Run is the start command
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
	dbList, err := c.DescribeDBInstancesByTags()
	if err != nil {
		return err
	}
	askCount := 0

	// show db list
	if len(dbList) <= 0 {
		fmt.Printf("\ndb instance list not exist\n")
	} else {
		askCount++
		fmt.Printf("\nlist of own db instance\n")
		for i, db := range dbList {
			fmt.Printf("  [% d] DB Instance: %s\n", i+1, *db.DBInstanceIdentifier)
		}
	}
	// blank new line
	fmt.Println("")

	var snapList []*rds.DBSnapshot
	if c.OptSnap {
		// to get list created in this tool
		snapList, err = c.DescribeDBSnapshotsByTags()
		if err != nil {
			return err
		}

		// show snapshot list
		if len(snapList) <= 0 {
			fmt.Printf("db snapshot list not exist\n")
		} else {
			askCount++
			fmt.Printf("list of own db snapshot\n")
			for i, snap := range snapList {
				fmt.Printf("  [% d] DB Snapshot: %s\n", i+1, *snap.DBSnapshotIdentifier)
			}
		}
		// blank new line
		fmt.Println("")
	}

	// list does not exist
	if askCount <= 0 {
		return nil
	}

	// confirm delete
	var askResp string
	if c.OptForce {
		askResp = "yes"
	} else {
		askResp, err = askQuestion("you want to delete all of those? [y/n]:")
		if err != nil {
			log.Errorf("%s", err.Error())
			return err
		}
	}
	// blank new line
	fmt.Println("")

	switch askResp {
	case "y", "Y", "yes", "YES", "Yes":
		// delete db instance
		err = c.DeleteDBResources(dbList)
		if err != nil {
			return err
		}
		// delete db snapshot
		if c.OptSnap {
			err = c.DeleteDBResources(snapList)
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
