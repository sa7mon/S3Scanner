#########
#
# AWS S3scanner - Scans domain names for S3 buckets
# 
# Author:  Dan Salmon (twitter.com/bltjetpack, github.com/sa7mon)
# Created: 6/19/17
# License: Creative Commons (CC BY-NC-SA 4.0))
#
#########

import argparse
import requests
import sh

defaultRegion = "us-west-1"


def pprint(good, message):
    if good:
        # print in green
        print("\033[0;32m" + message + "\033[0;m")
        logFile.write(message + "\n")
    else:
        # print in red
        print("\033[0;91m" + message + "\033[0;m")


def getBucketSize(bucketName):
    """ Use awscli to 'ls' the bucket which will give us the total size of the bucket."""

    a = sh.aws('s3', 'ls', '--summarize', '--human-readable', '--recursive', '--no-sign-request','s3://' + bucketName)

    # Get the last line of the output, get everything to the right of the colon, and strip whitespace
    return a.splitlines()[len(a.splitlines())-1].split(":")[1].strip()


def checkBucket(bucketName, region):
    """ Does a simple GET request with the Requests library and interprets the results.

    site - A domain name without protocol (http[s])
    region - An s3 region. See: https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region
    """

    # Concat domain name with the default region
    bucketDomain = 'http://' + bucketName + '.s3-' + region + '.amazonaws.com'
    try:
        r = requests.get(bucketDomain)
    except requests.exceptions.ConnectionError:
        # Couldn't resolve the hostname. Definitely not a bucket.
        message = "{0:>16} : {1}".format("[not found]", bucketName)
        pprint(False, message)
        return
    if r.status_code == 200:
        # Successfully found a bucket!
        message = "{0:<7}{1:>9} : {2}".format("[found]", "[open]", bucketName + ":" + region + " - " + getBucketSize(site))
        pprint(True, message)
    elif r.status_code == 301:
        # We got the region wrong. The 'x-amz-bucket-region' header will give us the correct one.
        checkBucket(site, r.headers['x-amz-bucket-region'])
    elif r.status_code == 403:
        # We probably need to have an AWS account defined.
        message = "{0:>15} : {1}".format("[found] [closed]", bucketName + ":" + region)
        pprint(False, message)
    elif r.status_code == 404:
        # This is definitely not a valid bucket name.
        message = "{0:>15} : {1}".format("[not found]", bucketName)
        pprint(False, message)
    else:
        raise ValueError("Got an unhandled status code back: " + str(r.status_code) + " for site: " + bucketName + ":" + region)


# Instantiate the parser
parser = argparse.ArgumentParser(description='Find S3 sites!')

# Declare arguments
parser.add_argument('-o', '--outFile', required=True, help='Name of file to save the successfully checked domains in')
parser.add_argument('domains', help='Name of text file containing domains to check')

# Parse the args
args = parser.parse_args()

# Open log file for writing
logFile = open(args.outFile, 'a+')

with open(args.domains, 'r') as f:
    for line in f:
        site = line.rstrip()
        checkBucket(site, defaultRegion)
