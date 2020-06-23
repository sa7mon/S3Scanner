#########
#
# AWS S3scanner - Scans domain names for S3 buckets
# 
# Author:  Dan Salmon (twitter.com/bltjetpack, github.com/sa7mon)
# Created: 6/19/17
# License: MIT
#
#########

import argparse
import logging
from os import path
import sys

import coloredlogs

import s3utils as s3
from s3Bucket import s3Bucket, BucketExists, Permission
from S3Service import S3Service

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
parser.add_argument('--dangerous', action='store_true', help='Include Write and WriteACP permissions checks')
# parser.add_argument('--version', action='version', version=CURRENT_VERSION,
#                    help='Display the current version of this tool')
parser.add_argument('buckets_file', help='Name of text file containing buckets to check')


# Parse the args
args = parser.parse_args()

s3service = S3Service()
anonS3Service = S3Service(forceNoCreds=True)

if s3service.aws_creds_configured is False:
    print("Warning: AWS credentials not configured - functionality will be limited. Run:"
               " `aws configure` to fix this.\n")

bucketsIn = set()

if path.isfile(args.buckets_file):
    with open(args.buckets_file, 'r') as f:
        for line in f:
            line = line.rstrip()            # Remove any extra whitespace
            bucketsIn.add(line)

if args.dangerous:
    print("INFO: Including dangeous checks. WARNING: This may change bucket ACL destructively")

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

    # 3. Do dangerous/destructive checks
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
