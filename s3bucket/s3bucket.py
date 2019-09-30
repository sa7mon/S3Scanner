import re
from enum import Enum
import s3utils as s3


class Permission(Enum):
	ALLOWED = 1,
	DENIED = 0,
	UNKNOWN = -1


class BucketExists(Enum):
    YES = 1,
    NO = 0,
    UNKNOWN = -1


def bytesToHumanReadable(bytes_size):
	pass


class s3BucketObject:
	def __init__(self, size, last_modified, key):
		self.size = size
		self.last_modified = last_modified
		self.key = key

	def getHumanReadableSize(self):
		return bytesToHumanReadable(self.size)


class s3Bucket:
    """
    :raises: ValueError - if bucket name is invalid
    """

    exists = BucketExists.UNKNOWN
    objects = set()
    bucketSize = 0

	def __init__(self, name):
        valid = self.__checkBucketName(bucket)
        if not valid:
            raise ValueError("Invalid bucket name")

		self.ListBucket = Permission.UNKNOWN
		# self.GetBucketAcl = Permission.UNKNOWN
		# self.HeadBucket = Permission.UNKNOWN
		# self.PutObject = Permission.UNKNOWN
		# self.PutBucketAcl = Permission.UNKNOWN


    def __checkBucketName(name):
        """ Checks to make sure bucket names input are valid according to S3 naming conventions
        :param name: Name of bucket to check
        :return: Boolean - whether or not the name is valid
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
        return bool(re.match(pattern, bucket_name))


	def addObject(obj):
		self.objects.append(obj)
		bucketSize += obj.size


	def getHumanReadableSize(self):
		return bytesToHumanReadable(self.bucketSize)
