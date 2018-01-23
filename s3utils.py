import sh
import requests


def getBucketSize(bucketName):
    """
    Use awscli to 'ls' the bucket which will give us the total size of the bucket.
    NOTE:
        Function assumes the bucket exists and doesn't catch errors if it doesn't.
    """

    a = sh.aws('s3', 'ls', '--summarize', '--human-readable', '--recursive', '--no-sign-request','s3://' + bucketName)

    # Get the last line of the output, get everything to the right of the colon, and strip whitespace
    return a.splitlines()[len(a.splitlines())-1].split(":")[1].strip()


def checkBucket(bucketName, region):
    """ Does a simple GET request with the Requests library and interprets the results.

    site - A domain name without protocol (http[s])
    region - An s3 region. See: https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region  """

    bucketDomain = 'http://' + bucketName + '.s3-' + region + '.amazonaws.com'

    try:
        r = requests.get(bucketDomain)
    except requests.exceptions.ConnectionError:  # Couldn't resolve the hostname. Definitely not a bucket.
        message = "{0:>16} : {1}".format("[not found]", bucketName)
        return 900, message
    if r.status_code == 200:    # Successfully found a bucket!
        message = "{0:<7}{1:>9} : {2}".format("[found]", "[open]", bucketName
                                              + ":" + region + " - " + getBucketSize(bucketName))
        return 200, message
    elif r.status_code == 301:  # We tried the wrong region. 'x-amz-bucket-region' header will give us the correct one.
        return 301, r.headers['x-amz-bucket-region']
    elif r.status_code == 403:  # Bucket exists, but we're not allowed to LIST it.
        message = "{0:>15} : {1}".format("[found] [closed]", bucketName + ":" + region)
        return 403, message
    elif r.status_code == 404:  # This is definitely not a valid bucket name.
        message = "{0:>15} : {1}".format("[not found]", bucketName)
        return 404, message
    else:
        raise ValueError("Got an unhandled status code back: " + str(r.status_code) + " for site: " + bucketName + ":" + region)
