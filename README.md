<h1 align="center">
S3Scanner
</h1>

<p align="center">
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg"/></a>
<a href="https://github.com/sponsors/sa7mon/"><img src="https://img.shields.io/github/sponsors/sa7mon" /></a>
<a href="https://github.com/sa7mon/S3Scanner/issues"><img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat"/></a>
<a href="https://github.com/sa7mon/S3Scanner/releases/latest"><img src="https://img.shields.io/github/v/release/sa7mon/s3scanner" /></a>
</p>
<p align="center">
<a href="#features">Features</a> - <a href="#usage">Usage</a> - <a href="#quick-start">Quick Start</a> - <a href="#installation">Installation</a> - <a href="https://github.com/sa7mon/S3Scanner/discussions">Discuss</a> 
</p>
<br>
A tool to find open S3 buckets in AWS or other cloud providers:

- AWS
- DigitalOcean
- DreamHost
- GCP
- Linode
- Scaleway
- Custom

<img alt="demo" src="https://github.com/sa7mon/S3Scanner/assets/3712226/cfa16801-2a44-4ae9-ad85-9dd466390cd9">

# Features

* ⚡️ Multi-threaded scanning
* 🔭 Supports many built-in S3 storage providers or custom
* 🕵️‍♀️ Scans all bucket permissions to find misconfigurations
* 💾 Save results to Postgres database
* 🐇 Connect to RabbitMQ for automated scanning at scale
* 🐳 Docker support

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

# 🚀 Support
If you've found this tool useful, please consider donating to support its development. You can find sponsor options on the side of this repo page or in [FUNDING.yml](.github/FUNDING.yml)

<div align="center"><a href="https://www.tines.com/?utm_source=oss&utm_medium=sponsorship&utm_campaign=s3scanner"><img src="https://user-images.githubusercontent.com/3712226/146481766-a331b010-29c4-4537-ac30-9a4b4aad06b3.png" height=50 width=140></a></div>

<p align="center">Huge thank you to <a href="https://www.tines.com/?utm_source=oss&utm_medium=sponsorship&utm_campaign=s3scanner">tines</a> for being an ongoing sponsor of this project.</p>

# Quick Start

Scan AWS for bucket names listed in a file, enumerate all objects
  ```shell
  $ s3scanner -bucket-file names.txt -enumerate
   ```

Scan a bucket in GCP, enumerate all objects, and save results to database
  ```shell
  $ s3scanner -provider gcp -db -bucket my-bucket -enumerate
  ```

# Installation

| Platform                  | Version                                                                                                                                                      | Steps                                                                                      |
|---------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------|
| Homebrew (MacOS)          | [![homebrew version](https://img.shields.io/homebrew/v/s3scanner)](https://github.com/Homebrew/homebrew-core/blob/master/Formula/s/s3scanner.rb)             | `brew install s3scanner`                                                                   |
| Kali Linux                | [![Kali package](https://repology.org/badge/version-for-repo/kali_rolling/s3scanner.svg?header=Kali+Linux)](https://repology.org/project/s3scanner/versions) | `apt install s3scanner`                                                                    |
| Parrot OS                 | [![Parrot package](https://repology.org/badge/version-for-repo/parrot/s3scanner.svg?header=Parrot+OS)](https://repology.org/project/s3scanner/versions)      | `apt install s3scanner`                                                                    |
| Docker                    | ![Docker release](https://img.shields.io/github/v/release/sa7mon/s3scanner?label=Docker)                                                                     | `docker run ghcr.io/sa7mon/s3scanner`                                                      |
| Winget (Windows)          | [![Winget](https://repology.org/badge/version-for-repo/winget/s3scanner.svg?header=Winget)](https://repology.org/project/s3scanner/versions)                 | `winget install s3scanner`                                                                 |
| Go                        | ![Golang](https://img.shields.io/github/v/release/sa7mon/s3scanner?label=Go)                                                                                 | `go install -v github.com/sa7mon/s3scanner@latest`                                         |
| Other (Build from source) | ![GitHub release](https://img.shields.io/github/v/release/sa7mon/s3scanner?label=Git)                                                                        | `git clone git@github.com:sa7mon/S3Scanner.git && cd S3Scanner && go build -o s3scanner .` |

# Using

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

# Development

A docker compose file is included which creates 4 containers:

* rabbitmq
* postgres
* app
* mitm

2 profiles are configured:

- `dev` - Standard development environment
- `dev-mitm` - Environment configured with `mitmproxy` for easier observation of HTTP traffic when debugging or adding new providers.

To bring up the dev environment run `make dev` or `make dev-mitm`. Drop into the `app` container with `docker exec -it -w /app app_dev sh`, then `go run .`
If using the `dev-mitm` profile, open `http://127.0.0.1:8081` in a browser to view and manipulate HTTP calls being made from the app container.

# Config File

If using flags that require config options, `s3scanner` will search for `config.yml` in:
 
* (current directory)
* `/etc/s3scanner/`
* `$HOME/.s3scanner/`

```yaml
# Required by -db
db:
  uri: "postgresql://user:pass@db.host.name:5432/schema_name"

# Required by -mq
mq:
  queue_name: "aws"
  uri: "amqp://user:pass@localhost:5672"

# providers.custom required by `-provider custom`
#   address_style - Addressing style used by endpoints.
#     type: string
#     values: "path" or "vhost"
#   endpoint_format - Format of endpoint URLs. Should contain '$REGION' as placeholder for region name
#     type: string
#   insecure - Ignore SSL errors
#     type: boolean
# regions must contain at least one option
providers:
  custom: 
    address_style: "path"
    endpoint_format: "https://$REGION.vultrobjects.com"
    insecure: false
    regions:
      - "ewr1"
```

When `s3scanner` parses the config file, it will take the `endpoint_format` and replace `$REGION` for all `regions` listed to create a list of endpoint URLs.

# S3 compatible APIs

**Note:** `S3Scanner` currently only supports scanning for anonymous user permissions of non-AWS services

📚 More information on non-AWS APIs can be found [in the project wiki](https://github.com/sa7mon/S3Scanner/wiki/S3-Compatible-APIs).

## Permissions

This tool will attempt to get all available information about a bucket, but it's up to you to interpret the results.

[Possible permissions](https://docs.aws.amazon.com/AmazonS3/latest/user-guide/set-bucket-permissions.html) for buckets:

* Read - List and view all files
* Write - Write files to bucket
* Read ACP - Read all Access Control Policies attached to bucket
* Write ACP - Write Access Control Policies to bucket
* Full Control - All above permissions

Any or all of these permissions can be set for the 2 main user groups:
* Authenticated Users
* Public Users (those without AWS credentials set)
* Individual users/groups (out of scope of this tool)

**What this means:** Just because a bucket doesn't allow reading/writing ACLs doesn't mean you can't read/write files in the bucket. Conversely, you may be able to list ACLs but not read/write to the bucket

# License

MIT
