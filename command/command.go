package command

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // required as SQL driver at the time of connection

	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/uchimanajet7/rds-try/config"
	"github.com/uchimanajet7/rds-try/logger"
	"github.com/uchimanajet7/rds-try/query"
	"github.com/uchimanajet7/rds-try/utils"
)

// CmdInterface interface is the Help and Run and Synopsis function
type CmdInterface interface {
	// return long help string
	Help() string

	// run command with arguments
	Run(args []string) int

	// return short help string
	Synopsis() string
}

// Command struct is the OutConfig and RDSConfig and RDSClient and ARNPrefix variable
type Command struct {
	OutConfig config.OutConfig
	RDSConfig config.RDSConfig
	RDSClient *rds.RDS
	ARNPrefix string
}

var log = logger.GetLogger("command")

var (
	// ErrDBInstancetNotFound is the "DB Instance is not found" error
	ErrDBInstancetNotFound = errors.New("DB Instance is not found")
	// ErrSnapshotNotFound is the "DB　Snapshot is not found" error
	ErrSnapshotNotFound = errors.New("DB　Snapshot is not found")
	// ErrDriverNotFound is the "DB　Driver is not found" error
	ErrDriverNotFound = errors.New("DB　Driver is not found")
	// ErrRdsTypesNotFound is the "RDS Types is not found" error
	ErrRdsTypesNotFound = errors.New("RDS Types is not found")
	// ErrRdsARNsNotFound is the "RDS ARN Types is not found" error
	ErrRdsARNsNotFound = errors.New("RDS ARN Types is not found")
)

func (c *Command) describeDBInstances(input *rds.DescribeDBInstancesInput) ([]*rds.DBInstance, error) {
	output, err := c.RDSClient.DescribeDBInstances(input)

	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}

	return output.DBInstances, err
}

// DescribeDBInstance is show the aws rds db instance infomations
// all status in target, result return only one
func (c *Command) DescribeDBInstance(dbIdentifier string) (*rds.DBInstance, error) {
	// set DB ID
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: &dbIdentifier,
	}
	output, err := c.describeDBInstances(input)

	if err != nil {
		return nil, err
	}

	dbLen := len(output)
	if dbLen < 1 {
		log.Errorf("%s", ErrDBInstancetNotFound.Error())
		return nil, ErrDBInstancetNotFound
	}

	return output[dbLen-1], err
}

// check tag count
func (c *Command) checkListTagsForResource(rdstypes interface{}) (bool, error) {
	// want to filter by tag name and value
	// see also
	// ListTagsForResource - Amazon Relational Database Service
	// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_ListTagsForResource.html

	// get tag list
	state := false
	arn := c.getARNString(rdstypes)
	if arn == "" {
		log.Errorf("%s", ErrRdsARNsNotFound.Error())
		return state, ErrRdsARNsNotFound
	}

	tagOutput, err := c.RDSClient.ListTagsForResource(
		&rds.ListTagsForResourceInput{
			ResourceName: &arn,
		})

	if err != nil {
		log.Errorf("%s", err.Error())
		return state, err
	}
	if len(tagOutput.TagList) <= 0 {
		return state, err
	}

	// check tag name and value
	tagCount := 0
	for _, tag := range tagOutput.TagList {
		switch *tag.Key {
		case rtNameText:
			// if the rt_name tag exists, should the prefix value has become an application name
			if strings.HasPrefix(*tag.Value, utils.GetPrefix()) {
				tagCount++
			}
		case rtTimeText:
			// rt_time tag exists
			tagCount++
		}
	}

	// to-do: Fixed value the number you are using to the judgment has been hard-coded
	if tagCount >= 2 {
		state = true
		return state, err
	}

	return state, err
}

// DescribeDBInstancesByTags is aws rds db instances list up by tags
func (c *Command) DescribeDBInstancesByTags() ([]*rds.DBInstance, error) {
	input := &rds.DescribeDBInstancesInput{}

	output, err := c.describeDBInstances(input)

	if err != nil {
		return nil, err
	}

	var dbInstances []*rds.DBInstance
	for _, instance := range output {
		state, err := c.checkListTagsForResource(instance)
		if err != nil {
			return nil, err
		}

		if state {
			dbInstances = append(dbInstances, instance)
		}
	}

	return dbInstances, err
}

// ModifyDBInstance is modify aws rds db instance setting
func (c *Command) ModifyDBInstance(dbIdentifier string, dbInstance *rds.DBInstance) (*rds.DBInstance, error) {
	var vpcIDs []*string
	for _, vpcID := range dbInstance.VPCSecurityGroups {
		vpcIDs = append(vpcIDs, vpcID.VPCSecurityGroupID)
	}

	apply := true
	input := &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: &dbIdentifier,
		DBParameterGroupName: dbInstance.DBParameterGroups[0].DBParameterGroupName,
		VPCSecurityGroupIDs:  vpcIDs,
		ApplyImmediately:     &apply, // "ApplyImmediately" is always true
	}

	output, err := c.RDSClient.ModifyDBInstance(input)

	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}

	return output.DBInstance, err
}

// RebootDBInstance is reboot aws rds db instance
func (c *Command) RebootDBInstance(dbIdentifier string) (*rds.DBInstance, error) {
	input := &rds.RebootDBInstanceInput{
		DBInstanceIdentifier: &dbIdentifier,
	}

	output, err := c.RDSClient.RebootDBInstance(input)

	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}

	return output.DBInstance, err
}

// RestoreDBInstanceFromDBSnapshotArgs struct is the DBInstanceClass and DBIdentifier and MultiAZ and Snapshot and Instance variable
type RestoreDBInstanceFromDBSnapshotArgs struct {
	DBInstanceClass string
	DBIdentifier    string
	MultiAZ         bool
	Snapshot        *rds.DBSnapshot
	Instance        *rds.DBInstance
}

// RestoreDBInstanceFromDBSnapshot is restore aws rds db instance from db snap shot
func (c *Command) RestoreDBInstanceFromDBSnapshot(args *RestoreDBInstanceFromDBSnapshotArgs) (*rds.DBInstance, error) {
	input := &rds.RestoreDBInstanceFromDBSnapshotInput{
		DBInstanceClass:      &args.DBInstanceClass,
		DBInstanceIdentifier: &args.DBIdentifier,
		MultiAZ:              &args.MultiAZ,
		DBSnapshotIdentifier: args.Snapshot.DBSnapshotIdentifier,
		DBSubnetGroupName:    args.Instance.DBSubnetGroup.DBSubnetGroupName,
		StorageType:          args.Instance.StorageType,
		Tags:                 getSpecifyTags(), // It must always be set to not forget
	}

	output, err := c.RDSClient.RestoreDBInstanceFromDBSnapshot(input)

	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}

	return output.DBInstance, err
}

// DescribeDBSnapshotsByTags is show aws rds snap shot by tags
func (c *Command) DescribeDBSnapshotsByTags() ([]*rds.DBSnapshot, error) {
	input := &rds.DescribeDBSnapshotsInput{}

	output, err := c.describeDBSnapshots(input)

	if err != nil {
		return nil, err
	}

	var dbSnapshots []*rds.DBSnapshot
	for _, snapshot := range output {
		state, err := c.checkListTagsForResource(snapshot)
		if err != nil {
			return nil, err
		}

		if state {
			dbSnapshots = append(dbSnapshots, snapshot)
		}
	}

	return dbSnapshots, err
}

func (c *Command) describeDBSnapshots(input *rds.DescribeDBSnapshotsInput) ([]*rds.DBSnapshot, error) {
	output, err := c.RDSClient.DescribeDBSnapshots(input)

	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}

	return output.DBSnapshots, err
}

// DescribeLatestDBSnapshot is show latest aws rds db snap shot
// the target only "available"
func (c *Command) DescribeLatestDBSnapshot(dbIdentifier string) (*rds.DBSnapshot, error) {
	input := &rds.DescribeDBSnapshotsInput{
		DBInstanceIdentifier: &dbIdentifier,
	}

	output, err := c.describeDBSnapshots(input)

	if err != nil {
		return nil, err
	}

	// want to filter by status "available"
	var dbSnapshots []*rds.DBSnapshot
	for _, snapshot := range output {
		if *snapshot.Status != "available" {
			log.Debugf("DB Snapshot Status : %s", *snapshot.Status)
			continue
		}

		dbSnapshots = append(dbSnapshots, snapshot)
	}

	dbLen := len(dbSnapshots)
	if dbLen < 1 {
		log.Errorf("%s", ErrSnapshotNotFound.Error())
		return nil, ErrSnapshotNotFound
	}

	return dbSnapshots[dbLen-1], err
}

// DescribeDBSnapshot is show aws rds db snap shot
// all status in target, result return only one
func (c *Command) DescribeDBSnapshot(snapshotIdentifier string) (*rds.DBSnapshot, error) {
	input := &rds.DescribeDBSnapshotsInput{
		DBSnapshotIdentifier: &snapshotIdentifier,
	}

	output, err := c.describeDBSnapshots(input)

	if err != nil {
		return nil, err
	}

	dbLen := len(output)
	if dbLen < 1 {
		log.Errorf("%s", ErrSnapshotNotFound.Error())
		return nil, ErrSnapshotNotFound
	}

	return output[dbLen-1], err
}

// DeleteDBInstance is delete aws rds db instance
// delete DB instance and skip create snapshot
func (c *Command) DeleteDBInstance(dbIdentifier string) (*rds.DBInstance, error) {
	skip := true
	input := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: &dbIdentifier,
		SkipFinalSnapshot:    &skip, // "SkipFinalSnapshot" is always true
	}

	output, err := c.RDSClient.DeleteDBInstance(input)

	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}

	return output.DBInstance, err
}

// CreateDBSnapshot is create aws rds db snap shot
func (c *Command) CreateDBSnapshot(dbIdentifier string) (*rds.DBSnapshot, error) {
	snapshotID := utils.GetFormatedDBDisplayName(dbIdentifier)
	input := &rds.CreateDBSnapshotInput{
		DBInstanceIdentifier: &dbIdentifier,
		DBSnapshotIdentifier: &snapshotID,
		Tags:                 getSpecifyTags(), // It must always be set to not forget
	}

	output, err := c.RDSClient.CreateDBSnapshot(input)

	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}

	return output.DBSnapshot, err
}

// DeleteDBSnapshot is delete aws rds db snap shot
func (c *Command) DeleteDBSnapshot(snapshotIdentifier string) (*rds.DBSnapshot, error) {
	input := &rds.DeleteDBSnapshotInput{
		DBSnapshotIdentifier: &snapshotIdentifier,
	}

	output, err := c.RDSClient.DeleteDBSnapshot(input)

	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}

	return output.DBSnapshot, err
}

// CheckPendingStatus is check pending status aws rds db instance setting
// "Pending Status" If the return value is ture
func (c *Command) CheckPendingStatus(dbInstance *rds.DBInstance) bool {
	for _, item := range dbInstance.DBParameterGroups {
		if *item.ParameterApplyStatus != "in-sync" {
			return true
		}
	}

	for _, item := range dbInstance.VPCSecurityGroups {
		if *item.Status != "active" {
			return true
		}
	}

	return false
}

// DeleteDBResources is aws rds db instance or snap shot
func (c *Command) DeleteDBResources(rdstypes interface{}) error {
	switch rdstype := rdstypes.(type) {
	case []*rds.DBSnapshot:
		for i, item := range rdstype {
			resp, err := c.DeleteDBSnapshot(*item.DBSnapshotIdentifier)
			if err != nil {
				return err
			}
			log.Infof("[% d] deleted DB Snapshot: %s", i+1, *resp.DBSnapshotIdentifier)
		}
	case []*rds.DBInstance:
		for i, item := range rdstype {
			resp, err := c.DeleteDBInstance(*item.DBInstanceIdentifier)
			if err != nil {
				return err
			}
			log.Infof("[% d] deleted DB Instance: %s", i+1, *resp.DBInstanceIdentifier)
		}
	default:
		log.Errorf("%s", ErrRdsTypesNotFound.Error())
	}

	return nil
}

// WaitForStatusAvailable is the 30 seconds intervals checked aws rds state
// wait for status available
func (c *Command) WaitForStatusAvailable(rdstypes interface{}) <-chan bool {
	receiver := make(chan bool)
	// 30 seconds intervals checked
	ticker := time.NewTicker(30 * time.Second)
	// 30 minutes time out
	timeout := time.After(30 * time.Minute)

	go func() {
		for {
			select {
			case tick := <-ticker.C:
				var rdsStatus string

				log.Debugf("tick: %s", tick)

				switch rdstype := rdstypes.(type) {
				case *rds.DBSnapshot:
					dbSnapshot, err := c.DescribeDBSnapshot(*rdstype.DBSnapshotIdentifier)

					if err != nil {
						receiver <- false

						ticker.Stop()
					}

					rdsStatus = *dbSnapshot.Status
					log.Infof("DB Snapshot Status: %s", rdsStatus)
				case *rds.DBInstance:
					dbInstance, err := c.DescribeDBInstance(*rdstype.DBInstanceIdentifier)

					if err != nil {
						receiver <- false

						ticker.Stop()
					}

					rdsStatus = *dbInstance.DBInstanceStatus
					log.Infof("DB Instance Status: %s", rdsStatus)
				default:
					log.Errorf("%s", ErrRdsTypesNotFound.Error())
				}

				if rdsStatus == "available" {
					receiver <- true
					log.Infof("Status: %s", rdsStatus)

					ticker.Stop()
				}
			case out := <-timeout:
				receiver <- false
				log.Infof("time out: %s", out)

				ticker.Stop()
			}
		}
	}()

	return receiver
}

// ExecuteSQLArgs struct is Engine and Endpoint and Queries variable
type ExecuteSQLArgs struct {
	Engine   string // rds engine name
	Endpoint *rds.Endpoint
	Queries  []query.Query
}

// ExecuteSQL is execute SQL to aws rds
func (c *Command) ExecuteSQL(args *ExecuteSQLArgs) ([]time.Duration, error) {
	driver, dsn := c.getDbOpenValues(args)

	if driver == "" {
		log.Errorf("%s", ErrDriverNotFound.Error())
		return nil, ErrDriverNotFound
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}
	defer db.Close()

	times := make([]time.Duration, 0, len(args.Queries))
	for _, value := range args.Queries {
		log.Debugf("query value : %s", value)

		sTime := time.Now()
		log.Infof("query start time: %s", sTime)

		result, err := db.Query(value.SQL)
		if err != nil {
			log.Errorf("%s", err.Error())
			return times, err
		}

		eTime := time.Now()
		log.Infof("query end time: %s", eTime)

		times = append(times, eTime.Sub(sTime))

		// output csv file
		cols, _ := result.Columns()
		if c.OutConfig.File && len(cols) > 0 {
			fileName := value.Name + "-" + utils.GetFormatedTime() + ".csv"
			outPath := utils.GetHomeDir()
			if c.OutConfig.Root != "" {
				outPath = c.OutConfig.Root
			}

			outState := writeCSVFile(
				&writeCSVFileArgs{
					Rows:     result,
					FileName: fileName,
					Path:     outPath,
					Bom:      c.OutConfig.Bom,
				})
			log.Debugf("out_state:%+v", outState)
		}

		result.Close()
	}

	return times, nil
}

func (c *Command) getDbOpenValues(args *ExecuteSQLArgs) (string, string) {
	var driverName string
	var dataSourceName string

	engine := strings.ToLower(args.Engine)
	log.Debugf("aws engine name: %s", engine)

	// convert from "aws engine name" to "golang db driver name"
	// see also
	// CreateDBInstance - Amazon Relational Database Service
	// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstance.html
	// Request Parameters "Engine" is Valid Values
	// Valid Values: MySQL | oracle-se1 | oracle-se | oracle-ee | sqlserver-ee | sqlserver-se | sqlserver-ex | sqlserver-web | postgres
	//
	// SQLDrivers · golang/go Wiki · GitHub
	// https://github.com/golang/go/wiki/SQLDrivers
	//
	// to-do: correspondence of mysql only
	switch {
	case strings.Contains(engine, "mysql"):
		driverName = "mysql"
		dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%d)/", c.RDSConfig.User, c.RDSConfig.Pass, *args.Endpoint.Address, *args.Endpoint.Port)
	case strings.Contains(engine, "oracle"):
		driverName = "oracle"
	case strings.Contains(engine, "sqlserver"):
		driverName = "sqlserver"
	case strings.Contains(engine, "postgres"):
		driverName = "postgres"
	default:
		log.Errorf("failed to convert. no matching SQL driver: %s", engine)
	}

	log.Debugf("golang db driver name: %s", driverName)
	log.Debugf("golang db data source name: %s", dataSourceName)

	return driverName, dataSourceName
}

func (c *Command) getARNString(rdstypes interface{}) string {
	// edit ARN string
	// see also
	// Tagging Amazon RDS Resources - Amazon Relational Database Service
	// http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html

	var arn string
	switch rdstype := rdstypes.(type) {
	case *rds.DBSnapshot:
		arn = c.ARNPrefix + "snapshot:" + *rdstype.DBSnapshotIdentifier
	case *rds.DBInstance:
		arn = c.ARNPrefix + "db:" + *rdstype.DBInstanceIdentifier
	default:
		log.Errorf("%s", ErrRdsARNsNotFound.Error())
	}

	log.Debugf("ARN: %s", arn)

	return arn
}

const rtNameText = "rt_name"
const rtTimeText = "rt_time"

// use the tag for identification
func getSpecifyTags() []*rds.Tag {
	var tagList []*rds.Tag

	// append name
	keyName := rtNameText
	valueName := utils.GetFormatedAppName()
	tagName := &rds.Tag{
		Key:   &keyName,
		Value: &valueName,
	}
	tagList = append(tagList, tagName)

	// append time
	keyTime := rtTimeText
	valueTime := utils.GetFormatedTime()
	tagTime := &rds.Tag{
		Key:   &keyTime,
		Value: &valueTime,
	}
	tagList = append(tagList, tagTime)

	return tagList
}

type writeCSVFileArgs struct {
	Rows     *sql.Rows
	FileName string
	Path     string
	Bom      bool
}

func writeCSVFile(args *writeCSVFileArgs) bool {
	const BOM = string('\uFEFF')

	cols, err := args.Rows.Columns()
	if err != nil {
		log.Errorf("%s", err.Error())
		return false
	}

	// is append bom?
	if args.Bom {
		// When making the extension a txt, UTF8 can be used in Excel.
		args.FileName = fmt.Sprintf("utf8-bom_%s", args.FileName)
	}
	outPath := path.Join(args.Path, args.FileName)

	// all user access OK
	file, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, 0777)
	defer file.Close()

	// set empty
	err = file.Truncate(0)

	// write csv
	writer := csv.NewWriter(file)

	// add BOM
	if args.Bom {
		boms := make([]string, 1)
		boms[0] = BOM + fmt.Sprintf("# character encoding : utf-8 with BOM")
		writer.Write(boms)
	}

	// write header
	writer.Write(cols)

	// Result is your slice string.
	rawResult := make([][]byte, len(cols))
	result := make([]string, len(cols))

	// A temporary interface{} slice
	dest := make([]interface{}, len(cols))
	// Put pointers to each string in the interface slice
	for i := range rawResult {
		dest[i] = &rawResult[i]
	}

	for args.Rows.Next() {
		err = args.Rows.Scan(dest...)
		if err != nil {
			log.Errorf("%s", err.Error())
		}

		for i, raw := range rawResult {
			if raw == nil {
				result[i] = "null"
			} else {
				result[i] = string(raw)
			}
		}
		writer.Write(result)
	}
	writer.Flush()

	return true
}
