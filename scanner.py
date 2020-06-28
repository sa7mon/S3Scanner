#########
#
# S3scanner - Audit unsecured S3 buckets
# 
# Author:  Dan Salmon (twitter.com/bltjetpack, github.com/sa7mon)
# Created: 6/19/17
# License: MIT
#
#########

import argparse
from os import path
import os
from s3Bucket import s3Bucket, BucketExists, Permission
from S3Service import S3Service

CURRENT_VERSION = '1.0.0'


# We want to use both formatter classes, so a custom class it is
class CustomFormatter(argparse.RawTextHelpFormatter, argparse.RawDescriptionHelpFormatter):
    pass


"""
Default arguments
    scanner.py [--version --help]

Scan mode
    scanner.py scan [--dangerous] [--bucket-file <file.txt>] bucket_name
    
Dump mode
    scanner.py dump [--dump-dir] [--bucket-file <file.txt>] bucket_name
"""
# Instantiate the parser
parser = argparse.ArgumentParser(description='s3scanner: Audit unsecured S3 buckets\n'
                                             '           by Dan Salmon - github.com/sa7mon, @bltjetpack\n',
                                 prog='s3scanner', allow_abbrev=False, formatter_class=CustomFormatter)
# Declare arguments
parser.add_argument('--version', action='version', version=CURRENT_VERSION,
                    help='Display the current version of this tool')
subparsers = parser.add_subparsers(title='mode', dest='mode', help='')

# Scan mode
parser_scan = subparsers.add_parser('scan', help='Scan bucket permissions')
parser_scan.add_argument('--dangerous', action='store_true', help='Include Write and WriteACP permissions checks')
parser_group = parser_scan.add_mutually_exclusive_group(required=True)
parser_group.add_argument('--buckets-file', '-f', dest='buckets_file',
                          help='Name of text file containing bucket names to check', metavar='FILE')
parser_group.add_argument('--bucket', '-b', dest='bucket', help='Name of bucket to check')
# TODO: Get help output to not repeat metavar names - i.e. --bucket FILE, -f FILE
#   https://stackoverflow.com/a/9643162/2307994

# Dump mode
parser_dump = subparsers.add_parser('dump', help='Dump the contents of buckets')
parser_dump.add_argument('--dump-dir', '-d', )

# Parse the args
args = parser.parse_args()

s3service = S3Service()
anonS3Service = S3Service(forceNoCreds=True)

if s3service.aws_creds_configured is False:
    print("Warning: AWS credentials not configured - functionality will be limited. Run:"
          " `aws configure` to fix this.\n")

if args.mode == 'scan':
    bucketsIn = set()
    if args.buckets_file is not None:
        if path.isfile(args.buckets_file):
            with open(args.buckets_file, 'r') as f:
                for line in f:
                    line = line.rstrip()  # Remove any extra whitespace
                    bucketsIn.add(line)
        else:
            print("Error: '%s' is not a file" % args.buckets_file)
    elif args.bucket is not None:
        bucketsIn.add(args.bucket)

    if args.dangerous:
        print("INFO: Including dangerous checks. WARNING: This may change bucket ACL destructively")

    for bucketName in bucketsIn:
        try:
            b = s3Bucket(bucketName)
        except ValueError as ve:
            if str(ve) == "Invalid bucket name":
                print("[%s] Invalid bucket name" % bucketName)
                continue
            else:
                print("[%s] %s" % (bucketName, str(ve)))
                continue

        # Check if bucket exists first
        s3service.check_bucket_exists(b)

        if b.exists == BucketExists.NO:
            print("[%s] Bucket doesn't exist" % b.name)
            continue

        checkAllUsersPerms = True
        checkAuthUsersPerms = True

        # 1. Check for ReadACP
        anonS3Service.check_perm_read_acl(b)  # Check for AllUsers
        if s3service.aws_creds_configured:
            s3service.check_perm_read_acl(b)  # Check for AuthUsers

        # If FullControl is allowed for either AllUsers or AnonUsers, skip the remainder of those tests
        if b.AuthUsersFullControl == Permission.ALLOWED:
            checkAuthUsersPerms = False
        if b.AllUsersFullControl == Permission.ALLOWED:
            checkAllUsersPerms = False

        # 2. Check for Read
        if checkAllUsersPerms:
            anonS3Service.check_perm_read(b)
        if s3service.aws_creds_configured and checkAuthUsersPerms:
            s3service.check_perm_read(b)

        # Do dangerous/destructive checks
        if args.dangerous:
            # 3. Check for Write
            if checkAllUsersPerms:
                anonS3Service.check_perm_write(b)
            if s3service.aws_creds_configured and checkAuthUsersPerms:
                s3service.check_perm_write(b)

            # 4. Check for WriteACP
            if checkAllUsersPerms:
                pass
            if s3service.aws_creds_configured and checkAuthUsersPerms:
                pass

        print("[%s] %s" % (b.name, b.getHumanReadablePermissions()))
elif args.mode == 'dump':
    pass
else:
    print("Invalid mode")

