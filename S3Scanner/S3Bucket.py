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


def bytes_to_human_readable(bytes_in, suffix='B'):
    """
    Convert number of bytes to a "human-readable" format. i.e. 1024 -> 1KB
    Shamelessly copied from: https://stackoverflow.com/a/1094933/2307994

    :param int bytes_in: Number of bytes to convert
    :param str suffix: Suffix to convert to - i.e. B/KB/MB
    :return: str human-readable string
    """
    for unit in ['', 'K', 'M', 'G', 'T', 'P', 'E', 'Z']:
        if abs(bytes_in) < 1024.0:
            return "%3.1f%s%s" % (bytes_in, unit, suffix)
        bytes_in /= 1024.0
    return "%.1f%s%s" % (bytes_in, 'Yi', suffix)


class S3BucketObject:
    """
    Represents an object stored in a bucket.
    __eq__ and __hash__ are implemented to take full advantage of the set() deduplication
    __lt__ is implemented to enable object sorting
    """
    def __init__(self, size, last_modified, key):
        self.size = size
        self.last_modified = last_modified
        self.key = key

    def __eq__(self, other):
        return self.key == other.key

    def __hash__(self):
        return hash(self.key)

    def __lt__(self, other):
        return self.key < other.key

    def __repr__(self):
        return "Key: %s, Size: %s, LastModified: %s" % (self.key, str(self.size), str(self.last_modified))

    def get_human_readable_size(self):
        return bytes_to_human_readable(self.size)


class S3Bucket:
    """
    Represents a bucket which holds objects
    """
    exists = BucketExists.UNKNOWN
    objects = set()  # A collection of S3BucketObject
    bucketSize = 0
    objects_enumerated = False
    foundACL = None

    def __init__(self, name):
        """
        Constructor method

        :param str name: Name of bucket
        :raises ValueError: If bucket name is invalid according to `_check_bucket_name()`
        """
        check = self._check_bucket_name(name)
        if not check['valid']:
            raise ValueError("Invalid bucket name")
        
        self.name = check['name']

        self.AuthUsersRead = Permission.UNKNOWN
        self.AuthUsersWrite = Permission.UNKNOWN
        self.AuthUsersReadACP = Permission.UNKNOWN
        self.AuthUsersWriteACP = Permission.UNKNOWN
        self.AuthUsersFullControl = Permission.UNKNOWN

        self.AllUsersRead = Permission.UNKNOWN
        self.AllUsersWrite = Permission.UNKNOWN
        self.AllUsersReadACP = Permission.UNKNOWN
        self.AllUsersWriteACP = Permission.UNKNOWN
        self.AllUsersFullControl = Permission.UNKNOWN

    def _check_bucket_name(self, name):
        """
        Checks to make sure bucket names input are valid according to S3 naming conventions

        :param str name: Name of bucket to check
        :return: dict: ['valid'] - bool: whether or not the name is valid, ['name'] - str: extracted bucket name
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

    def add_object(self, obj):
        """
        Adds object to bucket. Updates the `objects` and `bucketSize` properties of the bucket

        :param S3BucketObject obj: Object to add to bucket
        :return: None
        """
        self.objects.add(obj)
        self.bucketSize += obj.size

    def get_human_readable_size(self):
        return bytes_to_human_readable(self.bucketSize)

    def get_human_readable_permissions(self):
        """
        Returns a human-readable string of allowed permissions for this bucket
        i.e. "AuthUsers: [Read | WriteACP], AllUsers: [FullControl]"

        :return: str: Human-readable permissions
        """
        # Add AuthUsers permissions
        authUsersPermissions = []
        if self.AuthUsersFullControl == Permission.ALLOWED:
            authUsersPermissions.append("FullControl")
        else:
            if self.AuthUsersRead == Permission.ALLOWED:
                authUsersPermissions.append("Read")
            if self.AuthUsersWrite == Permission.ALLOWED:
                authUsersPermissions.append("Write")
            if self.AuthUsersReadACP == Permission.ALLOWED:
                authUsersPermissions.append("ReadACP")
            if self.AuthUsersWriteACP == Permission.ALLOWED:
                authUsersPermissions.append("WriteACP")
        # Add AllUsers permissions
        allUsersPermissions = []
        if self.AllUsersFullControl == Permission.ALLOWED:
            allUsersPermissions.append("FullControl")
        else:
            if self.AllUsersRead == Permission.ALLOWED:
                allUsersPermissions.append("Read")
            if self.AllUsersWrite == Permission.ALLOWED:
                allUsersPermissions.append("Write")
            if self.AllUsersReadACP == Permission.ALLOWED:
                allUsersPermissions.append("ReadACP")
            if self.AllUsersWriteACP == Permission.ALLOWED:
                allUsersPermissions.append("WriteACP")
        return f"AuthUsers: [{', '.join(authUsersPermissions)}], AllUsers: [{', '.join(allUsersPermissions)}]"
