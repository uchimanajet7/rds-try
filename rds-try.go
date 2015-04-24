package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/iam"
	"github.com/awslabs/aws-sdk-go/service/rds"

	"github.com/uchimanajet7/rds-try/command"
	"github.com/uchimanajet7/rds-try/config"
	"github.com/uchimanajet7/rds-try/logger"
	"github.com/uchimanajet7/rds-try/utils"
)

var log = logger.GetLogger("main")

func showHelp() func() {
	return func() {
		helpText := fmt.Sprintf("\nUsage: %s [globals] <command> [options]\n\n", utils.GetAppName())
		helpText += "Globals:\n"

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
		textLength := 0
		for k, v := range globals {
			keys = append(keys, k)

			if len(v) > textLength {
				textLength = len(v)
			}
		}
		sort.Strings(keys)

		// to prepare the output format
		for _, k := range keys {
			optionText := fmt.Sprintf("%s%s", globals[k], strings.Repeat(" ", textLength-len(globals[k])))
			helpText += fmt.Sprintf("  %s  %s\n", optionText, k)
		}

		helpText += "\nCommands:\n"

		// make commad list
		// to-do: want to change the "command name" that has been hard-coded
		commandList := map[string]command.CmdInterface{
			"es": &command.EsCommand{},
			"ls": &command.LsCommand{},
			"rm": &command.RmCommand{},
		}

		// to store the keys in slice in sorted order
		keys = nil
		textLength = 0
		for k := range commandList {
			keys = append(keys, k)

			if len(k) > textLength {
				textLength = len(k)
			}
		}
		sort.Strings(keys)

		// to prepare the output format
		for _, k := range keys {
			commandText := fmt.Sprintf("%s%s", k, strings.Repeat(" ", textLength-len(k)))
			helpText += fmt.Sprintf("  %s  %s\n", commandText, commandList[k].Synopsis())
		}

		helpText += "\nOptions:\n"
		helpText += "  show commands options help <command> -h, --help\n"

		// show a help string
		fmt.Println(helpText)
	}
}

var (
	helpFlag    bool
	versionFlag bool
	configFlag  string
	nameFlag    string
)

func resolveArgs() (*config.Config, int) {
	// register flag name
	flag.BoolVar(&helpFlag, "help", false, "show this help message and exit")
	flag.BoolVar(&helpFlag, "h", false, "show this help message and exit")
	flag.BoolVar(&versionFlag, "version", false, "show version message and exit")
	flag.BoolVar(&versionFlag, "v", false, "show version message and exit")
	flag.StringVar(&configFlag, "config", "", "specify an alternate config file")
	flag.StringVar(&configFlag, "c", "", "specify an alternate config file")
	flag.StringVar(&nameFlag, "name", "default", "specify an alternate rds environment name")
	flag.StringVar(&nameFlag, "n", "default", "specify an alternate rds environment name")

	// set help func
	flag.Usage = showHelp()
	flag.Parse()

	// show help
	if helpFlag {
		flag.Usage()
		return nil, 0
	}
	// show version
	if versionFlag {
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
	configFile := config.GetDefaultPath()
	if configFlag != "" {
		configFile = configFlag
	}
	conf, err := config.LoadConfig(configFile)
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
	if conf.Log.JSON {
		log.SetJSONLogFormat()
	}
	logRoot := utils.GetHomeDir()
	if conf.Log.Root != "" {
		logRoot = conf.Log.Root
	}

	logPath := path.Join(logRoot, fmt.Sprintf("%s.log", utils.GetFormatedFileDisplayName()))
	// need to run the caller always "defer log_file.Close()"
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, 1
	}

	log.Debugf("Log File: %s", logPath)

	return logFile, 0
}

func getCommandStruct(conf *config.Config) (*command.Command, int) {
	// aws
	awsConfig := aws.DefaultConfig
	awsConfig.Credentials = conf.GetAWSCreds()
	awsConfig.Region = conf.Rds[nameFlag].Region

	// new iam
	awsIam := iam.New(awsConfig)

	// IAM info
	iamUsers, err := awsIam.ListUsers(&iam.ListUsersInput{})
	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, 1
	}
	if len(iamUsers.Users) <= 0 {
		log.Errorf("iam user not found")
		return nil, 1
	}

	// edit IAM ARN
	// arn:aws:iam::<account>:user/<username> to arn:aws:rds:<region>:<account>:
	// see also
	// Tagging Amazon RDS Resources - Amazon Relational Database Service
	// http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html#USER_Tagging.ARN
	iamSplit := strings.SplitAfter(*iamUsers.Users[0].ARN, "::")
	iamAccount := iamSplit[len(iamSplit)-1]
	iamSplit = strings.Split(iamAccount, ":")
	iamAccount = iamSplit[0]

	// new rds
	awsRds := rds.New(awsConfig)

	commandStruct := &command.Command{
		OutConfig: conf.Out,
		RDSConfig: conf.Rds[nameFlag],
		RDSClient: awsRds,
		ARNPrefix: "arn:aws:rds:" + conf.Rds[nameFlag].Region + ":" + iamAccount + ":",
	}
	log.Debugf("Command: %+v", commandStruct)

	return commandStruct, 0
}

func main() {
	var exCode int
	defer func() { os.Exit(exCode) }()

	// environment variable is set up in order to correspond to multi-core CPU
	if envvar := os.Getenv("GOMAXPROCS"); envvar == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// resolve command line
	conf, exCode := resolveArgs()
	if exCode != 0 || conf == nil {
		return
	}

	// log setting
	logFile, exCode := setLogOptions(conf)
	if exCode != 0 || logFile == nil {
		return
	}
	defer func() {
		// remove the log file of capacity zero
		fi, err := logFile.Stat()
		logFile.Close()
		if err == nil {
			if fi.Size() <= 0 {
				os.Remove(logFile.Name())
			}
		}
	}()
	log.SetFileOutPut(logFile)

	// check rds env name
	if _, ok := conf.Rds[nameFlag]; !ok {
		log.Errorf("rds environment information name not found:[%s]", nameFlag)
		exCode = 1
		return
	}

	// get base command struct
	commandStruct, exCode := getCommandStruct(conf)
	if exCode != 0 || commandStruct == nil {
		return
	}

	// call commands
	args := flag.Args()
	var commandList command.CmdInterface

	// to-do: want to change the "command name" that has been hard-coded
	switch flag.Args()[0] {
	case "es":
		commandList = &command.EsCommand{
			Command: commandStruct,
		}
	case "ls":
		commandList = &command.LsCommand{
			Command: commandStruct,
		}
	case "rm":
		commandList = &command.RmCommand{
			Command: commandStruct,
		}
	default:
		flag.Usage()
		exCode = 1
		return
	}

	// run commands
	exCode = commandList.Run(args[1:])
	log.Debugf("ex_code: %d", exCode)
}
