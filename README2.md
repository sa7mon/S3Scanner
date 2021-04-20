# S3Scanner
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Build Status](https://travis-ci.org/sa7mon/S3Scanner.svg?branch=master)](https://travis-ci.org/sa7mon/S3Scanner)

A tool to find open S3 buckets and dump their contents :droplet:

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
  
