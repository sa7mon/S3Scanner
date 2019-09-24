import os
import re
import sh

import boto3
from botocore.exceptions import ClientError, NoCredentialsError
from botocore.handlers import disable_signing
import requests

SIZE_CHECK_TIMEOUT = 8    # How long to wait for getBucketSize to return
AWS_CREDS_CONFIGURED = True
ERROR_CODES = ['AccessDenied', 'AllAccessDisabled', '[Errno 21] Is a directory:']


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

        message = "{0:>11} : {1}".format("[found]", bucket + " | " + str(size) + " Bytes | ACLs: " + str(b["acls"]))
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

    # Dump the bucket into bucket folder
    bucketDir = './buckets/' + bucketName

    dumped = None

    try:
        if not AWS_CREDS_CONFIGURED:
            sh.aws('s3', 'sync', 's3://' + bucketName, bucketDir, '--no-sign-request', _fg=False)
            dumped = True
        else:
            sh.aws('s3', 'sync', 's3://' + bucketName, bucketDir, _fg=False)
            dumped = True
    except sh.ErrorReturnCode_1 as e:
        # Loop through our list of known errors. If found, dumping failed.
        foundErr = False
        err_message = e.stderr.decode('utf-8')
        for err in ERROR_CODES:
            if err in err_message:
                foundErr = True
                break
        if foundErr:                       # We caught a known error while dumping
            if not os.listdir(bucketDir):  # The bucket directory is empty. The dump didn't work
                dumped = False
            else:                          # The bucket directory is not empty. At least 1 of the files was downloaded.
                dumped = True
        else:
            raise e

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
    s3 = boto3.resource('s3')
    try:
        if AWS_CREDS_CONFIGURED is False:
            s3.meta.client.meta.events.register('choose-signer.s3.*', disable_signing)
        bucket = s3.Bucket(bucketName)
        size_bytes = 0
        for item in bucket.objects.all():  # Note: Only does a max of 1000 items
            size_bytes += item.size
        return size_bytes

    except sh.TimeoutException:
        return "Unknown Size - timeout"
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
    """ If we find an open bucket, save the contents of the bucket listing to file. """

    # Dump the bucket into bucket folder
    bucketDir = './list-buckets/' + bucketName + '.txt'
    if not os.path.exists('./list-buckets/'):
        os.makedirs('./list-buckets/')

    try:
        if AWS_CREDS_CONFIGURED:
            sh.aws('s3', 'ls', '--recursive', 's3://' + bucketName, _out=bucketDir)
        else:
            sh.aws('s3', 'ls', '--recursive', '--no-sign-request', 's3://' + bucketName, _out=bucketDir)
    except sh.ErrorReturnCode_255 as e:
        if "AccessDenied" in e.stderr.decode("utf-8"):
            return "AccessDenied"
        else:
            raise e
