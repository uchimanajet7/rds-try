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

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/rds"
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
	test_region := "rds-try-test-1"
	aws_conf := aws.DefaultConfig
	aws_conf.Credentials = aws.Creds("awsAccesskey1", "awsSecretKey2", "")
	aws_conf.Region = test_region
	aws_conf.Endpoint = server.URL
	aws_conf.HTTPClient = httpClient

	aws_rds := rds.New(aws_conf)

	test_name := utils.GetAppName() + "-test"
	temp_dir, _ := ioutil.TempDir("", test_name)
	defer os.RemoveAll(temp_dir)

	out := config.OutConfig{
		Root: temp_dir,
		File: true,
		Bom:  true,
	}

	rds := config.RDSConfig{
		MultiAz: false,
		DBId:    utils.GetFormatedDBDisplayName(test_name),
		Region:  test_region,
		User:    "test-admin",
		Pass:    "pass-pass",
		Type:    "db.m3.medium",
	}

	cmd_test := &Command{
		OutConfig: out,
		RDSConfig: rds,
		RDSClient: aws_rds,
		ARNPrefix: "arn:aws:rds:" + test_region + ":" + "123456789" + ":",
	}

	return server, cmd_test
}

func TestDescribeDBInstances(t *testing.T) {
	ts, tc := getTestClient(200, sr_DescribeDBInstancesResponse)
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
	ts, tc := getTestClient(200, sr_DescribeDBInstanceResponse)
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
	ts, tc := getTestClient(200, sr_ListTagsForResourceResponse)
	defer ts.Close()

	db_id := "rds-try-test"
	var ti rds.DBInstance
	ti.DBInstanceIdentifier = &db_id

	ri, err := tc.checkListTagsForResource(&ti)

	if err != nil {
		t.Errorf("[checkListTagsForResource] result error: %s", err.Error())
	}
	if !ri {
		t.Error("tags count not match")
	}
}

func TestModifyDBInstance(t *testing.T) {
	ts, tc := getTestClient(200, sr_DescribeDBInstanceResponse)
	defer ts.Close()

	id := "rds-try-test-db-1"
	ri, err := tc.DescribeDBInstance(id)

	tsm, tcm := getTestClient(200, sr_ModifyDBInstanceResponse)
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
	ts, tc := getTestClient(200, sr_RebootDBInstanceResponse)
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
	ts, tc := getTestClient(200, sr_DescribeDBSnapshotsResponse)
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
	ts, tc := getTestClient(200, sr_DescribeDBSnapshotResponse)
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
	ts, tc := getTestClient(200, sr_DescribeDBSnapshotResponse)
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
	ts, tc := getTestClient(200, sr_DeleteDBInstanceResponse)
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
	ts, tc := getTestClient(200, sr_CreateDBSnapshotResponse)
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
	ts, tc := getTestClient(200, sr_DeleteDBSnapshotResponse)
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
	ts, tc := getTestClient(200, sr_DescribeDBInstanceResponse)
	defer ts.Close()

	id := "rds-try-test-db-1"
	ri, _ := tc.DescribeDBInstance(id)
	state := tc.CheckPendingStatus(ri)

	if state {
		t.Error("CheckPendingStatus not match")
	}
}

func TestRestoreDBInstanceFromDBSnapshot(t *testing.T) {
	ts, tc := getTestClient(200, sr_DescribeDBInstanceResponse)
	defer ts.Close()

	id := "rds-try-test-db-1"
	ri, _ := tc.DescribeDBInstance(id)

	tss, tcs := getTestClient(200, sr_DescribeDBSnapshotResponse)
	defer tss.Close()

	ids := "before-test-1"
	ris, _ := tcs.DescribeDBSnapshot(ids)

	tsr, tcr := getTestClient(200, sr_RestoreDBInstanceFromDBSnapshotResponse)
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
		case rt_name_text, rt_time_text:
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
	test_name := utils.GetAppName() + "-test"
	temp_dir, _ := ioutil.TempDir("", test_name)

	temp_file, err := ioutil.TempFile(temp_dir, utils.GetAppName()+"-test")
	if err != nil {
		t.Errorf("failed to create the temp file: %s", err.Error())
	}
	f_stat, _ := temp_file.Stat()
	f_name := temp_file.Name()

	temp_file.Sync()
	temp_file.Close()
	defer os.RemoveAll(temp_dir)

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
		FileName: f_stat.Name(),
		Path:     temp_dir,
		Bom:      false,
	}

	csv_s := writeCSVFile(args)

	file, err := os.OpenFile(f_name, os.O_RDONLY, 0777)
	defer file.Close()

	f_stat, _ = file.Stat()

	if !csv_s {
		t.Errorf("[writeCSVFile] result error: %s", err.Error())
	}
	if f_stat.Size() <= 0 {
		t.Errorf("csv file not out put: %d", f_stat.Size())
	}
}

// It takes 30 seconds every time
func TestWaitForStatusAvailable(t *testing.T) {
	ts, tc := getTestClient(200, sr_DescribeDBInstanceResponse)
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
