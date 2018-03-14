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
import coloredlogs
import sys
import botocore

# We want to use both formatter classes, so a custom class it is
class CustomFormatter(argparse.RawTextHelpFormatter, argparse.RawDescriptionHelpFormatter):
    pass


# Instantiate the parser
parser = argparse.ArgumentParser(description='#  s3scanner - Find S3 buckets and dump!\n'
                                             '#\n'
                                             '#  Author: Dan Salmon - @bltjetpack, github.com/sa7mon\n',
                                 prog='s3scanner', formatter_class=CustomFormatter)

# Declare arguments
parser.add_argument('-o', '--out-file', required=False, dest='outFile',
                    help='Name of file to save the successfully checked buckets in (Default: buckets.txt)')
parser.add_argument('-c', '--include-closed', required=False, dest='includeClosed', action='store_true',
                    help='Include found but closed buckets in the out-file')
parser.add_argument('-r', '--default-region', dest='',
                    help='AWS region to default to (Default: us-west-1)')
parser.add_argument('-d', '--dump', required=False, dest='dump', action='store_true',
                    help='Dump all found open buckets locally')
parser.add_argument('-l', '--list', required=False, dest='list', action='store_true',
                    help='List all found open buckets locally')
parser.add_argument('buckets', help='Name of text file containing buckets to check')

parser.set_defaults(defaultRegion='us-west-1')
parser.set_defaults(includeClosed=False)
parser.set_defaults(outFile='./buckets.txt')
parser.set_defaults(dump=False)

# If there are no args supplied, print the full help text instead of the short usage text
if len(sys.argv) == 1:
    parser.print_help()
    sys.exit(1)

# Parse the args
args = parser.parse_args()

# Create file logger
flog = logging.getLogger('s3scanner-file')
flog.setLevel(logging.DEBUG)              # Set log level for logger object

# Create file handler which logs even debug messages
fh = logging.FileHandler(args.outFile)
fh.setLevel(logging.DEBUG)

# Add the handler to logger
flog.addHandler(fh)

# Create secondary logger for logging to screen
slog = logging.getLogger('s3scanner-screen')
slog.setLevel(logging.INFO)

# Logging levels for the screen logger:
#   INFO  = found, open
#   WARN  = found, closed
#   ERROR = not found
# The levels serve no other purpose than to specify the output color

levelStyles = {
        'info': {'color': 'blue'},
        'warning': {'color': 'yellow'},
        'error': {'color': 'red'}
        }

fieldStyles = {
        'asctime': {'color': 'white'}
        }

# Use coloredlogs to add color to screen logger. Define format and styles.
coloredlogs.install(level='DEBUG', logger=slog, fmt='%(asctime)s   %(message)s',
                    level_styles=levelStyles, field_styles=fieldStyles)

if not s3.checkAwsCreds():
    s3.awsCredsConfigured = False
    slog.error("Warning: AWS credentials not configured. Open buckets will be shown as closed. Run:"
               " `aws configure` to fix this.\n")

with open(args.buckets, 'r') as f:
    for line in f:
        line = line.rstrip()            # Remove any extra whitespace
        region = args.defaultRegion

        # Determine what kind of input we're given. Options:
        #   bucket name   i.e. mybucket
        #   domain name   i.e. flaws.cloud
        #   full S3 url   i.e. flaws.cloud.s3-us-west-2.amazonaws.com
        #   bucket:region i.e. flaws.cloud:us-west-2

        if ".amazonaws.com" in line:    # We were given a full s3 url
            bucket = line[:line.rfind(".s3")]
            region = line[len(line[:line.rfind(".s3")]) + 4:line.rfind(".amazonaws.com")]
        elif ":" in line:               # We were given a bucket in 'bucket:region' format
            region = line.split(":")[1]
            bucket = line.split(":")[0]
        else:                           # We were either given a bucket name or domain name
            bucket = line

        valid = s3.checkBucketName(bucket)

        if not valid:
            message = "{0:>16} : {1}".format("[invalid]", bucket)
            slog.error(message)
            continue

        b = s3.checkAcl(bucket)

        if b["found"]:
            # print("Found bucket: " + bucket + " | Acls: " + str(b["acls"]))
            size = "Unknown Size - Closed"
            if str(b["acls"]) != "AccessDenied":
                size = s3.getBucketSize(bucket)

            message = "{0:<7}{1:>9} : {2}".format("[found]", "[open]", bucket + " | " + size + " | ACLs: " +
                                                  str(b["acls"]))
            slog.info(message)
            flog.debug(bucket)

            if args.dump:
                s3.dumpBucket(bucket)
            if args.list:
                if str(b["acls"]) != "AccessDenied":
                    s3.listBucket(bucket)
        else:
            message = "{0:>16} : {1}".format("[not found]", bucket)
            slog.error(message)


        # if result[0] in [900, 404]:     # These are our 'bucket not found' codes
        #     slog.error(result[1])
        #
        # elif result[0] == 403:          # Found but closed bucket. Only log if user says to.
        #     message = "{0:>15} : {1}".format("[found] [closed]", result[1] + ":" + result[2])
        #     slog.warning(message)
        #     if args.includeClosed:      # If user supplied '--include-closed' flag, log this bucket to file
        #         flog.debug(result[1] + ":" + result[2])
        #
        #
        # elif result[0] == 999:
        #     message = "{0:>16} : {1}".format("[invalid]", result[1])
        #     slog.error(message)
        #
        # else:
        #     raise ValueError("Got back unknown code from checkBucket(): " +
        #                      str(result[0]) + "Please report this to https://github.com/sa7mon/S3Scanner/issues")
