# S3Scanner
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Build Status](https://travis-ci.org/sa7mon/S3Scanner.svg?branch=master)](https://travis-ci.org/sa7mon/S3Scanner)

A tool to find open S3 buckets and dump their contents :droplet:

![1 - s3finder.py](https://user-images.githubusercontent.com/3712226/40662408-e1d19468-631b-11e8-8d69-0075a6c8ab0d.png)

### If you've earned a bug bounty using this tool, please consider donating to support it's development

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=XG5BGLQZPJ9H8)


## Usage

<pre>
usage: s3scanner [-h] [-o OUTFILE] [-d] [-l] [--version] buckets

#  s3scanner - Find S3 buckets and dump!
#
#  Author: Dan Salmon - @bltjetpack, github.com/sa7mon

positional arguments:
  buckets               Name of text file containing buckets to check

optional arguments:
  -h, --help            show this help message and exit
  -o OUTFILE, --out-file OUTFILE
                        Name of file to save the successfully checked buckets in (Default: buckets.txt)
  -d, --dump            Dump all found open buckets locally
  -l, --list            Save bucket file listing to local file: ./list-buckets/${bucket}.txt
  --version             Display the current version of this tool
</pre>

The tool takes in a list of bucket names to check. Found S3 buckets are output to file. The tool will also dump or list the contents of 'open' buckets locally.

### Interpreting Results

This tool will attempt to get all available information about a bucket, but it's up to you to interpret the results.

[Settings available](https://docs.aws.amazon.com/AmazonS3/latest/user-guide/set-bucket-permissions.html) for buckets:
* Object Access (object in this case refers to files stored in the bucket)
  * List Objects
  * Write Objects
* ACL Access
  * Read Permissions
  * Write Permissions
  
Any or all of these permissions can be set for the 2 main user groups:
* Authenticated Users
* Public Users (those without AWS credentials set)
* (They can also be applied to specific users, but that's out of scope)
  
**What this means:** Just because a bucket returns "AccessDenied" for it's ACLs doesn't mean you can't read/write to it.
Conversely, you may be able to list ACLs but not read/write to the bucket


## Installation
  1. (Optional) `virtualenv venv && source ./venv/bin/activate`
  2. `pip install -r requirements.txt`
  3. `python ./s3scanner.py`

(Compatibility has been tested with Python 2.7 and 3.6)

### Using Docker

 1. Build the [Docker](https://docs.docker.com/) image:

 ```bash
sudo docker build -t s3scanner https://github.com/sa7mon/S3Scanner.git
```

 2. Run the Docker image:

 ```bash
sudo docker run -v /input-data-dir/:/data s3scanner --out-file /data/results.txt /data/names.txt
```
This command assumes that `names.txt` with domains to enumerate is in `/input-data-dir/` on host machine.

## Examples
This tool accepts the following type of bucket formats to check:

- bucket name - `google-dev`
- domain name - `uber.com`, `sub.domain.com`
- full s3 url - `yahoo-staging.s3-us-west-2.amazonaws.com` (To easily combine with other tools like [bucket-stream](https://github.com/eth0izzle/bucket-stream))
- bucket:region - `flaws.cloud:us-west-2`

```bash
> cat names.txt
flaws.cloud
google-dev
testing.microsoft.com
yelp-production.s3-us-west-1.amazonaws.com
github-dev:us-east-1
```
	
1. Dump all open buckets, log both open and closed buckets to found.txt
	
	```bash
	> python ./s3scanner.py --include-closed --out-file found.txt --dump names.txt
	```
2. Just log open buckets to the default output file (buckets.txt)

	```bash
	> python ./s3scanner.py names.txt
	```
3. Save file listings of all open buckets to file
    ```bash
    > python ./s3scanner.py --list names.txt

    ```

## Contributing
Issues are welcome and Pull Requests are appreciated. All contributions should be compatible with both Python 2.7 and 3.6.

|    master    |    [![Build Status](https://travis-ci.org/sa7mon/S3Scanner.svg?branch=master)](https://travis-ci.org/sa7mon/S3Scanner)    |
|:------------:|:-------------------------------------------------------------------------------------------------------------------------:|
| enhancements | [![Build Status](https://travis-ci.org/sa7mon/S3Scanner.svg?branch=enhancements)](https://travis-ci.org/sa7mon/S3Scanner) |
|     bugs     |     [![Build Status](https://travis-ci.org/sa7mon/S3Scanner.svg?branch=bugs)](https://travis-ci.org/sa7mon/S3Scanner)     |

### Testing
* All test are currently in `test_scanner.py`
* Run tests with in 2.7 and 3.6 virtual environments.
* This project uses **pytest-xdist** to run tests. Use `pytest -n NUM` where num is number of parallel processes.
* Run individual tests like this: `pytest -q -s test_scanner.py::test_namehere`

### Contributors
* [Ohelig](https://github.com/Ohelig)
* [vysecurity](https://github.com/vysecurity)
* [janmasarik](https://github.com/janmasarik)
* [alanyee](https://github.com/alanyee)
* [hipotermia](https://github.com/hipotermia)

## License
License: [MIT](LICENSE.txt) https://opensource.org/licenses/MIT
