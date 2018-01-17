# S3Scanner

[![Build Status](https://travis-ci.org/sa7mon/S3Scanner.svg?branch=master)](https://travis-ci.org/sa7mon/S3Scanner)

A quick and dirty script to find unsecured S3 buckets and dump their contents :droplet:

The tool has 2 parts:

## 1 - s3finder.py
This script takes a list of domain names and checks if their hosted on Amazon S3. Found S3 domains are output to file with their corresponding region in format "domain:region". 

**Usage:** `$> python s3finder.py -o output.txt domainsToCheck.txt`

![1 - s3finder.py](https://user-images.githubusercontent.com/3712226/30464005-9e8ab0c8-9994-11e7-8757-e32f1db975da.png)

**Install:**
1. (Optional) `virtualenv venv && source ./venv/bin/activate`
2. `pip install -r requirements.txt`

## 2 - s3dumper.sh
This script takes in a list of domains with regions made by s3finder.py. For each domain, it checks if there are publicly readable buckets and dumps them if so. 

**Usage:** `$> s3dumper.sh output.txt`

**Support:**
Only tested with Python 2.7.14

**Requirements:** [aws-cli](http://docs.aws.amazon.com/cli/latest/userguide/installing.html)

![2 - s3dumper.sh](https://user-images.githubusercontent.com/3712226/30464321-8e8e6d34-9996-11e7-8739-94f13e082877.png)

# Contributing
Please make pull requests if you can improve on the code at all (which is certain as the code can be greatly optimized).

# License
Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International [(CC BY-NC-SA 4.0)](https://creativecommons.org/licenses/by-nc-sa/4.0/)
