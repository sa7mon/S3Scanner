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
import logging
from os import path
import sys

import coloredlogs

import s3utils as s3

CURRENT_VERSION = '1.0.0'



# We want to use both formatter classes, so a custom class it is
class CustomFormatter(argparse.RawTextHelpFormatter, argparse.RawDescriptionHelpFormatter):
    pass


# Instantiate the parser
parser = argparse.ArgumentParser(description='#  s3scanner - Find S3 buckets and dump!\n'
                                             '#\n'
                                             '#  Author: Dan Salmon - @bltjetpack, github.com/sa7mon\n',
                                 prog='s3scanner', formatter_class=CustomFormatter)

# Declare arguments
parser.add_argument('-o', '--out-file', dest='outFile', default='./buckets.txt',
                    help='Name of file to save the successfully checked buckets in (Default: buckets.txt)')
# parser.add_argument('-c', '--include-closed', dest='includeClosed', action='store_true', default=False,
#                     help='Include found but closed buckets in the out-file')
parser.add_argument('-d', '--dump', dest='dump', action='store_true', default=False,
                    help='Dump all found open buckets locally')
parser.add_argument('-l', '--list', dest='list', action='store_true',
                    help='Save bucket file listing to local file: ./list-buckets/${bucket}.txt')
parser.add_argument('--version', action='version', version=CURRENT_VERSION,
                    help='Display the current version of this tool')
parser.add_argument('buckets', help='Name of text file containing buckets to check')


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
#   INFO  = found
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
    s3.AWS_CREDS_CONFIGURED = False
    slog.error("Warning: AWS credentials not configured. Open buckets will be shown as closed. Run:"
               " `aws configure` to fix this.\n")

if path.isfile(args.buckets):
    with open(args.buckets, 'r') as f:
        for line in f:
            line = line.rstrip()            # Remove any extra whitespace
            s3.checkBucket(line, slog, flog, args.dump, args.list)
else:
    # It's a single bucket
    s3.checkBucket(args.buckets, slog, flog, args.dump, args.list)
