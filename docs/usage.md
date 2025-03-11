---
weight: -1
---

# Usage

```
INPUT: (1 required)
-bucket        string  Name of bucket to check.
-bucket-file   string  File of bucket names to check.
-mq                    Connect to RabbitMQ to get buckets. Requires config file key "mq". Default: "false"

OUTPUT:
-db       Save results to a Postgres database. Requires config file key "db.uri". Default: "false"
-json     Print logs to stdout in JSON format instead of human-readable. Default: "false"

OPTIONS:
-enumerate           Enumerate bucket objects (can be time-consuming). Default: "false"
-provider    string  Object storage provider: aws, custom, digitalocean, dreamhost, gcp, linode, scaleway - custom requires config file. Default: "aws"
-threads     int     Number of threads to scan with. Default: "4"

DEBUG:
-verbose     Enable verbose logging. Default: "false"
-version     Print version Default: "false"

If config file is required these locations will be searched for config.yml: "." "/etc/s3scanner/" "$HOME/.s3scanner/"
```

## Input

`s3scanner` requires exactly one type of input: `-bucket`, `-bucket-file`, or `-mq`.

```
INPUT: (1 required)
  -bucket        string  Name of bucket to check.
  -bucket-file   string  File of bucket names to check.
  -mq                    Connect to RabbitMQ to get buckets. Requires config file key "mq". Default: "false"
```

*`-bucket`*
------------

Scan a single bucket

```shell
s3scanner -bucket secret_uploads
```

*`-bucket-file`*
----------------
Scans every bucket name listed in file

```
s3scanner -bucket-file names.txt
```
where `names.txt` contains one bucket name per line

```
$ cat names.txt
bucket123
assets
image-uploads
```

Bucket names listed multiple times will only be scanned once.

*`-mq`*
-------

Connects to a RabbitMQ server and consumes messages containing bucket names to scan.

```
s3scanner -mq
```

Messages should be JSON-encoded [`Bucket`](https://github.com/sa7mon/s3scanner/blob/main/bucket/bucket.go) objects - refer to [`mqingest`](https://github.com/sa7mon/s3scanner/blob/main/cmd/mqingest/mqingest.go) for a Golang publishing example.

`-mq` requires the `mq.uri` and `mq.queue_name` config file keys. See Config File section for example.

## Output

```
OUTPUT:
  -db       Save results to a Postgres database. Requires config file key "db.uri". Default: "false"
  -json     Print logs to stdout in JSON format instead of human-readable. Default: "false"
```

*`-db`*
----------

Saves all scan results to a PostgreSQL database

```shell
s3scanner -bucket images -db
```

* Requires the `db.uri` config file key. See Config File section for example.
* If using `-db`, results will also be printed to the console if using `-json` or the default human-readable output mode.
* `s3scanner` runs Gorm's [Auto Migration](https://gorm.io/docs/migration.html#Auto-Migration) feature each time it connects two the database. If
  the schema already has tables with names Gorm expects, it may change these tables' structure. It is recommended to create a Postgres schema dedicated to `s3scanner` results.

*`-json`*
----------

Instead of outputting scan results to console in human-readable format, output machine-readable JSON.

```shell
s3scanner -bucket images -json
```

This will print one JSON object per line to the console, which can then be piped to `jq` or other tools that accept JSON input.

**Example**: Print bucket name and region for all buckets that exist

```shell
$ s3scanner -bucket-file names.txt -json | jq -r '. | select(.bucket.exists==1) | [.bucket.name, .bucket.region] | join(" - ")'       
10000 - eu-west-1
10000.pizza - ap-southeast-1
images_staging - us-west-2
```

## Options

```
OPTIONS:
  -enumerate           Enumerate bucket objects (can be time-consuming). Default: "false"
  -provider    string  Object storage provider: aws, custom, digitalocean, dreamhost, gcp, linode, scaleway - custom requires config file. Default: "aws"
  -threads     int     Number of threads to scan with. Default: "4"
```

*`-enumerate`*
--------------

Enumerate all objects stored in bucket. By default, `s3scanner` will only check permissions of buckets.
```shell
s3scanner -bucket attachments -enumerate
```

* **Note:** This can take a long time if there are a large number of objects stored.
* When enumerating, `s3scanner` will request "pages" of 1,000 objects. If there are more than 5,000 pages of objects, it will skip the rest.

*`-provider`*
-------------

Name of storage provider to use when checking buckets.

```shell
s3scanner -bucket assets -provider gcp
```

* Use "custom" when targeting a currently unsupported or local network storage provider.
* "custom" provider requires config file keys under `providers.custom` listed in the Config File section.

*`-threads`*
------------

Number of threads to scan with.

```shell
s3scanner -bucket secret_docs -threads 8
```

* Increasing threads will increase the number of buckets being scanned simultaneously, but will not speed up object enumeration. Enumeration is currently single-threaded per bucket.

## Debug

```
DEBUG:
  -verbose     Enable verbose logging. Default: "false"
  -version     Print version Default: "false"
```

*`-verbose`*
------------

Enables verbose logging of debug messages. This option will produce a lot of logs and is not recommended to use unless filing a bug report.

```shell
s3scanner -bucket spreadsheets -verbose
```

*`-version`*
------------

Print the version info and exit.

```shell
s3scanner -version
```

* Will print `dev` if compiled from source.