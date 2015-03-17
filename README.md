rds-try
===============
[![Build Status](https://travis-ci.org/uchimanajet7/rds-try.svg?branch=master)](https://travis-ci.org/uchimanajet7/rds-try)

rds-try can perform the following operations against  [Amazon RDS](
http://aws.amazon.com/jp/rds/)

##### for MySQL
- Instance operation
 - Instance creation from snapshot
    - Create a new snapshot
 - List of the created instance
    - List of the created snapshot
 - Remove created instance
    - Remove created snapshot
- Query execution
 - Query execution time measurement
 - Query results file output

##### for Others
- `Not Supported（2015/03/16）`
 - for PostgreSQL
 - for Oracle
 - for SQL Server
 - for Aurora

Create a new RDS instance from the latest snapshot and Run the specified SQL against an instance that is created
Executed SQL measurement execution time, possible to save the execution results as csv file

##Usage

[Download here](https://github.com/uchimanajet7/rds-try/releases) and Run in the downloaded location where the binary file your path
The path of the file reading and writing will be the home directory of the running OS user if there is no specified

```ini
Usage: rds-try [globals] <command> [options]

Globals:
  -h, --help     show this help message and exit
  -v, --version  show version message and exit
  -c, --config   specify an alternate config file
  -n, --name     specify an alternate rds environment name

Commands:
  es  restore db and get results by execute sql
  ls  list up own db instances and snapshots
  rm  delete your created db instances and snapshots

Options:
  show commands options help <command> -h, --help
```
==The path of the file reading and writing will be the home directory of the running OS user if there is no specified==

**Globals**

| Name | Description |
|--------|--------|
|-h, --help |show help message and exit|
|-v, --version |show version message and exit|
|-c, --config |specify an alternate config file|
|-n, --name |specify an alternate rds environment name|

**Commands**

| Name | Description |
|--------|--------|
|es |restore DB from a snapshot and run the SQL against DB|
|ls |show a list of the DB instance and snapshot that created in this tool|
|rm |remove all the DB instance and snapshot that created with this tool|

_ _ _

##### Command usage: es
```ini
Usage: rds-try es [options]

Options:
  -q, --query  specify an alternate query file
  -s, --snap   create snapshot before restore
  -t, --type   specify an alternate db instance class
```

**Options**

| Name | Description |
|--------|--------|
|-q, --query |specifies the query file to be executed|
|-s, --snap |create snapshot before restore|
|-t, --type |specifies [DB Instance Classes](http://aws.amazon.com/rds/details/#DB_Instance_Classes) |

_ _ _
##### Command usage: ls
```ini
Usage: rds-try ls [options]

Options:
  -s, --snap  include own db snapshots to list
```

**Options**

| Name | Description |
|--------|--------|
|-s, --snap |include snapshots to list|

_ _ _
##### Command usage: rm
```ini
Usage: rds-try rm [options]

Options:
  -s, --snap   include own db snapshots to delete
  -f, --force  forced delete without confirmation
```

**Options**

| Name | Description |
|--------|--------|
|-s, --snap |include snapshot to delete|
|-f, --force |forced delete without confirmation|

##Use API and Authority
Calling the following AWS API by using [awslabs/aws-sdk-go](https://github.com/awslabs/aws-sdk-go)
Please set the appropriate permissions on the AWS IAM user

- RDS
 - CreateDBSnapshot
 - DeleteDBInstance
 - DeleteDBSnapshot
 - DescribeDBInstances
 - DescribeDBSnapshots
 - ListTagsForResource
 - ModifyDBInstance
 - RebootDBInstance
 - RestoreDBInstanceFromDBSnapshot
- IAM
 - ListUsers

== The following is an example. Permission settings, please be sure to check in their own ==

##### Setting example: IAM policy 
```ini
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iam:ListUsers"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "rds:CreateDBSnapshot",
        "rds:DeleteDBInstance",
        "rds:DeleteDBSnapshot",
        "rds:DescribeDBInstances",
        "rds:DescribeDBSnapshots",
        "rds:ListTagsForResource",
        "rds:ModifyDBInstance",
        "rds:RebootDBInstance",
        "rds:RestoreDBInstanceFromDBSnapshot"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
```
Use the `ListUsers` of IAM API to get IAM user's Amazon Resource Name (ARN)
If you can get ARN only RDS API, Not necessary to use `ListUsers`

In addition, for writing to ** log file ** and ** query result file **, Please specify location where you run OS user has write permissions

##### (Reference) AWS API documentation
- Welcome - Amazon Relational Database Service
 - http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/Welcome.html
- Welcome - AWS Identity and Access Management
 - http://docs.aws.amazon.com/IAM/latest/APIReference/Welcome.html
- Tagging Amazon RDS Resources - Amazon Relational Database Service
 - http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html#USER_Tagging.ARN

##Limitation
`2015/03/16` There are the following limitations
- Executed only for ** RDS for MySQL **
- Need ** RDS instance running ** To run, It became snapshot of original
 - Because you have to get and set configuration information from instance of running
- ** Write permission is required ** because it can not be the output of the log file to OFF
- Has not been confirmed to work only in the following ** three types of OS **
 - Windows 8.1 x64
 - OS X Yosemite (10.10) x64
 - CentOS release 6.5 (Final) x64
- Exit ** without retry ** if an error occurred in API
- Not a mechanism, such as to ** notify the end of execution **
- The contents of Tag are you to determine whether or not made a DB instance and snapshot with this tool. There is likely to be mistaken with a modified or coincidence
- Even if forced to stop in the middle of the execution of the command, it does not roll back. So DB instance and snapshot created will be as it is. Please be addressed individually as needed

##Config file
Described using the [toml-lang/toml](https://github.com/toml-lang/toml) format
Description example, please refer to the following and `rds-try.conf.example` file

##### Description example: rds-try.conf

```ini
# set aws access_key and secret_key
[aws]
access_key = "your AWS_ACCESS_KEY_ID"
secret_key = "your AWS_SECRET_ACCESS_KEY"

# set out put environment informations
[out]
root = "/home/awsuser"
file = true
bom = false

# set log environment informations
[log]
root = "/home/awsuser/log"
verbose = false
json = true

# set rds environment informations
[rds.default]
multi_az = false
db_id = "your DB Instance Identifier"
region = "your Region"
user = "rdstestuser"
pass = "redsmysqlpass"
type = "db.m3.medium"
```
**aws**

| Name | Type | Description |
|--------|--------|--------|
| access_key | String | Specifies the access key of the AWS user |
| secret_key | String | Specifies the secret access key of AWS user |

- `Optional`
- aws - GoDoc
 - http://godoc.org/github.com/awslabs/aws-sdk-go/aws#DetectCreds
- Because internally utilizes the function, is performed by using the environmental variables and IAM Role If not specified

**out**

| Name | Type | Description |
|--------|--------|--------|
| root | String |Specifies the root directory to output the file.<br> It is run OS user's home directory if it is not specified |
| file | Boolean | Specify true if you want to output the query results to a file.<br> It is false if not specified|
| bom | Boolean | Specify true if you want to add a [BOM](http://en.wikipedia.org/wiki/Byte_order_mark) in a file that is output in UTF-8.<br> It is false if not specified|

- `Optional`
- The values in the description each is adopted if it is not specified

**log**

| Name | Type | Description |
|--------|--------|--------|
| root | String |Specifies the root directory to output a log file.<br> It is run OS user's home directory if it is not specified |
| verbose | Boolean |Specify true if you want to and detailed display the output.<br> It is false if not specified |
| json | Boolean |Specify true if you want to output to a log file in JSON format.<br> It is false if not specified |

- `Optional`
- The values in the description each is adopted if it is not specified

**rds.* **

| Name | Type | Description |
|--------|--------|--------|
| multi_az | Boolean |Specify true to MultiAZ placement.<br> It is false if not specified |
| db_id | String | ==Required==<br> Specifies the identifier name of DB |
| region | String | ==Required==<br> Specifies the AWS Region of DB |
| user | String | ==Required==<br> Specifies the user name for connecting to the DB |
| pass | String | ==Required==<br> Specify the password to connect to the DB |
| type | String | specifies [DB Instance Classes](http://aws.amazon.com/rds/details/#DB_Instance_Classes)<br> If not specified, the same DB Instance Classes and DB in start-up is adopted.<br> Arguments side has priority when there is specified by the argument |

- ==Required item==
- It is an error if this item can not be read
- Can specify the part following the **rds. ** in the argument global options
- If the group name is not the same I can describe multiple items
 - For example [rds.test1],[rds.test2],[rds.test3]...etc
- ** ”default” ** of the group name is the default value if there is no argument specified
- ** ”type” ** is subject to the same ** "DB Instance Classes" ** as the DB in start-up if there is no specified. Arguments side has priority when there is specified by the argument

##Query file
Described using the [toml-lang/toml](https://github.com/toml-lang/toml) format
Description example, please refer to the following and `rds-try.query.example` file

##### Description example: rds-try.query

```ini
[[query]]
name = "selectDB"
sql = "USE RDSTESTDB"

[[query]]
name = "selectID"
sql = "SELECT id FROM account"
```

**query**

| Name | Type | Description |
|--------|--------|--------|
| name | String | ==Required==<br> Specifies the display name of the query.<br> Is used as the file name in the case of outputting the file |
| sql | String | ==Required==<br> Specifies the SQL string to be executed |

- ==Required item==
- There is a need to fill in ** [[query]] ** format
- Are executed in the order in which they are entered

##Execution example

##### Command: es

![es](https://raw.github.com/wiki/uchimanajet7/rds-try/images/cmd_es.gif)

```ini
[awsuser@ip-10-0-0-5 ~]$./rds-try es -s
INFO[0000] start command : es                            module=command
INFO[0031] DB Snapshot Status: creating                  module=command
・・・
INFO[0121] Status: available                             module=command
INFO[0122] {DBInstanceClass:db.t1.micro DBIdentifier:rds-try-v0-0-1-2015-02-25-10-37-12-test-db MultiAZ:false Snapshot:0xc20828ec30 Instance:0xc20820a000}  module=command
INFO[0152] DB Instance Status: creating                  module=command
・・・
INFO[0723] Status: available                             module=command
INFO[0723] query start time: 2015-02-25 10:47:14.631464535 +0900 JST  module=command
INFO[0723] query end time: 2015-02-25 10:47:14.669333642 +0900 JST  module=command
INFO[0723] query start time: 2015-02-25 10:47:14.669521571 +0900 JST  module=command
INFO[0740] query end time: 2015-02-25 10:47:31.685896983 +0900 JST  module=command
INFO[0753] query start time: 2015-02-25 10:47:44.924241467 +0900 JST  module=command
INFO[0762] query end time: 2015-02-25 10:47:53.844155869 +0900 JST  module=command
INFO[0777] query start time: 2015-02-25 10:48:08.402075463 +0900 JST  module=command
INFO[0782] query end time: 2015-02-25 10:48:13.734800522 +0900 JST  module=command
INFO[0794] query start time: 2015-02-25 10:48:25.47375115 +0900 JST  module=command
INFO[0795] query end time: 2015-02-25 10:48:26.719605144 +0900 JST  module=command

runtime result:
  query name   : selectDB
  query runtime: 37.869107ms

  query name   : selectID
  query runtime: 17.016375412s

  query name   : selectName
  query runtime: 8.919914402s

  query name   : selectMemberID
  query runtime: 5.332725059s

  query name   : selectAge
  query runtime: 1.245853994s

--------------------------------
  total runtime: 32.553 sec

INFO[0795] end command : es                              module=command
[awsuser@ip-10-0-0-5 ~]$
```

##### Command: ls

![ls](https://raw.github.com/wiki/uchimanajet7/rds-try/images/cmd_ls.gif)

```ini
[awsuser@ip-10-0-0-5 ~]$./rds-try ls -s 
INFO[0000] start command : ls                            module=command

your create db instance list by tag name
  [ 1] DB Instance: rds-try-v0-0-1-2015-02-24-20-13-22-test-db
  [ 2] DB Instance: rds-try-v0-0-1-2015-02-24-20-38-15-test-db

your create db snapshot list by tag name
  [ 1] DB Snapshot: rds-try-v0-0-1-2015-02-24-20-11-21-test-db
  [ 2] DB Snapshot: rds-try-v0-0-1-2015-02-24-20-35-44-test-db
  [ 3] DB Snapshot: rds-try-v0-0-1-2015-02-24-21-01-45-test-db

INFO[0001] end command : ls                              module=command
[awsuser@ip-10-0-0-5 ~]$
```

##### Command: rm

![rm](https://raw.github.com/wiki/uchimanajet7/rds-try/images/cmd_rm.gif)

```ini
[awsuser@ip-10-0-0-5 ~]$./rds-try rm -s 
INFO[0000] start command : rm                            module=command

your create db instance list by tag name
  [ 1] DB Instance: rds-try-v0-0-1-2015-02-24-20-13-22-test-db
  [ 2] DB Instance: rds-try-v0-0-1-2015-02-24-20-38-15-test-db

your create db snapshot list by tag name
  [ 1] DB Snapshot: rds-try-v0-0-1-2015-02-24-20-11-21-test-db
  [ 2] DB Snapshot: rds-try-v0-0-1-2015-02-24-20-35-44-test-db
  [ 3] DB Snapshot: rds-try-v0-0-1-2015-02-24-21-01-45-test-db

may be delete all? [y/n]: y

INFO[0011] [ 1] deleted DB Instance: rds-try-v0-0-1-2015-02-24-20-13-22-test-db  module=command
INFO[0011] [ 2] deleted DB Instance: rds-try-v0-0-1-2015-02-24-20-38-15-test-db  module=command
INFO[0011] [ 1] deleted DB Snapshot: rds-try-v0-0-1-2015-02-24-20-11-21-test-db  module=command
INFO[0012] [ 2] deleted DB Snapshot: rds-try-v0-0-1-2015-02-24-20-35-44-test-db  module=command
INFO[0012] [ 3] deleted DB Snapshot: rds-try-v0-0-1-2015-02-24-21-01-45-test-db  module=command
INFO[0012] end command : rm                              module=command
[awsuser@ip-10-0-0-5 ~]$
```
