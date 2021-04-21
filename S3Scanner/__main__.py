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
from sys import exit
from .S3Bucket import S3Bucket, BucketExists, Permission
from .S3Service import S3Service
from concurrent.futures import ThreadPoolExecutor, as_completed
from .exceptions import InvalidEndpointException

CURRENT_VERSION = '2.0.0'
AWS_ENDPOINT = 'https://s3.amazonaws.com'


# We want to use both formatter classes, so a custom class it is
class CustomFormatter(argparse.RawTextHelpFormatter, argparse.RawDescriptionHelpFormatter):
    pass


def load_bucket_names_from_file(file_name):
    """
    Load in bucket names from a text file

    :param str file_name: Path to text file
    :return: set: All lines of text file
    """
    buckets = set()
    if path.isfile(file_name):
        with open(file_name, 'r') as f:
            for line in f:
                line = line.rstrip()  # Remove any extra whitespace
                buckets.add(line)
        return buckets
    else:
        print("Error: '%s' is not a file" % file_name)
        exit(1)


def scan_single_bucket(s3service, anons3service, do_dangerous, bucket_name):
    """
    Scans a single bucket for permission issues. Exists on its own so we can do multi-threading

    :param S3Service s3service: S3Service with credentials to use for scanning
    :param S3Service anonS3Service: S3Service without credentials to use for scanning
    :param bool do_dangerous: Whether or not to do dangerous checks
    :param str bucket_name: Name of bucket to check
    :return: None
    """
    try:
        b = S3Bucket(bucket_name)
    except ValueError as ve:
        if str(ve) == "Invalid bucket name":
            print(f"{bucket_name} | bucket_invalid_name")
            return
        else:
            print(f"{bucket_name} | {str(ve)}")
            return

    # Check if bucket exists first
    # Use credentials if configured and if we're hitting AWS. Otherwise, check anonymously
    if s3service.endpoint_url == AWS_ENDPOINT:
        s3service.check_bucket_exists(b)
    else:
        anons3service.check_bucket_exists(b)

    if b.exists == BucketExists.NO:
        print(f"{b.name} | bucket_not_exist")
        return
    checkAllUsersPerms = True
    checkAuthUsersPerms = True

    # 1. Check for ReadACP
    anons3service.check_perm_read_acl(b)  # Check for AllUsers
    if s3service.aws_creds_configured and s3service.endpoint_url == AWS_ENDPOINT:
        s3service.check_perm_read_acl(b)  # Check for AuthUsers

    # If FullControl is allowed for either AllUsers or AnonUsers, skip the remainder of those tests
    if b.AuthUsersFullControl == Permission.ALLOWED:
        checkAuthUsersPerms = False
    if b.AllUsersFullControl == Permission.ALLOWED:
        checkAllUsersPerms = False

    # 2. Check for Read
    if checkAllUsersPerms:
        anons3service.check_perm_read(b)
    if s3service.aws_creds_configured and checkAuthUsersPerms and s3service.endpoint_url == AWS_ENDPOINT:
        s3service.check_perm_read(b)

    # Do dangerous/destructive checks
    if do_dangerous:
        # 3. Check for Write
        if checkAllUsersPerms:
            anons3service.check_perm_write(b)
        if s3service.aws_creds_configured and checkAuthUsersPerms:
            s3service.check_perm_write(b)

        # 4. Check for WriteACP
        # TODO: Actually check this permission
        if checkAllUsersPerms:
            pass
        if s3service.aws_creds_configured and checkAuthUsersPerms:
            pass

    print(f"{b.name} | bucket_exists | {b.get_human_readable_permissions()}")


def main():
    # Instantiate the parser
    parser = argparse.ArgumentParser(description='s3scanner: Audit unsecured S3 buckets\n'
                                                 '           by Dan Salmon - github.com/sa7mon, @bltjetpack\n',
                                     prog='s3scanner', allow_abbrev=False, formatter_class=CustomFormatter)
    # Declare arguments
    parser.add_argument('--version', action='version', version=CURRENT_VERSION,
                        help='Display the current version of this tool')
    parser.add_argument('--threads', '-t', type=int, default=4, dest='threads', help='Number of threads to use. Default: 4',
                        metavar='n')
    parser.add_argument('--endpoint-url', '-u', dest='endpoint_url',
                        help='URL of S3-compliant API. Default: https://s3.amazonaws.com',
                        default='https://s3.amazonaws.com')
    parser.add_argument('--endpoint-address-style', '-s', dest='endpoint_address_style', choices=['path', 'vhost'],
                        default='path', help='Address style to use for the endpoint. Default: path')
    parser.add_argument('--insecure', '-i', dest='verify_ssl', action='store_false', help='Do not verify SSL')
    subparsers = parser.add_subparsers(title='mode', dest='mode', help='(Must choose one)')

    # Scan mode
    parser_scan = subparsers.add_parser('scan', help='Scan bucket permissions')
    parser_scan.add_argument('--dangerous', action='store_true',
                             help='Include Write and WriteACP permissions checks (potentially destructive)')
    parser_group = parser_scan.add_mutually_exclusive_group(required=True)
    parser_group.add_argument('--buckets-file', '-f', dest='buckets_file',
                              help='Name of text file containing bucket names to check', metavar='file')
    parser_group.add_argument('--bucket', '-b', dest='bucket', help='Name of bucket to check', metavar='bucket')
    # TODO: Get help output to not repeat metavar names - i.e. `--bucket FILE, -f FILE`
    #   https://stackoverflow.com/a/9643162/2307994

    # Dump mode
    parser_dump = subparsers.add_parser('dump', help='Dump the contents of buckets')
    parser_dump.add_argument('--dump-dir', '-d', required=True, dest='dump_dir', help='Directory to dump bucket into')
    dump_parser_group = parser_dump.add_mutually_exclusive_group(required=True)
    dump_parser_group.add_argument('--buckets-file', '-f', dest='dump_buckets_file',
                                   help='Name of text file containing bucket names to check', metavar='file')
    dump_parser_group.add_argument('--bucket', '-b', dest='dump_bucket', help='Name of bucket to check', metavar='bucket')
    parser_dump.add_argument('--verbose', '-v', dest='dump_verbose', action='store_true',
                             help='Enable verbose output while dumping bucket(s)')

    # Parse the args
    args = parser.parse_args()

    if 'http://' not in args.endpoint_url and 'https://' not in args.endpoint_url:
        print("Error: endpoint_url must start with http:// or https:// scheme")
        exit(1)

    s3service = None
    anons3service = None
    try:
        s3service = S3Service(endpoint_url=args.endpoint_url, verify_ssl=args.verify_ssl, endpoint_address_style=args.endpoint_address_style)
        anons3service = S3Service(forceNoCreds=True, endpoint_url=args.endpoint_url, verify_ssl=args.verify_ssl, endpoint_address_style=args.endpoint_address_style)
    except InvalidEndpointException as e:
        print(f"Error: {e.message}")
        exit(1)

    if s3service.aws_creds_configured is False:
        print("Warning: AWS credentials not configured - functionality will be limited. Run:"
              " `aws configure` to fix this.\n")

    bucketsIn = set()

    if args.mode == 'scan':
        if args.buckets_file is not None:
            bucketsIn = load_bucket_names_from_file(args.buckets_file)
        elif args.bucket is not None:
            bucketsIn.add(args.bucket)

        if args.dangerous:
            print("INFO: Including dangerous checks. WARNING: This may change bucket ACL destructively")

        with ThreadPoolExecutor(max_workers=args.threads) as executor:
            futures = {
                executor.submit(scan_single_bucket, s3service, anons3service, args.dangerous, bucketName): bucketName for bucketName in bucketsIn
            }
            for future in as_completed(futures):
                if future.exception():
                    print(f"Bucket scan raised exception: {futures[future]} - {future.exception()}")

    elif args.mode == 'dump':
        if args.dump_dir is None or not path.isdir(args.dump_dir):
            print("Error: Given --dump-dir does not exist or is not a directory")
            exit(1)
        if args.dump_buckets_file is not None:
            bucketsIn = load_bucket_names_from_file(args.dump_buckets_file)
        elif args.dump_bucket is not None:
            bucketsIn.add(args.dump_bucket)

        for bucketName in bucketsIn:
            try:
                b = S3Bucket(bucketName)
            except ValueError as ve:
                if str(ve) == "Invalid bucket name":
                    print(f"{bucketName} | bucket_name_invalid")
                    continue
                else:
                    print(f"{bucketName} | {str(ve)}")
                    continue

            # Check if bucket exists first
            s3service.check_bucket_exists(b)

            if b.exists == BucketExists.NO:
                print(f"{b.name} | bucket_not_exist")
                continue

            s3service.check_perm_read(b)

            if b.AuthUsersRead != Permission.ALLOWED:
                anons3service.check_perm_read(b)
                if b.AllUsersRead != Permission.ALLOWED:
                    print(f"{b.name} | Error: no read permissions")
                else:
                    # Dump bucket without creds
                    print(f"{b.name} | Enumerating bucket objects...")
                    anons3service.enumerate_bucket_objects(b)
                    print(f"{b.name} | Total Objects: {str(len(b.objects))}, Total Size: {b.get_human_readable_size()}")
                    anons3service.dump_bucket_multithread(bucket=b, dest_directory=args.dump_dir,
                                                          verbose=args.dump_verbose, threads=args.threads)
            else:
                # Dump bucket with creds
                print(f"{b.name} | Enumerating bucket objects...")
                s3service.enumerate_bucket_objects(b)
                print(f"{b.name} | Total Objects: {str(len(b.objects))}, Total Size: {b.get_human_readable_size()}")
                s3service.dump_bucket_multithread(bucket=b, dest_directory=args.dump_dir,
                                                  verbose=args.dump_verbose, threads=args.threads)
    else:
        print("Invalid mode")
        parser.print_help()


if __name__ == "__main__":
    main()
