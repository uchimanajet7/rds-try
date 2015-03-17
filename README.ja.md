rds-try
===============

[![Build Status](https://travis-ci.org/uchimanajet7/rds-try.svg?branch=master)](https://travis-ci.org/uchimanajet7/rds-try)

rds-tryは [Amazon RDS](
http://aws.amazon.com/jp/rds/) に対して以下の操作を行えます

##### for MySQL
- インスタンス操作
 - スナップショットからのインスタンス作成
    - 新規スナップショットの作成
 - 作成したインスタンスの一覧表示
    - 作成したスナップショットの一覧表示
 - 作成したインスタンスの削除
    - 作成したスナップショットの削除
- クエリー実行
 - クエリー実行時間計測
 - クエリー結果ファイル出力

##### for Others
- `未対応（2015/03/16 現在）`
 - for PostgreSQL
 - for Oracle
 - for SQL Server
 - for Aurora

最新スナップショットからRDS 新規インスタンスを作成し、作成されたインスタンスに対して指定されたSQLを実行します
実行されたSQLは実行時間が計測され、実行結果をcsvファイルとして保存することが出来ます

##使用法

[ここ](https://github.com/uchimanajet7/rds-try/releases) からダウンロードしたバイナリファイルをパスの通った場所において実行してください。ファイル読み書きのパスは指定がなければ実行OSユーザーのホームディレクトリになります

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
==ファイルの読み書きは指定がなければユーザーのホームディレクトリが対象となります==

**グローバルオプション**

| 名称 | 説明 |
|--------|--------|
|-h, --help |ヘルプを表示します|
|-v, --version |バージョンを表示します|
|-c, --config |コンフィグファイルを指定します|
|-n, --name |コンフィグファイル中の利用するRDS変数グループ名を指定します|

**コマンド**

| 名称 | 説明 |
|--------|--------|
|es |スナップショットからRDSを復元しクエリを実行します|
|ls |このツールで作成したRDSインスタンス一覧を表示します|
|rm |このツールで作成したRDSインスタンスをすべて削除します|

_ _ _

##### es コマンド使用法
```ini
Usage: rds-try es [options]

Options:
  -q, --query  specify an alternate query file
  -s, --snap   create snapshot before restore
  -t, --type   specify an alternate db instance class
```

**オプション**

| 名称 | 説明 |
|--------|--------|
|-q, --query |実行するクエリファイルを指定します|
|-s, --snap |スナップショットを作成してから実行します|
|-t, --type |復元するRDSの [RDSインスタンスクラス](http://aws.amazon.com/jp/rds/details/#DB_インスタンスクラス) を指定します|

_ _ _
##### ls コマンド使用法
```ini
Usage: rds-try ls [options]

Options:
  -s, --snap  include own db snapshots to list
```

**オプション**

| 名称 | 説明 |
|--------|--------|
|-s, --snap |スナップショットも一覧表示の対象にします|

_ _ _
##### rm コマンド使用法
```ini
Usage: rds-try rm [options]

Options:
  -s, --snap   include own db snapshots to delete
  -f, --force  forced delete without confirmation
```

**オプション**

| 名称 | 説明 |
|--------|--------|
|-s, --snap |スナップショットも削除対象にします|
|-f, --force |確認を行わずに削除を実行します|

##利用APIと権限
[awslabs/aws-sdk-go](https://github.com/awslabs/aws-sdk-go) を利用して以下のAPIを呼び出していますので、これを参考にAWSのIAMユーザーに適切な権限を設定してください

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

== 以下は一例となります。権限設定は必ず自身での確認をお願いします ==

##### IAMポリシー設定例
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

IAMの`ListUsers`は、実行AWSユーザーのAmazon Resource Name（ARN）を取得するために利用しています。RDS関連APIのみでARNを取得可能な方法が見つかれば不要となります

また、**ログファイル**と**クエリー結果ファイル**に書き込みを行いますので、実行OSユーザーが**書き込み権限**を持っている場所を指定してください


##### （参考）AWS API ドキュメント
- Welcome - Amazon Relational Database Service
 - http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/Welcome.html
- Welcome - AWS Identity and Access Management
 - http://docs.aws.amazon.com/IAM/latest/APIReference/Welcome.html
- Tagging Amazon RDS Resources - Amazon Relational Database Service
 - http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html#USER_Tagging.ARN

##制限事項
`2015/03/16 現在` 以下の制限事項があります
- 対応しているのは**RDS for MySQL**のみです
- 実行にはスナップショット元の**動作している**RDSインスタンスが必要です
 - 設定情報を動作中のインスタンスから取得して設定しているためです
- ログファイルの出力をOFFに出来ないため**書き込み権限**が必要です
- 動作確認を行ったOSは以下の**3種**のみです
 - Windows 8.1 64ビット版
 - OS X Yosemite (10.10)
 - CentOS release 6.5 (Final) 64ビット版
- API でエラーが起こった場合**再試行**せずに終了します
- 実行終了を**通知**するような仕組みはありません
- DBスナップショット・DBインスタンスをこのツールで作ったかどうか判別しているのは**Tagの内容**なので、改変や偶然の一致で誤認する可能性があります
- コマンドの実行を途中で強制的に止めても、ロールバックしないので作成されたDBスナップショットやDBインスタンスはそのままとなります。必要に応じて個別で対処してください

##コンフィグファイル
[toml-lang/toml](https://github.com/toml-lang/toml) フォーマットを使って記述します
記述例は `rds-try.conf.example` ファイルと以下を参照してください

##### rds-try.conf 記述例

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

| 名称 | 型 | 説明 |
|--------|--------|--------|
| access_key | 文字列 | AWSユーザーのアクセスキーを指定します |
| secret_key | 文字列 | AWSユーザーのシークレットアクセスキーを指定します |

- `省略可能`
- aws - GoDoc
 - http://godoc.org/github.com/awslabs/aws-sdk-go/aws#DetectCreds
- 内部では上記関数を利用しているので、指定がない場合は環境変数やIAM Roleを利用して実行されます

**out**

| 名称 | 型 | 説明 |
|--------|--------|--------|
| root | 文字列 |ファイルを出力するrootディレクトを指定します。<br> 指定がない場合は実行OSユーザーのホームディレクトリとなります |
| file | Boolean | クエリー結果をファイルに出力する場合はtrueを指定します。<br> 指定がない場合はfalseとなります|
| bom | Boolean | UTF-8で出力されるファイルに [BOM](http://ja.wikipedia.org/wiki/%E3%83%90%E3%82%A4%E3%83%88%E3%82%AA%E3%83%BC%E3%83%80%E3%83%BC%E3%83%9E%E3%83%BC%E3%82%AF) を付加する場合はtrueを指定します。<br> 指定がない場合はfalseとなります|

- `省略可能`
- 指定がない場合はそれぞれ説明中の値が採用されます

**log**

| 名称 | 型 | 説明 |
|--------|--------|--------|
| root | 文字列 |ログファイルを出力するrootディレクトを指定します。<br> 指定がない場合は実行OSユーザーのホームディレクトリとなります |
| verbose | Boolean |ログファイルと画面への出力を詳細表示とする場合はtrueを指定します。<br> 指定がない場合はfalseとなります |
| json | Boolean |ログファイルへの出力をJSON形式にする場合はtrueを指定します。<br> 指定がない場合はfalseとなります |

- `省略可能`
- 指定がない場合はそれぞれ説明中の値が採用されます

**rds.* **

| 名称 | 型 | 説明 |
|--------|--------|--------|
| multi_az | Boolean |復元するDBをMultiAZ配置にする場合はtrueを指定します。指定がない場合はfalseとなります |
| db_id | 文字列 | ==必須==<br> 復元するDBの識別子名を指定します |
| region | 文字列 | ==必須==<br> 復元するDBのAWSリージョンを指定します |
| user | 文字列 | ==必須==<br> 復元するDBへ接続するためのユーザー名を指定します |
| pass | 文字列 | ==必須==<br> 復元するDBへ接続するためのパスワードを指定します |
| type | 文字列 | 復元するRDSの [RDSインスタンスクラス](http://aws.amazon.com/jp/rds/details/#DB_インスタンスクラス) を指定します。<br> 指定がない場合は起動中のスナップショット元DBと同じインスタンスクラスが採用されます。<br> 引数で指定があった場合は引数側が優先されます |

- ==必須項目==
- この項目が読み込めない場合はエラーとなります
- **rds. ** に続く部分を引数のグローバルオプションで指定が出来ます
- このグループ名が同一でなければ複数の項目を記述できます
 - [rds.test1],[rds.test2],[rds.test3]...etc
- グループ名の**default**は引数指定がない場合の**規定値**になります
- type は指定がなければ起動中のRDSと同じインスタンスクラスが適用されます。引数で指定があった場合は引数が最優先となります


##クエリーファイル
[toml-lang/toml](https://github.com/toml-lang/toml) フォーマットを使って記述します
記述例は `rds-try.query.example` ファイルと以下を参照してください

##### rds-try.query 記述例

```ini
[[query]]
name = "selectDB"
sql = "USE RDSTESTDB"

[[query]]
name = "selectID"
sql = "SELECT id FROM account"
```

**query**

| 名称 | 型 | 説明 |
|--------|--------|--------|
| name | 文字列 | ==必須==<br> クエリーの表示名称を指定します。<br> ファイルを出力する場合にはファイル名として使用します |
| sql | 文字列 | ==必須==<br> 実行するSQL文字列を指定します |

- ==必須項目==
- ** [[query]] ** の書式で記入する必要があります
- 記入されている順番で実行されます

##実行例

##### es コマンド

![es コマンド](https://raw.github.com/wiki/uchimanajet7/rds-try/images/cmd_es.gif)

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

##### ls コマンド

![ls コマンド](https://raw.github.com/wiki/uchimanajet7/rds-try/images/cmd_ls.gif)

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

##### rm コマンド

![rm コマンド](https://raw.github.com/wiki/uchimanajet7/rds-try/images/cmd_rm.gif)

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
