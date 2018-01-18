import sh


def getBucketSize(bucketName):
    """ Use awscli to 'ls' the bucket which will give us the total size of the bucket."""

    a = sh.aws('s3', 'ls', '--summarize', '--human-readable', '--recursive', '--no-sign-request','s3://' + bucketName)

    # Get the last line of the output, get everything to the right of the colon, and strip whitespace
    return a.splitlines()[len(a.splitlines())-1].split(":")[1].strip()