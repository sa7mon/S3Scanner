import sh
import requests
import os
import subprocess

sizeCheckTimeout = 8    # How long to wait for getBucketSize to return


def checkBucket(bucketName, region):
    """ Does a simple GET request with the Requests library and interprets the results.

    site - A domain name without protocol (http[s])
    region - An s3 region. See: https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region  """

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

        # Check if we can list the bucket
        try: 
            output = subprocess.check_output("aws s3 ls s3://" + bucketName, shell=True, stderr=subprocess.STDOUT)
        except subprocess.CalledProcessError as e:
            return 403, bucketName, region

        if not "An error occured (" in output:
            return 200, bucketName, region, "0"

        return 403, bucketName, region
    elif r.status_code == 404:  # This is definitely not a valid bucket name.
        message = "{0:>15} : {1}".format("[not found]", bucketName)
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

def listBucket(bucketName, region):

    # Check to make sure the bucket is open
    b = checkBucket(bucketName, region)
    if b[0] != 200:
        raise ValueError("The specified bucket is not open.")

    # Dump the bucket into bucket folder
    bucketDir = './list-buckets/' + bucketName + '.txt'
    if not os.path.exists('./list-buckets/'):
        os.makedirs('./list-buckets/')

    try: 
        output = subprocess.check_output("aws s3 ls --recursive --no-sign-request s3://" + bucketName, shell=True, stderr=subprocess.STDOUT)
    except subprocess.CalledProcessError as e:
        raise ValueError("The specified bucket is not open.")

    if not "An error occured (" in output:
        f = open(bucketDir, 'w')
        f.write(bucketName + '\r\n')
        f.write(output)
        f.close()
    else:
        raise ValueError("The specified bucket is not open.")


    


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

