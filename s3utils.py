import sh
import os
import boto3
import requests

sizeCheckTimeout = 8    # How long to wait for getBucketSize to return
awsCredsConfigured = True

client = boto3.client('s3')


def checkAcl(bucket):
    allUsersGrants = []
    authUsersGrants = []

    s3 = boto3.resource('s3')

    try:
        bucket_acl = s3.BucketAcl(bucket)
        bucket_acl.load()
    except client.exceptions.NoSuchBucket:
        return {"found": False, "acls": {}}

    except client.exceptions.ClientError as e:
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

    Returns: True if AWS credentials are properly configured. False if not.
    """
    try:
        sh.aws('sts', 'get-caller-identity', '--output', 'text', '--query', 'Account')
    except sh.ErrorReturnCode_255 as e:
        if "Unable to locate credentials" in e.stderr.decode("utf-8"):
            return False
        else:
            raise e

    return True


def checkBucketName(bucketName):
    if (len(bucketName) < 3) or (len(bucketName) > 63):  # Bucket names can be 3-63 (inclusively) characters long.
        return False

    for char in bucketName:  # Bucket names can contain letters, numbers, periods, and hyphens
        if char.lower() not in "abcdefghijklmnopqrstuvwxyz0123456789.-":
            return False

    return True


def checkBucketWithoutCreds(bucketName):
    """ Does a simple GET request with the Requests library and interprets the results.
    site - A domain name without protocol (http[s]) """

    bucketUrl = 'http://' + bucketName + '.s3.amazonaws.com'

    r = requests.head(bucketUrl)

    if r.status_code == 200:    # Successfully found a bucket!
        return True
    elif r.status_code == 403:  # Bucket exists, but we're not allowed to LIST it.
        return True
    elif r.status_code == 404:  # This is definitely not a valid bucket name.
        return False
    else:
        raise ValueError("Got an unhandled status code back: " + str(r.status_code) + " for site: " + bucketName)


def dumpBucket(bucketName):

    # Dump the bucket into bucket folder
    bucketDir = './buckets/' + bucketName
    if not os.path.exists(bucketDir):
        os.makedirs(bucketDir)

    sh.aws('s3', 'sync', 's3://'+bucketName, bucketDir, '--no-sign-request', _fg=True)

    # Check if folder is empty. If it is, delete it
    if not os.listdir(bucketDir):
        # Delete empty folder
        os.rmdir(bucketDir)


def getBucketSize(bucketName):
    """
    Use awscli to 'ls' the bucket which will give us the total size of the bucket.
    NOTE:
        Function assumes the bucket exists and doesn't catch errors if it doesn't.
    """
    try:
        if awsCredsConfigured:
            a = sh.aws('s3', 'ls', '--summarize', '--human-readable', '--recursive', 's3://' +
                       bucketName, _timeout=sizeCheckTimeout)
        else:
            a = sh.aws('s3', 'ls', '--summarize', '--human-readable', '--recursive', '--no-sign-request',
                       's3://' + bucketName, _timeout=sizeCheckTimeout)
        # Get the last line of the output, get everything to the right of the colon, and strip whitespace
        return a.splitlines()[len(a.splitlines()) - 1].split(":")[1].strip()
    except sh.TimeoutException:
        return "Unknown Size - timeout"
    except sh.ErrorReturnCode_255 as e:
        if "AccessDenied" in e.stderr.decode("UTF-8"):
            return "AccessDenied"
        elif "AllAccessDisabled" in e.stderr.decode("UTF-8"):
            return "AllAccessDisabled"
        else:
            raise e


def listBucket(bucketName):
    """ If we find an open bucket, save the contents of the bucket listing to file. """

    # Dump the bucket into bucket folder
    bucketDir = './list-buckets/' + bucketName + '.txt'
    if not os.path.exists('./list-buckets/'):
        os.makedirs('./list-buckets/')

    try:
        if awsCredsConfigured:
            sh.aws('s3', 'ls', '--recursive', 's3://' + bucketName, _out=bucketDir)
        else:
            sh.aws('s3', 'ls', '--recursive', '--no-sign-request', 's3://' + bucketName, _out=bucketDir)
    except sh.ErrorReturnCode_255:
        raise ValueError("Bucket doesn't seem open.")

