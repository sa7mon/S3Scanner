# S3Scanner

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
  
