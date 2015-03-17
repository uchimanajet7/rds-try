package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"

	"github.com/awslabs/aws-sdk-go/gen/iam"
	"github.com/awslabs/aws-sdk-go/gen/rds"

	"github.com/uchimanajet7/rds-try/command"
	"github.com/uchimanajet7/rds-try/config"
	"github.com/uchimanajet7/rds-try/logger"
	"github.com/uchimanajet7/rds-try/utils"
)

var log = logger.GetLogger("main")

func showHelp() func() {
	return func() {
		help_text := fmt.Sprintf("\nUsage: %s [globals] <command> [options]\n\n", utils.GetAppName())
		help_text += "Globals:\n"

		globals := make(map[string]string)

		// list global options
		flag.VisitAll(func(f *flag.Flag) {
			if value, ok := globals[f.Usage]; ok {
				// key exists
				if len(f.Name) > len(value) {
					globals[f.Usage] = fmt.Sprintf("-%s, --%s", value, f.Name)
				} else {
					globals[f.Usage] = fmt.Sprintf("-%s, --%s", f.Name, value)
				}
			} else {
				// key does not exist
				globals[f.Usage] = f.Name
			}
		})

		// to store the keys in slice in sorted order
		var keys []string
		text_len := 0
		for k, v := range globals {
			keys = append(keys, k)

			if len(v) > text_len {
				text_len = len(v)
			}
		}
		sort.Strings(keys)

		// to prepare the output format
		for _, k := range keys {
			opt_text := fmt.Sprintf("%s%s", globals[k], strings.Repeat(" ", text_len-len(globals[k])))
			help_text += fmt.Sprintf("  %s  %s\n", opt_text, k)
		}

		help_text += "\nCommands:\n"

		// make commad list
		// to-do: want to change the "command name" that has been hard-coded
		command_list := map[string]command.CmdInterface{
			"es": &command.EsCommand{},
			"ls": &command.LsCommand{},
			"rm": &command.RmCommand{},
		}

		// to store the keys in slice in sorted order
		keys = nil
		text_len = 0
		for k := range command_list {
			keys = append(keys, k)

			if len(k) > text_len {
				text_len = len(k)
			}
		}
		sort.Strings(keys)

		// to prepare the output format
		for _, k := range keys {
			cmd_text := fmt.Sprintf("%s%s", k, strings.Repeat(" ", text_len-len(k)))
			help_text += fmt.Sprintf("  %s  %s\n", cmd_text, command_list[k].Synopsis())
		}

		help_text += "\nOptions:\n"
		help_text += "  show commands options help <command> -h, --help\n"

		// show a help string
		fmt.Println(help_text)
	}
}

var (
	help_flag    bool
	version_flag bool
	config_flag  string
	name_flag    string
)

func resolveArgs() (*config.Config, int) {
	// register flag name
	flag.BoolVar(&help_flag, "help", false, "show this help message and exit")
	flag.BoolVar(&help_flag, "h", false, "show this help message and exit")
	flag.BoolVar(&version_flag, "version", false, "show version message and exit")
	flag.BoolVar(&version_flag, "v", false, "show version message and exit")
	flag.StringVar(&config_flag, "config", "", "specify an alternate config file")
	flag.StringVar(&config_flag, "c", "", "specify an alternate config file")
	flag.StringVar(&name_flag, "name", "default", "specify an alternate rds environment name")
	flag.StringVar(&name_flag, "n", "default", "specify an alternate rds environment name")

	// set help func
	flag.Usage = showHelp()
	flag.Parse()

	// show help
	if help_flag {
		flag.Usage()
		return nil, 0
	}
	// show version
	if version_flag {
		fmt.Printf("%s %s\n", utils.GetAppName(), utils.GetAppVersion())
		return nil, 0
	}
	// show help
	// to-do: want to change the "command name" that has been hard-coded
	if len(flag.Args()) <= 0 || flag.Args()[0] != "es" && flag.Args()[0] != "ls" && flag.Args()[0] != "rm" {
		flag.Usage()
		return nil, 1
	}

	// load config file
	conf_file := config.GetDefaultPath()
	if config_flag != "" {
		conf_file = config_flag
	}
	conf, err := config.LoadConfig(conf_file)
	if err != nil {
		return nil, 1
	}
	log.Debugf("Config: %+v", conf)

	return conf, 0
}

// need to run the caller always "defer log_file.Close()"
func setLogOptions(conf *config.Config) (*os.File, int) {
	// log setting
	if conf.Log.Verbose {
		log.SetLogLevelDebug()
	}
	if conf.Log.Json {
		log.SetJsonLogFormat()
	}
	log_root := utils.GetHomeDir()
	if conf.Log.Root != "" {
		log_root = conf.Log.Root
	}

	log_path := path.Join(log_root, fmt.Sprintf("%s.log", utils.GetFormatedFileDisplayName()))
	// need to run the caller always "defer log_file.Close()"
	log_file, err := os.OpenFile(log_path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, 1
	}

	log.Debugf("Log File: %s", log_path)

	return log_file, 0
}

func getCommandStruct(conf *config.Config) (*command.Command, int) {
	// aws
	aws_creds := conf.GetAWSCreds()
	aws_iam := iam.New(aws_creds, conf.Rds[name_flag].Region, nil)

	// IAM info
	iam_users, err := aws_iam.ListUsers(&iam.ListUsersRequest{})
	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, 1
	}
	if len(iam_users.Users) <= 0 {
		log.Errorf("iam user not found")
		return nil, 1
	}

	// edit IAM ARN
	// arn:aws:iam::<account>:user/<username> to arn:aws:rds:<region>:<account>:
	// see also
	// Tagging Amazon RDS Resources - Amazon Relational Database Service
	// http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html#USER_Tagging.ARN
	iam_split := strings.SplitAfter(*iam_users.Users[0].ARN, "::")
	iam_account := iam_split[len(iam_split)-1]
	iam_split = strings.Split(iam_account, ":")
	iam_account = iam_split[0]

	aws_rds := rds.New(aws_creds, conf.Rds[name_flag].Region, nil)

	cmd_st := &command.Command{
		OutConfig: conf.Out,
		RDSConfig: conf.Rds[name_flag],
		RDSClient: aws_rds,
		ARNPrefix: "arn:aws:rds:" + conf.Rds[name_flag].Region + ":" + iam_account + ":",
	}
	log.Debugf("Command: %+v", cmd_st)

	return cmd_st, 0
}

func main() {
	var ex_code int
	defer func() { os.Exit(ex_code) }()

	// environment variable is set up in order to correspond to multi-core CPU
	if envvar := os.Getenv("GOMAXPROCS"); envvar == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// resolve command line
	conf, ex_code := resolveArgs()
	if ex_code != 0 || conf == nil {
		return
	}

	// log setting
	log_file, ex_code := setLogOptions(conf)
	if ex_code != 0 || log_file == nil {
		return
	}
	defer func() {
		// remove the log file of capacity zero
		fi, err := log_file.Stat()
		log_file.Close()
		if err == nil {
			if fi.Size() <= 0 {
				os.Remove(log_file.Name())
			}
		}
	}()
	log.SetFileOutPut(log_file)

	// check rds env name
	if _, ok := conf.Rds[name_flag]; !ok {
		log.Errorf("rds environment information name not found:[%s]", name_flag)
		ex_code = 1
		return
	}

	// get base command struct
	cmd_st, ex_code := getCommandStruct(conf)
	if ex_code != 0 || cmd_st == nil {
		return
	}

	// call commands
	args := flag.Args()
	var cmd_list command.CmdInterface

	// to-do: want to change the "command name" that has been hard-coded
	switch flag.Args()[0] {
	case "es":
		cmd_list = &command.EsCommand{
			Command: cmd_st,
		}
	case "ls":
		cmd_list = &command.LsCommand{
			Command: cmd_st,
		}
	case "rm":
		cmd_list = &command.RmCommand{
			Command: cmd_st,
		}
	default:
		flag.Usage()
		ex_code = 1
		return
	}

	// run commands
	ex_code = cmd_list.Run(args[1:])
	log.Debugf("ex_code: %d", ex_code)
}
