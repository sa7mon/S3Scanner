# S3Scanner
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Build Status](https://travis-ci.org/sa7mon/S3Scanner.svg?branch=master)](https://travis-ci.org/sa7mon/S3Scanner)

A tool to find open S3 buckets and dump their contents💧

<img src="https://user-images.githubusercontent.com/3712226/115632654-d4f8c280-a2cd-11eb-87ee-c70bbd4f1edb.png" width="85%"/>

## Usage
<pre>
usage: s3scanner [-h] [--version] [--threads n] [--endpoint-url ENDPOINT_URL] [--endpoint-address-style {path,vhost}] [--insecure] {scan,dump} ...

s3scanner: Audit unsecured S3 buckets
           by Dan Salmon - github.com/sa7mon, @bltjetpack

optional arguments:
  -h, --help            show this help message and exit
  --version             Display the current version of this tool
  --threads n, -t n     Number of threads to use. Default: 4
  --endpoint-url ENDPOINT_URL, -u ENDPOINT_URL
                        URL of S3-compliant API. Default: https://s3.amazonaws.com
  --endpoint-address-style {path,vhost}, -s {path,vhost}
                        Address style to use for the endpoint. Default: path
  --insecure, -i        Do not verify SSL

mode:
  {scan,dump}           (Must choose one)
    scan                Scan bucket permissions
    dump                Dump the contents of buckets
</pre>

## 🚀 Support
If you've found this tool useful, please consider donating to support its development. You can find sponsor options on the side of this repo page or in [FUNDING.yml](.github/FUNDING.yml)

<div align="center"><a href="https://www.tines.com/?utm_source=oss&utm_medium=sponsorship&utm_campaign=s3scanner"><img src="https://user-images.githubusercontent.com/3712226/146481766-a331b010-29c4-4537-ac30-9a4b4aad06b3.png" height=50 width=140></a></div>

<p align="center">Huge thank you to <a href="https://www.tines.com/?utm_source=oss&utm_medium=sponsorship&utm_campaign=s3scanner">tines</a> for being an ongoing sponsor of this project.</p>
           

## Installation

```shell
pip3 install s3scanner
```

or via Docker:

```shell
docker build . -t s3scanner:latest
docker run --rm s3scanner:latest scan --bucket my-buket
```

or from source:

```shell
git clone git@github.com:sa7mon/S3Scanner.git
cd S3Scanner
pip3 install -r requirements.txt
python3 -m S3Scanner
```

## Features

* ⚡️ Multi-threaded scanning
* 🔭 Supports tons of S3-compatible APIs
* 🕵️‍♀️ Scans all bucket permissions to find misconfigurations
* 💾 Dump bucket contents to a local folder
* 🐳 Docker support

## Examples

* Scan AWS buckets listed in a file with 8 threads
  ```shell
  $ s3scanner --threads 8 scan --buckets-file ./bucket-names.txt
   ```
* Scan a bucket in Digital Ocean Spaces 
  ```shell
  $ s3scanner --endpoint-url https://sfo2.digitaloceanspaces.com scan --bucket my-bucket
  ```
* Dump a single AWS bucket
  ```shell
  $ s3scanner dump --bucket my-bucket-to-dump
  ```
* Scan a single Dreamhost Objects bucket which uses the vhost address style and an invalid SSL cert
  ```shell
  $ s3scanner --endpoint-url https://objects.dreamhost.com --endpoint-address-style vhost --insecure scan --bucket my-bucket
  ```

## S3-compatible APIs

`S3Scanner` can scan and dump buckets in S3-compatible APIs services other than AWS by using the
`--endpoint-url` argument. Depending on the service, you may also need the `--endpoint-address-style`
or `--insecure` arguments as well. 

Some services have different endpoints corresponding to different regions

**Note:** `S3Scanner` currently only supports scanning for anonymous user permissions of non-AWS services

| Service | Example Endpoint | Address Style | Insecure ? |
|---------|------------------|:-------------:|:----------:|
| DigitalOcean Spaces (SFO2 region) | https://sfo2.digitaloceanspaces.com | path | No |  
| Dreamhost | https://objects.dreamhost.com | vhost | Yes |
| Linode Object Storage (eu-central-1 region) | https://eu-central-1.linodeobjects.com | vhost | No |
| Scaleway Object Storage (nl-ams region) | https://s3.nl-ams.scw.cloud | path | No |
| Wasabi Cloud Storage | http://s3.wasabisys.com/ | path | Yes |

📚 Current status of non-AWS APIs can be found [in the project wiki](https://github.com/sa7mon/S3Scanner/wiki/S3-Compatible-APIs)

## Interpreting Results

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

## Contributors
* [Ohelig](https://github.com/Ohelig)
* [vysecurity](https://github.com/vysecurity)
* [janmasarik](https://github.com/janmasarik)
* [alanyee](https://github.com/alanyee)
* [klau5dev](https://github.com/klau5dev)
* [hipotermia](https://github.com/hipotermia)

## License

MIT