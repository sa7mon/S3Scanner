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


def pprint(good, message, log):
    if good:
        # print in green
        print("\033[0;32m" + message + "\033[0;m")
    else:
        # print in red
        print("\033[0;91m" + message + "\033[0;m")

    if log:
        logFile.write(message + "\n")

# Instantiate the parser
parser = argparse.ArgumentParser(description='Find AWS S3 buckets!')

# Declare arguments
parser.add_argument('-o', '--out-file', required=False, dest='bucketsFile',
                    help='Name of file to save the successfully checked domains in. Default: buckets.txt')
parser.add_argument('-c', '--include-closed', required=False, dest='includeClosed', action='store_true',
                    help='Include found but closed buckets in the outFile. Default: false')
parser.add_argument('-r', '--default-region', dest='',
                    help='AWS region to check first for buckets. Default: us-west-1')
parser.add_argument('domains', help='Name of text file containing domains to check')

parser.set_defaults(defaultRegion='us-west-1')
parser.set_defaults(includeClosed=False)
parser.set_defaults(bucketsFile='./buckets.txt')


# Parse the args
args = parser.parse_args()


# Open log file for writing
logFile = open(args.bucketsFile, 'a+')

with open(args.domains, 'r') as f:
    for line in f:
        site = line.rstrip()
        result = s3.checkBucket(site, args.defaultRegion)

        if result[0] == 301:
            result = s3.checkBucket(site, result[1])
        if result[0] in [900, 404]:     # These are our 'bucket not found' codes
            pprint(False, result[1], False)
        elif result[0] == 403:          # Found but closed bucket. Only log if user says to.
            pprint(False, result[1], args.includeClosed)
        elif result[0] == 200:          # The only 'bucket found and open' codes
            pprint(True, result[1], True)
        else:
            raise ValueError("Got back unknown code from checkBucket()")
