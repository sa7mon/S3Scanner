import sh
import requests
import os


sizeCheckTimeout = 8    # How long to wait for getBucketSize to return


def checkBucket(bucketName, region):
    """ Does a simple GET request with the Requests library and interprets the results.

    site - A domain name without protocol (http[s])
    region - An s3 region. See: https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region  """

    if len(bucketName) < 3 or len(bucketName) > 63:  # Bucket names can be 3-63 (inclusively) characters long.
        return 999, bucketName

    bucketDomain = 'http://' + bucketName + '.s3-' + region + '.amazonaws.com'

    try:
        r = requests.head(bucketDomain)
    except requests.exceptions.ConnectionError:  # Couldn't resolve the hostname. Definitely not a bucket.
        message = "{0:>16} : {1}".format("[not found]", bucketName)
        return 900, message
    if r.status_code == 200:    # Successfully found a bucket!
        size = getBucketSize(bucketName)
        return 200, bucketName, region, size

    elif r.status_code == 301:  # We tried the wrong region. 'x-amz-bucket-region' header will give us the correct one.
        return 301, r.headers['x-amz-bucket-region']

    elif r.status_code == 403:  # Bucket exists, but we're not allowed to LIST it.
        return 403, bucketName, region
    elif r.status_code == 404:  # This is definitely not a valid bucket name.
        message = "{0:>16} : {1}".format("[not found]", bucketName)
        return 404, message
    else:
        raise ValueError("Got an unhandled status code back: " + str(r.status_code) + " for site: " + bucketName + ":" + region)


def dumpBucket(bucketName, region):

    # Check to make sure the bucket is open
    b = checkBucket(bucketName, region)
    if b[0] != 200:
        raise ValueError("The specified bucket is not open.")

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
        a = sh.aws('s3', 'ls', '--summarize', '--human-readable', '--recursive', '--no-sign-request', 's3://' +
                   bucketName, _timeout=sizeCheckTimeout)
    except sh.TimeoutException:
        return "Unknown Size"
    # Get the last line of the output, get everything to the right of the colon, and strip whitespace
    return a.splitlines()[len(a.splitlines())-1].split(":")[1].strip()

