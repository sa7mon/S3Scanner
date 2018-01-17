#########
#
# AWS S3scanner - Scans domain names for A records that resolve to *.amazonaws.com
# 
# Author:  Dan Salmon (twitter.com/bltjetpack, github.com/sa7mon)
# Created: 6/19/17
# License: Creative Commons (CC BY-NC-SA 4.0))
#
#########

import argparse
import requests

defaultRegion = "us-west-1"


def pprint(good, message):
    if good:
        # print in green
        print("\033[0;32m" + message + "\033[0;m")
        logFile.write(message + "\n")
    else:
        # print in red
        print("\033[0;91m" + message + "\033[0;m")


def checkSite(site, region):
    """ Does a simple GET request with the Requests library and interprets the results.

    site - A domain name without protocol (http[s])
    region - An s3 region. See: https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region
    """

    # Concat domain name with the default region
    bucketDomain = 'http://' + site + '.s3-' + region + '.amazonaws.com'
    r = requests.get(bucketDomain)

    if r.status_code == 200:
        # Successfully found a bucket!
        pprint(True, site + ":" +
               region)
    elif r.status_code == 301:
        # We got the region wrong. The 'x-amz-bucket-region' header will give us the correct one.
        checkSite(site, r.headers['x-amz-bucket-region'])
    elif r.status_code == 403:
        # We probably need to have an AWS account defined.
        pprint(False, site + ":" + region + ' - Need an authenticated user to verify')
    elif r.status_code == 404:
        # This is definitely not a valid bucket name.
        pprint(False, site)
    else:
        raise ValueError("Got an unhandled status code back: " + str(r.status_code) + " for site: " + site + ":" + region)


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
        checkSite(site, defaultRegion)
