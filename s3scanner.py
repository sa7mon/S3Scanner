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
import os.path

currentVersion = '1.0.0'



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
# parser.add_argument('-c', '--include-closed', required=False, dest='includeClosed', action='store_true',
#                     help='Include found but closed buckets in the out-file')
parser.add_argument('-d', '--dump', required=False, dest='dump', action='store_true',
                    help='Dump all found open buckets locally')
parser.add_argument('-l', '--list', required=False, dest='list', action='store_true',
                    help='List all found open buckets locally')
parser.add_argument('--version', required=False, dest='version', action='store_true',
                    help='Display the current version of this tool')
parser.add_argument('buckets', help='Name of text file containing buckets to check')

# parser.set_defaults(includeClosed=False)
parser.set_defaults(outFile='./buckets.txt')
parser.set_defaults(dump=False)


if len(sys.argv) == 1:              # No args supplied, print the full help text instead of the short usage text
    parser.print_help()
    sys.exit(0)
elif len(sys.argv) == 2:
    if sys.argv[1] == '--version':  # Only --version arg supplied. Print the version and exit.
        print(currentVersion)
        sys.exit(0)


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
    s3.awsCredsConfigured = False
    slog.error("Warning: AWS credentials not configured. Open buckets will be shown as closed. Run:"
               " `aws configure` to fix this.\n")

if os.path.isfile(args.buckets):
    with open(args.buckets, 'r') as f:
        for line in f:
            line = line.rstrip()            # Remove any extra whitespace
            s3.checkBucket(line, slog, flog, args.dump, args.list)
else:
    # It's a single bucket
    s3.checkBucket(args.buckets, slog, flog, args.dump, args.list)