import os
import re
import signal
from contextlib import contextmanager
import datetime

import boto3
from botocore.exceptions import ClientError, NoCredentialsError, HTTPClientError
from botocore.handlers import disable_signing
from botocore import UNSIGNED
from botocore.client import Config
import requests


SIZE_CHECK_TIMEOUT = 30    # How long to wait for getBucketSize to return
AWS_CREDS_CONFIGURED = True
ERROR_CODES = ['AccessDenied', 'AllAccessDisabled', '[Errno 21] Is a directory:']


class TimeoutException(Exception): pass

@contextmanager
def time_limit(seconds):
    def signal_handler(signum, frame):
        raise TimeoutException("Timed out!")
    signal.signal(signal.SIGALRM, signal_handler)
    signal.alarm(seconds)
    try:
        yield
    finally:
        signal.alarm(0)



def checkAcl(bucket):
    """
    Attempts to retrieve a bucket's ACL. This also functions as the main 'check if bucket exists' function.
    By trying to get the ACL, we combine 2 steps to minimize potentially slow network calls.

    :param bucket: Name of bucket to try to get the ACL of
    :return: A dictionary with 2 entries:
        found - Boolean. True/False whether or not the bucket was found
        acls - dictionary. If ACL was retrieved, contains 2 keys: 'allUsers' and 'authUsers'. If ACL was not
                            retrieved,
    """
    allUsersGrants = []
    authUsersGrants = []

    s3 = boto3.resource('s3')

    try:
        bucket_acl = s3.BucketAcl(bucket)
        bucket_acl.load()
    except s3.meta.client.exceptions.NoSuchBucket:
        return {"found": False, "acls": {}}

    except ClientError as e:
        if e.response['Error']['Code'] == "AccessDenied":
            return {"found": True, "acls": "AccessDenied"}
        elif e.response['Error']['Code'] == "AllAccessDisabled":
            return {"found": True, "acls": "AllAccessDisabled"}
        else:
            raise e

    for grant in bucket_acl.grants:
        if 'URI' in grant['Grantee']:
            if grant['Grantee']['URI'] == "http://acs.amazonaws.com/groups/global/AllUsers":
                allUsersGrants.append(grant['Permission'])
            elif grant['Grantee']['URI'] == "http://acs.amazonaws.com/groups/global/AuthenticatedUsers":
                authUsersGrants.append(grant['Permission'])

    return {"found": True, "acls": {"allUsers": allUsersGrants, "authUsers": authUsersGrants}}


def checkAwsCreds():
    """
    Checks to see if the user has credentials for AWS properly configured.
    This is essentially a requirement for getting accurate results.

    :return: True if AWS credentials are properly configured. False if not.
    """

    sts = boto3.client('sts')
    try:
        response = sts.get_caller_identity()
    except NoCredentialsError as e:
            return False

    return True


def checkBucket(inBucket, slog, flog, argsDump, argsList):
    # Determine what kind of input we're given. Options:
    #   bucket name   i.e. mybucket
    #   domain name   i.e. flaws.cloud
    #   full S3 url   i.e. flaws.cloud.s3-us-west-2.amazonaws.com
    #   bucket:region i.e. flaws.cloud:us-west-2
    
    if ".amazonaws.com" in inBucket:    # We were given a full s3 url
        bucket = inBucket[:inBucket.rfind(".s3")]
    elif ":" in inBucket:               # We were given a bucket in 'bucket:region' format
        bucket = inBucket.split(":")[0]
    else:                           # We were either given a bucket name or domain name
        bucket = inBucket

    valid = checkBucketName(bucket)

    if not valid:
        message = "{0:>11} : {1}".format("[invalid]", bucket)
        slog.error(message)
        # continue
        return

    if AWS_CREDS_CONFIGURED:
        b = checkAcl(bucket)
    else:
        a = checkBucketWithoutCreds(bucket)
        b = {"found": a, "acls": "unknown - no aws creds"}

    if b["found"]:

        size = getBucketSize(bucket)  # Try to get the size of the bucket

        message = "{0:>11} : {1}".format("[found]", bucket + " | " + str(size) + " | ACLs: " + str(b["acls"]))
        slog.info(message)
        flog.debug(bucket)

        if argsDump:
            if size not in ["AccessDenied", "AllAccessDisabled"]:
                slog.info("{0:>11} : {1} - {2}".format("[found]", bucket, "Attempting to dump...this may take a while."))
                dumpBucket(bucket)
        if argsList:
            if str(b["acls"]) not in ["AccessDenied", "AllAccessDisabled"]:
                listBucket(bucket)
    else:
        message = "{0:>11} : {1}".format("[not found]", bucket)
        slog.error(message)


def checkBucketName(bucket_name):
    """ Checks to make sure bucket names input are valid according to S3 naming conventions
    :param bucketName: Name of bucket to check
    :return: Boolean - whether or not the name is valid
    """

    # Bucket names can be 3-63 (inclusively) characters long.
    # Bucket names may only contain lowercase letters, numbers, periods, and hyphens
    pattern = r'(?=^.{3,63}$)(?!^(\d+\.)+\d+$)(^(([a-z0-9]|[a-z0-9][a-z0-9\-]*[a-z0-9])\.)*([a-z0-9]|[a-z0-9][a-z0-9\-]*[a-z0-9])$)'

    
    return bool(re.match(pattern, bucket_name))


def checkBucketWithoutCreds(bucketName, triesLeft=2):
    """ Does a simple GET request with the Requests library and interprets the results.
    bucketName - A domain name without protocol (http[s]) """

    if triesLeft == 0:
        return False

    bucketUrl = 'http://' + bucketName + '.s3.amazonaws.com'

    r = requests.head(bucketUrl)

    if r.status_code == 200:    # Successfully found a bucket!
        return True
    elif r.status_code == 403:  # Bucket exists, but we're not allowed to LIST it.
        return True
    elif r.status_code == 404:  # This is definitely not a valid bucket name.
        return False
    elif r.status_code == 503:
        return checkBucketWithoutCreds(bucketName, triesLeft - 1)
    else:
        raise ValueError("Got an unhandled status code back: " + str(r.status_code) + " for bucket: " + bucketName +
                         ". Please open an issue at: https://github.com/sa7mon/s3scanner/issues and include this info.")


def dumpBucket(bucketName):
    global dumped
    # Dump the bucket into bucket folder
    bucketDir = './buckets/' + bucketName

    if not os.path.exists(bucketDir):
        os.makedirs(bucketDir)

    dumped = True
    
    s3 = boto3.client('s3')

    try:
        if AWS_CREDS_CONFIGURED is False:
            s3 = boto3.client('s3', config=Config(signature_version=UNSIGNED))
        
        for page in s3.get_paginator("list_objects_v2").paginate(Bucket=bucketName):
            if 'Contents' in page:
                for item in page['Contents']:
                    key = item['Key']
                    s3.download_file(bucketName, key, bucketDir+"/"+key)
        dumped = True
    except ClientError as e:
        # global dumped
        if e.response['Error']['Code'] == 'AccessDenied':
            pass  # TODO: Do something with the fact that we were denied
        dumped = False
    finally:
        # Check if folder is empty. If it is, delete it
        if not os.listdir(bucketDir):
            os.rmdir(bucketDir)
        return dumped


def getBucketSize(bucketName):
    """
    Use awscli to 'ls' the bucket which will give us the total size of the bucket.
    NOTE:
        Function assumes the bucket exists and doesn't catch errors if it doesn't.
    """
    s3 = boto3.client('s3')
    try:
        if AWS_CREDS_CONFIGURED is False:
            s3 = boto3.client('s3', config=Config(signature_version=UNSIGNED))
        size_bytes = 0
        with time_limit(SIZE_CHECK_TIMEOUT):
            for page in s3.get_paginator("list_objects_v2").paginate(Bucket=bucketName):
                if 'Contents' in page:
                    for item in page['Contents']:
                       size_bytes += item['Size']
        return str(size_bytes) + " bytes"

    except HTTPClientError as e:
        if "Timed out!" in str(e):
            return "Unknown Size - timeout"
        else:
            raise e
    except ClientError as e:
        if e.response['Error']['Code'] == 'AccessDenied':
            return "AccessDenied"
        elif e.response['Error']['Code'] == 'AllAccessDisabled':
            return "AllAccessDisabled"
        elif e.response['Error']['Code'] == 'NoSuchBucket':
            return "NoSuchBucket"
        else:
            raise e


def listBucket(bucketName):
    """ 
        If we find an open bucket, save the contents of the bucket listing to file. 
        Returns:
            None if no errors were encountered
    """

    # Dump the bucket into bucket folder
    bucketDir = './list-buckets/' + bucketName + '.txt'
    if not os.path.exists('./list-buckets/'):
        os.makedirs('./list-buckets/')

    s3 = boto3.client('s3')
    objects = []

    try:
        if AWS_CREDS_CONFIGURED is False:
            s3 = boto3.client('s3', config=Config(signature_version=UNSIGNED))
        
        for page in s3.get_paginator("list_objects_v2").paginate(Bucket=bucketName):
            if 'Contents' in page:
                for item in page['Contents']:
                    o = item['LastModified'].strftime('%Y-%m-%d %H:%M:%S') + " " + str(item['Size']) + " " + item['Key']
                    objects.append(o)

        with open(bucketDir, 'w') as f:
            for o in objects:
                f.write(o + "\n")

    except ClientError as e:
        if e.response['Error']['Code'] == 'AccessDenied':
            return "AccessDenied"
        else:
            raise e
