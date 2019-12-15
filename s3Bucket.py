import re
from enum import Enum


class Permission(Enum):
    ALLOWED = 1,
    DENIED = 0,
    UNKNOWN = -1


class BucketExists(Enum):
    YES = 1,
    NO = 0,
    UNKNOWN = -1


def bytes_to_human_readable(bytes_size, suffix='B'):
    """
        Shamelessly copied from: https://stackoverflow.com/a/1094933/2307994
    """
    for unit in ['', 'K', 'M', 'G', 'T', 'P', 'E', 'Z']:
        if abs(bytes_size) < 1024.0:
            return "%3.1f%s%s" % (bytes_size, unit, suffix)
        bytes_size /= 1024.0
    return "%.1f%s%s" % (bytes_size, 'Yi', suffix)


class s3BucketObject:
    """
        Represents a file stored in a bucket.
        __eq__ and __hash__ are implemented to take full advantage of the set() deduplication.
    """
    def __init__(self, size, last_modified, key):
        self.size = size
        self.last_modified = last_modified
        self.key = key

    def __eq__(self, other):
        return self.key == other.key

    def __hash__(self):
        return hash(self.key)

    def getHumanReadableSize(self):
        return bytes_to_human_readable(self.size)


class s3Bucket:
    """
    :raises: ValueError - if bucket name is invalid
    """

    exists = BucketExists.UNKNOWN
    objects = set()
    bucketSize = 0
    objects_enumerated = False

    def __init__(self, name):
        check = self.__checkBucketName(name)
        if not check['valid']:
            raise ValueError("Invalid bucket name")
        
        self.name = check['name']

        self.PermListBucket = Permission.UNKNOWN
        # self.PermHeadBucket = Permission.UNKNOWN -- part of "deep-dive". Unlikely to be different than listBucket
        self.PermGetBucketAcl = Permission.UNKNOWN
        # self.PutObject = Permission.UNKNOWN
        # self.PutBucketAcl = Permission.UNKNOWN

    def __checkBucketName(self, name):
        """ Checks to make sure bucket names input are valid according to S3 naming conventions
        :param name: Name of bucket to check
        :return: Dict
                ['valid'] - Boolean - whether or not the name is valid
                ['name'] - string - Bucket name extracted from the input
        """

        bucket_name = ""
        # Check if bucket name is valid and determine the format
        if ".amazonaws.com" in name:    # We were given a full s3 url
            bucket_name = name[:name.rfind(".s3")]
        elif ":" in name:               # We were given a bucket in 'bucket:region' format
            bucket_name = name.split(":")[0]
        else:                               # We were given a regular bucket name
            bucket_name = name

        # Bucket names can be 3-63 (inclusively) characters long.
        # Bucket names may only contain lowercase letters, numbers, periods, and hyphens
        pattern = r'(?=^.{3,63}$)(?!^(\d+\.)+\d+$)(^(([a-z0-9]|[a-z0-9][a-z0-9\-]*[a-z0-9])\.)*([a-z0-9]|[a-z0-9][a-z0-9\-]*[a-z0-9])$)'
        return {'valid': bool(re.match(pattern, bucket_name)), 'name': bucket_name}

    def addObject(self, obj):
        self.objects.add(obj)
        self.bucketSize += obj.size

    def getHumanReadableSize(self):
        return bytes_to_human_readable(self.bucketSize)
