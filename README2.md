# S3Scanner
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Build Status](https://travis-ci.org/sa7mon/S3Scanner.svg?branch=master)](https://travis-ci.org/sa7mon/S3Scanner)

A tool to find open S3 buckets and dump their contents💧

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

## Support
🚀 If you've found this tool useful, please consider donating to support its development

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=XG5BGLQZPJ9H8)

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/B0B54D93O)

## Installation

## Examples



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



## Tested non-AWS endpoints

* DigitalOcean Spaces Object Storage
  ```
  python3 scanner.py --endpoint-url https://sfo2.digitaloceanspaces.com scan --bucket my-bucket
  ```
* Linode Object Storage
  ```
   python3 scanner.py --endpoint-url https://eu-central-1.linodeobjects.com --endpoint-address-style vhost scan --bucket my-bucket
  ```
* Wasabi S3-compatible Cloud Storage
  ```
  python3 scanner.py --endpoint-url http://s3.wasabisys.com/ --insecure scan --bucket my-bucket
  ```
* Dreamhost Objects
  ```
  python3 scanner.py --endpoint-url https://objects.dreamhost.com --endpoint-address-style vhost --insecure scan --bucket my-bucket
  ```
* Scaleway Object Storage
  ```
  python3 scanner.py --endpoint-url https://s3.nl-ams.scw.cloud scan --bucket my-bucket
  ```
  
