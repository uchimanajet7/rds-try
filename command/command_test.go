package command

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/rds"
	testdb "github.com/erikstmartin/go-testdb"

	"github.com/uchimanajet7/rds-try/config"
	"github.com/uchimanajet7/rds-try/query"
	"github.com/uchimanajet7/rds-try/utils"
)

// return "httptest.Server" need call close !!
// and Command into OutConfig.Root need call remove !!
func getTestClient(code int, body string) (*httptest.Server, *Command) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintln(w, body)
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	httpClient := &http.Client{Transport: transport}

	// Override endpoints
	testRegion := "rds-try-test-1"
	awsConf := aws.DefaultConfig
	awsConf.Credentials = credentials.NewStaticCredentials("awsAccesskey1", "awsSecretKey2", "")
	awsConf.Region = testRegion
	awsConf.Endpoint = server.URL
	awsConf.HTTPClient = httpClient

	awsRds := rds.New(awsConf)

	testName := utils.GetAppName() + "-test"
	tempDir, _ := ioutil.TempDir("", testName)
	defer os.RemoveAll(tempDir)

	out := config.OutConfig{
		Root: tempDir,
		File: true,
		Bom:  true,
	}

	rds := config.RDSConfig{
		MultiAz: false,
		DBId:    utils.GetFormatedDBDisplayName(testName),
		Region:  testRegion,
		User:    "test-admin",
		Pass:    "pass-pass",
		Type:    "db.m3.medium",
	}

	cmdTest := &Command{
		OutConfig: out,
		RDSConfig: rds,
		RDSClient: awsRds,
		ARNPrefix: "arn:aws:rds:" + testRegion + ":" + "123456789" + ":",
	}

	return server, cmdTest
}

func TestDescribeDBInstances(t *testing.T) {
	ts, tc := getTestClient(200, srDescribeDBInstancesResponse)
	defer ts.Close()

	input := &rds.DescribeDBInstancesInput{}
	ri, err := tc.describeDBInstances(input)

	if err != nil {
		t.Errorf("[describeDBInstances] result error: %s", err.Error())
	}
	if len(ri) != 3 {
		t.Errorf("DB Instance count not match: %d", len(ri))
	}
}

func TestDescribeDBInstance(t *testing.T) {
	ts, tc := getTestClient(200, srDescribeDBInstanceResponse)
	defer ts.Close()

	id := "rds-try-test-db-1"
	ri, err := tc.DescribeDBInstance(id)

	if err != nil {
		t.Errorf("[DescribeDBInstance] result error: %s", err.Error())
	}
	if *ri.DBInstanceIdentifier != id {
		t.Errorf("DBInstanceIdentifier not match: %s/%s", *ri.DBInstanceIdentifier, id)
	}
}

func TestCheckListTagsForResourceMessage(t *testing.T) {
	ts, tc := getTestClient(200, srListTagsForResourceResponse)
	defer ts.Close()

	dbID := "rds-try-test"
	var ti rds.DBInstance
	ti.DBInstanceIdentifier = &dbID

	ri, err := tc.checkListTagsForResource(&ti)

	if err != nil {
		t.Errorf("[checkListTagsFor] result error: %s", err.Error())
	}
	if !ri {
		t.Error("tags count not match")
	}
}

func TestModifyDBInstance(t *testing.T) {
	ts, tc := getTestClient(200, srDescribeDBInstanceResponse)
	defer ts.Close()

	id := "rds-try-test-db-1"
	ri, err := tc.DescribeDBInstance(id)

	tsm, tcm := getTestClient(200, srModifyDBInstanceResponse)
	defer tsm.Close()

	rim, err := tcm.ModifyDBInstance(id, ri)

	if err != nil {
		t.Errorf("[ModifyDBInstance] result error: %s", err.Error())
	}
	if *rim.DBInstanceIdentifier != id {
		t.Errorf("DBInstanceIdentifier not match: %s/%s", *rim.DBInstanceIdentifier, id)
	}
}

func TestRebootDBInstance(t *testing.T) {
	ts, tc := getTestClient(200, srRebootDBInstanceResponse)
	defer ts.Close()

	id := "rds-try-test-db-1"
	ri, err := tc.RebootDBInstance(id)

	if err != nil {
		t.Errorf("[RebootDBInstance] result error: %s", err.Error())
	}
	if *ri.DBInstanceIdentifier != id {
		t.Errorf("DBInstanceIdentifier not match: %s/%s", *ri.DBInstanceIdentifier, id)
	}
	if *ri.DBInstanceStatus != "rebooting" {
		t.Errorf("DBInstanceStatus not match: %s/%s", *ri.DBInstanceStatus, "rebooting")
	}
}

func TestDescribeDBSnapshots(t *testing.T) {
	ts, tc := getTestClient(200, srDescribeDBSnapshotsResponse)
	defer ts.Close()

	input := &rds.DescribeDBSnapshotsInput{}
	ri, err := tc.describeDBSnapshots(input)

	if err != nil {
		t.Errorf("[describeDBSnapshots] result error: %s", err.Error())
	}
	if len(ri) != 3 {
		t.Errorf("DB Snapshot count not match: %d", len(ri))
	}
}

func TestDescribeLatestDBSnapshot(t *testing.T) {
	ts, tc := getTestClient(200, srDescribeDBSnapshotResponse)
	defer ts.Close()

	id := "rds-try-test-db-1"
	ri, err := tc.DescribeLatestDBSnapshot(id)

	if err != nil {
		t.Errorf("[DescribeLatestDBSnapshot] result error: %s", err.Error())
	}
	if *ri.DBInstanceIdentifier != id {
		t.Errorf("DBInstanceIdentifier not match: %s/%s", *ri.DBInstanceIdentifier, id)
	}
}

func TestDescribeDBSnapshot(t *testing.T) {
	ts, tc := getTestClient(200, srDescribeDBSnapshotResponse)
	defer ts.Close()

	id := "before-test-1"
	ri, err := tc.DescribeDBSnapshot(id)

	if err != nil {
		t.Errorf("[DescribeDBSnapshot] result error: %s", err.Error())
	}
	if *ri.DBSnapshotIdentifier != id {
		t.Errorf("DBSnapshotIdentifier not match: %s/%s", *ri.DBSnapshotIdentifier, id)
	}
}

func TestDeleteDBInstance(t *testing.T) {
	ts, tc := getTestClient(200, srDeleteDBInstanceResponse)
	defer ts.Close()

	id := "rds-try-test-db-1"
	ri, err := tc.DeleteDBInstance(id)

	if err != nil {
		t.Errorf("[DeleteDBInstance] result error: %s", err.Error())
	}
	if *ri.DBInstanceIdentifier != id {
		t.Errorf("DBInstanceIdentifier not match: %s/%s", *ri.DBInstanceIdentifier, id)
	}
	if *ri.DBInstanceStatus != "deleting" {
		t.Errorf("DBInstanceStatus not match: %s/%s", *ri.DBInstanceStatus, "deleting")
	}
}

func TestCreateDBSnapshot(t *testing.T) {
	ts, tc := getTestClient(200, srCreateDBSnapshotResponse)
	defer ts.Close()

	id := "rds-try-test-db-1"
	ri, err := tc.CreateDBSnapshot(id)

	if err != nil {
		t.Errorf("[CreateDBSnapshot] result error: %s", err.Error())
	}
	if *ri.DBInstanceIdentifier != id {
		t.Errorf("DBInstanceIdentifier not match: %s/%s", *ri.DBInstanceIdentifier, id)
	}
}

func TestDeleteDBSnapshot(t *testing.T) {
	ts, tc := getTestClient(200, srDeleteDBSnapshotResponse)
	defer ts.Close()

	id := "before-test-1"
	ri, err := tc.DeleteDBSnapshot(id)

	if err != nil {
		t.Errorf("[DeleteDBSnapshot] result error: %s", err.Error())
	}
	if *ri.DBSnapshotIdentifier != id {
		t.Errorf("DBSnapshotIdentifier not match: %s/%s", *ri.DBSnapshotIdentifier, id)
	}
}

func TestCheckPendingStatus(t *testing.T) {
	ts, tc := getTestClient(200, srDescribeDBInstanceResponse)
	defer ts.Close()

	id := "rds-try-test-db-1"
	ri, _ := tc.DescribeDBInstance(id)
	state := tc.CheckPendingStatus(ri)

	if state {
		t.Error("CheckPendingStatus not match")
	}
}

func TestRestoreDBInstanceFromDBSnapshot(t *testing.T) {
	ts, tc := getTestClient(200, srDescribeDBInstanceResponse)
	defer ts.Close()

	id := "rds-try-test-db-1"
	ri, _ := tc.DescribeDBInstance(id)

	tss, tcs := getTestClient(200, srDescribeDBSnapshotResponse)
	defer tss.Close()

	ids := "before-test-1"
	ris, _ := tcs.DescribeDBSnapshot(ids)

	tsr, tcr := getTestClient(200, srRestoreDBInstanceFromDBSnapshotResponse)
	defer tsr.Close()

	class := "db.t1.micro"
	az := false
	args := &RestoreDBInstanceFromDBSnapshotArgs{
		DBInstanceClass: class,
		DBIdentifier:    id,
		MultiAZ:         az,
		Snapshot:        ris,
		Instance:        ri,
	}

	rir, err := tcr.RestoreDBInstanceFromDBSnapshot(args)

	if err != nil {
		t.Errorf("[RestoreDBInstanceFromDBSnapshot] result error: %s", err.Error())
	}
	if *rir.DBInstanceIdentifier != id {
		t.Errorf("DBInstanceIdentifier not match: %s/%s", *rir.DBInstanceIdentifier, id)
	}
	if *rir.DBInstanceStatus != "creating" {
		t.Errorf("DBInstanceStatus not match: %s/%s", *rir.DBInstanceStatus, "creating")
	}
}

func TestGetARNString(t *testing.T) {
	ts, tc := getTestClient(200, "")
	defer ts.Close()

	id := "rds-try-test-db-1"
	di := &rds.DBInstance{
		DBInstanceIdentifier: &id,
	}

	arn := tc.getARNString(di)

	if arn == "" {
		t.Error("ARN string is nil")
	}
	if !strings.Contains(arn, id) {
		t.Errorf("ARN string not contains id: %s", arn)
	}
}

func TestGetSpecifyTags(t *testing.T) {
	tags := getSpecifyTags()

	if len(tags) != 2 {
		t.Errorf("GetSpecifyTags count not match: %d", len(tags))
	}
	for _, i := range tags {
		switch *i.Key {
		case rtNameText, rtTimeText:
			continue
		default:
			t.Errorf("GetSpecifyTags key name not match: %s", *i.Key)
		}
	}
}

func TestGetDbOpenValues(t *testing.T) {
	eg := "mysql"
	var port int64
	port = 3306
	addr := "rds-try-test-db.c6c2mntzugv0.us-west-2.rds.amazonaws.com"
	ep := &rds.Endpoint{
		Port:    &port,
		Address: &addr,
	}
	q := []query.Query{
		{
			Name: "q1",
			Sql:  "select * from account_id",
		},
		{
			Name: "q2",
			Sql:  "select count(*) from account_id",
		},
	}
	args := &ExecuteSQLArgs{
		Engine:   eg,
		Endpoint: ep,
		Queries:  q,
	}

	ts, tc := getTestClient(200, "")
	defer ts.Close()

	d, dn := tc.getDbOpenValues(args)

	if d != eg {
		t.Errorf("DB driver name not match: %s/%s", d, eg)
	}
	if !strings.Contains(dn, addr) {
		t.Errorf("DB data source name not match: %s", dn)
	}
}

func TestWriteCSVFile(t *testing.T) {
	testName := utils.GetAppName() + "-test"
	tempDir, _ := ioutil.TempDir("", testName)

	tempFile, err := ioutil.TempFile(tempDir, utils.GetAppName()+"-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}
	fStat, _ := tempFile.Stat()
	fName := tempFile.Name()

	tempFile.Sync()
	tempFile.Close()
	defer os.RemoveAll(tempDir)

	db, _ := sql.Open("testdb", "")
	defer db.Close()

	sql := "select id, name, age from users"
	columns := []string{"id", "name", "age", "created"}
	result := `
1,tim,20,2012-10-01 01:00:01
2,joe,25,2012-10-02 02:00:02
3,bob,30,2012-10-03 03:00:03
`
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	res, err := db.Query(sql)
	args := &writeCSVFileArgs{
		Rows:     res,
		FileName: fStat.Name(),
		Path:     tempDir,
		Bom:      false,
	}

	csvState := writeCSVFile(args)

	file, err := os.OpenFile(fName, os.O_RDONLY, 0777)
	defer file.Close()

	fStat, _ = file.Stat()

	if !csvState {
		t.Errorf("[writeCSVFile] result error: %s", err.Error())
	}
	if fStat.Size() <= 0 {
		t.Errorf("csv file not out put: %d", fStat.Size())
	}
}

// It takes 30 seconds every time
func TestWaitForStatusAvailable(t *testing.T) {
	ts, tc := getTestClient(200, srDescribeDBInstanceResponse)
	defer ts.Close()

	t.Log("It takes 30 seconds every time")
	fmt.Println(" ### It takes 30 seconds every time")

	id := "rds-try-test-db-1"
	ri, _ := tc.DescribeDBInstance(id)
	state := tc.WaitForStatusAvailable(ri)

	if !<-state {
		t.Error("WaitForStatusAvailable not match")
	}
}
