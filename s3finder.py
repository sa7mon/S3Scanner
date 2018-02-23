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
import logging


def pprint(good, message, log):
    if good:
        # print in green
        print("\033[0;32m" + message + "\033[0;m")
    else:
        # print in red
        print("\033[0;91m" + message + "\033[0;m")

    # if log:
    #     logFile.write(message + "\n")

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
# logFile = open(args.bucketsFile, 'a+')




# Create logger
flog = logging.getLogger('s3scanner')
flog.setLevel(logging.INFO)    # Log level for console?

# Create file handler which logs even debug messages
fh = logging.FileHandler(args.bucketsFile)
fh.setLevel(logging.DEBUG)

# Add the handlers to logger
flog.addHandler(fh)

# Create console handler with a higher log level
# ch = logging.StreamHandler()

# if args.verbose:
#     ch.setLevel(logging.DEBUG)
# else:
#     ch.setLevel(logging.ERROR)

# Create formatter and add it to the handlers
# formatter = logging.Formatter('%(asctime)s %(levelname)s %(message)s', "%Y-%m-%d %H:%M:%S")
# ch.setFormatter(formatter)
# fh.setFormatter(formatter)






with open(args.domains, 'r') as f:
    for line in f:
        site = line.rstrip()
        result = s3.checkBucket(site, args.defaultRegion)

        if result[0] == 301:
            result = s3.checkBucket(site, result[1])
        if result[0] in [900, 404]:     # These are our 'bucket not found' codes
            pprint(False, result[1], False)
        elif result[0] == 403:          # Found but closed bucket. Only log if user says to.
            message = "{0:>15} : {1}".format("[found] [closed]", result[1] + ":" + result[2])

            pprint(False, message, args.includeClosed)

            if args.includeClosed:
                flog.info(result[1] + ":" + result[2])
        elif result[0] == 200:          # The only 'bucket found and open' codes
            message = "{0:<7}{1:>9} : {2}".format("[found]", "[open]", result[1]
                            + ":" + result[2] + " - " + result[3])

            pprint(True, message, True)
            flog.info(result[1] + ":" + result[2])
        else:
            raise ValueError("Got back unknown code from checkBucket()")
