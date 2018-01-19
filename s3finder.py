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
import s3utils as s3

defaultRegion = "us-west-1"


def pprint(good, message):
    if good:
        # print in green
        print("\033[0;32m" + message + "\033[0;m")
        logFile.write(message + "\n")
    else:
        # print in red
        print("\033[0;91m" + message + "\033[0;m")


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
        result = s3.checkBucket(site, defaultRegion)

        if result[0] == 301:
            result = s3.checkBucket(site, result[1])
        if result[0] in [900, 403, 404]:  # These are our 'bucket not found' codes
            pprint(False, result[1])
        elif result[0] == 200:            # The only 'bucket found and open' codes
            pprint(True, result[1])
        else:
            raise ValueError("Got back unknown code from checkBucket()")


