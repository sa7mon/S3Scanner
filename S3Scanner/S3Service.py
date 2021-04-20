"""
    This will be a service that the client program will instantiate to then call methods
    passing buckets
"""
from boto3 import client  # TODO: Limit import to just boto3.client, probably
from S3Scanner.S3Bucket import S3Bucket, BucketExists, Permission, S3BucketObject
from botocore.exceptions import ClientError
import botocore.session
from botocore import UNSIGNED
from botocore.client import Config
import datetime
from S3Scanner.exceptions import AccessDeniedException, InvalidEndpointException, BucketMightNotExistException
from os.path import normpath
import pathlib
from concurrent.futures import ThreadPoolExecutor, as_completed
from functools import partial
from urllib3 import disable_warnings

ALL_USERS_URI = 'uri=http://acs.amazonaws.com/groups/global/AllUsers'
AUTH_USERS_URI = 'uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers'


class S3Service:
    def __init__(self, forceNoCreds=False, endpoint_url='https://s3.amazonaws.com', verify_ssl=True,
                 endpoint_address_style='path'):
        """
        Service constructor

        :param forceNoCreds: (Boolean) Force the service to not use credentials, even if the user has creds configured
        :param endpoint_url: (String) URL of S3 endpoint to use. Must include http(s):// scheme
        :param verify_ssl: (Boolean) Whether of not to verify ssl. Set to false if endpoint is http
        :param endpoint_address_style: (String) Addressing style of the endpoint. Must be 'path' or 'vhost'
        :returns None
        """
        self.endpoint_url = endpoint_url
        self.endpoint_address_style = 'path' if endpoint_address_style == 'path' else 'virtual'
        use_ssl = True if self.endpoint_url.startswith('http://') else False

        if not verify_ssl:
            disable_warnings()

        # DEBUG
        # boto3.set_stream_logger(name='botocore')

        # Validate the endpoint if it's not the default of AWS
        if self.endpoint_url != 'https://s3.amazonaws.com':
            if not self.validate_endpoint_url(use_ssl, verify_ssl, endpoint_address_style):
                raise InvalidEndpointException(message=f"Endpoint '{self.endpoint_url}' does not appear to be S3-compliant")

        # Check for AWS credentials
        session = botocore.session.get_session()
        if forceNoCreds or session.get_credentials() is None or session.get_credentials().access_key is None:
            self.aws_creds_configured = False
            self.s3_client = client('s3',
                                          config=Config(signature_version=UNSIGNED, s3={'addressing_style': self.endpoint_address_style}, connect_timeout=3,
                                         retries={'max_attempts': 2}),
                                          endpoint_url=self.endpoint_url, use_ssl=use_ssl, verify=verify_ssl)
        else:
            self.aws_creds_configured = True
            self.s3_client = client('s3', config=Config(s3={'addressing_style': self.endpoint_address_style}, connect_timeout=3,
                                         retries={'max_attempts': 2}),
                                          endpoint_url=self.endpoint_url, use_ssl=use_ssl, verify=verify_ssl)

        del session  # No longer needed

    def check_bucket_exists(self, bucket):
        """
        Checks if a bucket exists. Sets `exists` property of `bucket`

        :param S3Bucket bucket: Bucket to check
        :raises ValueError: If `bucket` is not an s3Bucket object
        :return: None
        """
        if not isinstance(bucket, S3Bucket):
            raise ValueError("Passed object was not type S3Bucket")

        bucket_exists = True

        try:
            self.s3_client.head_bucket(Bucket=bucket.name)
        except ClientError as e:
            if e.response['Error']['Code'] == '404':
                bucket_exists = False

        bucket.exists = BucketExists.YES if bucket_exists else BucketExists.NO

    def check_perm_read_acl(self, bucket):
        """
        Check for the READACP permission on `bucket` by trying to get the bucket ACL

        :param S3Bucket bucket: Bucket to check permission of
        :raises BucketMightNotExistException: If `bucket` existence hasn't been checked
        :raises ClientError: If we encounter an unexpected ClientError from boto client
        :return: None
        """

        if bucket.exists != BucketExists.YES:
            raise BucketMightNotExistException()

        try:
            bucket.foundACL = self.s3_client.get_bucket_acl(Bucket=bucket.name)
            self.parse_found_acl(bucket)  # If we can read ACLs, we know the rest of the permissions
        except ClientError as e:
            if e.response['Error']['Code'] == "AccessDenied" or e.response['Error']['Code'] == "AllAccessDisabled":
                if self.aws_creds_configured:
                    bucket.AuthUsersReadACP = Permission.DENIED
                else:
                    bucket.AllUsersReadACP = Permission.DENIED
            else:
                raise e

    def check_perm_read(self, bucket):
        """
        Checks for the READ permission on the bucket by attempting to list the objects.
        Sets the `AllUsersRead` and/or `AuthUsersRead` property of `bucket`.

        :param S3Bucket bucket: Bucket to check permission of
        :raises BucketMightNotExistException: If `bucket` existence hasn't been checked
        :raises ClientError: If we encounter an unexpected ClientError from boto client
        :return: None
        """
        if bucket.exists != BucketExists.YES:
            raise BucketMightNotExistException()

        list_bucket_perm_allowed = True
        try:
            self.s3_client.list_objects_v2(Bucket=bucket.name, MaxKeys=0)  # TODO: Compare this to doing a HeadBucket
        except ClientError as e:
            if e.response['Error']['Code'] == "AccessDenied" or e.response['Error']['Code'] == "AllAccessDisabled":
                list_bucket_perm_allowed = False
            else:
                print(f"ERROR: Error while checking bucket {bucket.name}")
                raise e
        if self.aws_creds_configured:
            # Don't mark AuthUsersRead as Allowed if it's only implicitly allowed due to AllUsersRead being allowed
            # We only want to make AuthUsersRead as Allowed if that permission is explicitly set for AuthUsers
            if bucket.AllUsersRead != Permission.ALLOWED:
                bucket.AuthUsersRead = Permission.ALLOWED if list_bucket_perm_allowed else Permission.DENIED
        else:
            bucket.AllUsersRead = Permission.ALLOWED if list_bucket_perm_allowed else Permission.DENIED

    def check_perm_write(self, bucket):
        """
        Check for WRITE permission by trying to upload an empty file to the bucket.
        File is named the current timestamp to ensure we're not overwriting an existing file in the bucket.

        NOTE: If writing to bucket succeeds using an AuthUser, only mark AuthUsersWrite as Allowed if AllUsers are
        Denied. Writing can succeed if AuthUsers are implicitly allowed due to AllUsers being allowed, but we only want
        to mark AuthUsers as Allowed if they are explicitly granted. If AllUsersWrite is Allowed and the write is
        successful by an AuthUser, we have no way of knowing if AuthUsers were granted permission explicitly

        :param S3Bucket bucket: Bucket to check permission of
        :raises BucketMightNotExistException: If `bucket` existence hasn't been checked
        :raises ClientError: If we encounter an unexpected ClientError from boto client
        :return: None
        """
        if bucket.exists != BucketExists.YES:
            raise BucketMightNotExistException()

        timestamp_file = str(datetime.datetime.now().timestamp()) + '.txt'

        try:
            # Try to create a new empty file with a key of the timestamp
            self.s3_client.put_object(Bucket=bucket.name, Key=timestamp_file, Body=b'')

            if self.aws_creds_configured:
                if bucket.AllUsersWrite != Permission.ALLOWED:  # If AllUsers have Write permission, don't mark AuthUsers as Allowed
                    bucket.AuthUsersWrite = Permission.ALLOWED
                else:
                    bucket.AuthUsersWrite = Permission.UNKNOWN
            else:
                bucket.AllUsersWrite = Permission.ALLOWED

            # Delete the temporary file
            self.s3_client.delete_object(Bucket=bucket.name, Key=timestamp_file)
        except ClientError as e:
            if e.response['Error']['Code'] == "AccessDenied" or e.response['Error']['Code'] == "AllAccessDisabled":
                if self.aws_creds_configured:
                    bucket.AuthUsersWrite = Permission.DENIED
                else:
                    bucket.AllUsersWrite = Permission.DENIED
            else:
                raise e

    def check_perm_write_acl(self, bucket):
        """
        Checks for WRITE_ACP permission by attempting to set an ACL on the bucket.
        WARNING: Potentially destructive - make sure to run this check last as it will include all discovered
        permissions in the ACL it tries to set, thus ensuring minimal disruption for the bucket owner.

        :param S3Bucket bucket: Bucket to check permission of
        :raises BucketMightNotExistException: If `bucket` existence hasn't been checked
        :raises ClientError: If we encounter an unexpected ClientError from boto client
        :return: None
        """
        if bucket.exists != BucketExists.YES:
            raise BucketMightNotExistException()

        # TODO: See if there's a way to simplify this section
        readURIs = []
        writeURIs = []
        readAcpURIs = []
        writeAcpURIs = []
        fullControlURIs = []

        if bucket.AuthUsersRead == Permission.ALLOWED:
            readURIs.append(AUTH_USERS_URI)
        if bucket.AuthUsersWrite == Permission.ALLOWED:
            writeURIs.append(AUTH_USERS_URI)
        if bucket.AuthUsersReadACP == Permission.ALLOWED:
            readAcpURIs.append(AUTH_USERS_URI)
        if bucket.AuthUsersWriteACP == Permission.ALLOWED:
            writeAcpURIs.append(AUTH_USERS_URI)
        if bucket.AuthUsersFullControl == Permission.ALLOWED:
            fullControlURIs.append(AUTH_USERS_URI)

        if bucket.AllUsersRead == Permission.ALLOWED:
            readURIs.append(ALL_USERS_URI)
        if bucket.AllUsersWrite == Permission.ALLOWED:
            writeURIs.append(ALL_USERS_URI)
        if bucket.AllUsersReadACP == Permission.ALLOWED:
            readAcpURIs.append(ALL_USERS_URI)
        if bucket.AllUsersWriteACP == Permission.ALLOWED:
            writeAcpURIs.append(ALL_USERS_URI)
        if bucket.AllUsersFullControl == Permission.ALLOWED:
            fullControlURIs.append(ALL_USERS_URI)

        if self.aws_creds_configured:   # Otherwise AWS will return "Request was missing a required header"
            writeAcpURIs.append(AUTH_USERS_URI)
        else:
            writeAcpURIs.append(ALL_USERS_URI)
        args = {'Bucket': bucket.name}
        if len(readURIs) > 0:
            args['GrantRead'] = ','.join(readURIs)
        if len(writeURIs) > 0:
            args['GrantWrite'] = ','.join(writeURIs)
        if len(readAcpURIs) > 0:
            args['GrantReadACP'] = ','.join(readAcpURIs)
        if len(writeAcpURIs) > 0:
            args['GrantWriteACP'] = ','.join(writeAcpURIs)
        if len(fullControlURIs) > 0:
            args['GrantFullControl'] = ','.join(fullControlURIs)
        try:
            self.s3_client.put_bucket_acl(**args)
            if self.aws_creds_configured:
                # Don't mark AuthUsersWriteACP as Allowed if it's due to implicit permission via AllUsersWriteACP
                # Only mark it as allowed if the AuthUsers group is explicitly allowed
                if bucket.AllUsersWriteACP != Permission.ALLOWED:
                    bucket.AuthUsersWriteACP = Permission.ALLOWED
                else:
                    bucket.AuthUsersWriteACP = Permission.UNKNOWN
            else:
                bucket.AllUsersWriteACP = Permission.ALLOWED
        except ClientError as e:
            if e.response['Error']['Code'] == "AccessDenied" or e.response['Error']['Code'] == "AllAccessDisabled":
                if self.aws_creds_configured:
                    bucket.AuthUsersWriteACP = Permission.DENIED
                else:
                    bucket.AllUsersWriteACP = Permission.DENIED
            else:
                raise e

    def dump_bucket_multithread(self, bucket, dest_directory, verbose=False, threads=4):
        """
        Takes a bucket and downloads all the objects to a local folder.
        If the object exists locally and is the same size as the remote object, the object is skipped.
        If the object exists locally and is a different size then the remote object, the local object is overwritten.

        :param S3Bucket bucket: Bucket whose contents we want to dump
        :param str dest_directory: Folder to save the objects to. Must include trailing slash
        :param bool verbose: Output verbose messages to the user
        :param int threads: Number of threads to use while dumping
        :return: None
        """
        # TODO: Let the user choose whether or not to overwrite local files if different

        print(f"{bucket.name} | Dumping contents using 4 threads...")
        func = partial(self.download_file, dest_directory, bucket, verbose)

        with ThreadPoolExecutor(max_workers=threads) as executor:
            futures = {
                executor.submit(func, obj): obj for obj in bucket.objects
            }

            for future in as_completed(futures):
                if future.exception():
                    print(f"{bucket.name} | Download failed: {futures[future]}")

        print(f"{bucket.name} | Dumping completed")

    def download_file(self, dest_directory, bucket, verbose, obj):
        """
        Download `obj` from `bucket` into `dest_directory`

        :param str dest_directory: Directory to store the object into
        :param S3Bucket bucket: Bucket to download the object from
        :param bool verbose: Output verbose messages to the user
        :param S3BucketObject obj: Object to downlaod
        :return: None
        """
        dest_file_path = pathlib.Path(normpath(dest_directory + obj.key))
        if dest_file_path.exists():
            if dest_file_path.stat().st_size == obj.size:
                if verbose:
                    print(f"{bucket.name} | Skipping {obj.key} - already downloaded")
                return
            else:
                if verbose:
                    print(f"{bucket.name} | Re-downloading {obj.key} - local size differs from remote")
        else:
            if verbose:
                print(f"{bucket.name} | Downloading {obj.key}")
        dest_file_path.parent.mkdir(parents=True, exist_ok=True)  # Equivalent to `mkdir -p`
        self.s3_client.download_file(bucket.name, obj.key, str(dest_file_path))

    def enumerate_bucket_objects(self, bucket):
        """
        Enumerate all the objects in a bucket. Sets the `BucketSize`, `objects`, and `objects_enumerated` properties
        of `bucket`.

        :param S3Bucket bucket: Bucket to enumerate objects of
        :raises Exception: If the bucket doesn't exist
        :raises AccessDeniedException: If we are denied access to the bucket objects
        :raises ClientError: If we encounter an unexpected ClientError from boto client
        :return: None
        """
        if bucket.exists == BucketExists.UNKNOWN:
            self.check_bucket_exists(bucket)
        if bucket.exists == BucketExists.NO:
            raise Exception("Bucket doesn't exist")

        try:
            for page in self.s3_client.get_paginator("list_objects_v2").paginate(Bucket=bucket.name):
                if 'Contents' not in page:  # No items in this bucket
                    bucket.objects_enumerated = True
                    return
                for item in page['Contents']:
                    obj = S3BucketObject(key=item['Key'], last_modified=item['LastModified'], size=item['Size'])
                    bucket.add_object(obj)
        except ClientError as e:
            if e.response['Error']['Code'] == "AccessDenied" or e.response['Error']['Code'] == "AllAccessDisabled":
                raise AccessDeniedException("AccessDenied while enumerating bucket objects")
        bucket.objects_enumerated = True

    def parse_found_acl(self, bucket):
        """
        Translate ACL grants into permission properties. If we were able to read the ACLs, we should be able to skip
        manually checking most permissions

        :param S3Bucket bucket: Bucket whose ACLs we want to parse
        :return: None
        """
        if bucket.foundACL is None:
            return

        if 'Grants' in bucket.foundACL:
            for grant in bucket.foundACL['Grants']:
                if grant['Grantee']['Type'] == 'Group':
                    if 'URI' in grant['Grantee'] and grant['Grantee']['URI'] == 'http://acs.amazonaws.com/groups/global/AuthenticatedUsers':
                        # Permissions have been given to the AuthUsers group
                        if grant['Permission'] == 'FULL_CONTROL':
                            bucket.AuthUsersRead = Permission.ALLOWED
                            bucket.AuthUsersWrite = Permission.ALLOWED
                            bucket.AuthUsersReadACP = Permission.ALLOWED
                            bucket.AuthUsersWriteACP = Permission.ALLOWED
                            bucket.AuthUsersFullControl = Permission.ALLOWED
                        elif grant['Permission'] == 'READ':
                            bucket.AuthUsersRead = Permission.ALLOWED
                        elif grant['Permission'] == 'READ_ACP':
                            bucket.AuthUsersReadACP = Permission.ALLOWED
                        elif grant['Permission'] == 'WRITE':
                            bucket.AuthUsersWrite = Permission.ALLOWED
                        elif grant['Permission'] == 'WRITE_ACP':
                            bucket.AuthUsersWriteACP = Permission.ALLOWED

                    elif 'URI' in grant['Grantee'] and grant['Grantee']['URI'] == 'http://acs.amazonaws.com/groups/global/AllUsers':
                        # Permissions have been given to the AllUsers group
                        if grant['Permission'] == 'FULL_CONTROL':
                            bucket.AllUsersRead = Permission.ALLOWED
                            bucket.AllUsersWrite = Permission.ALLOWED
                            bucket.AllUsersReadACP = Permission.ALLOWED
                            bucket.AllUsersWriteACP = Permission.ALLOWED
                            bucket.AllUsersFullControl = Permission.ALLOWED
                        elif grant['Permission'] == 'READ':
                            bucket.AllUsersRead = Permission.ALLOWED
                        elif grant['Permission'] == 'READ_ACP':
                            bucket.AllUsersReadACP = Permission.ALLOWED
                        elif grant['Permission'] == 'WRITE':
                            bucket.AllUsersWrite = Permission.ALLOWED
                        elif grant['Permission'] == 'WRITE_ACP':
                            bucket.AllUsersWriteACP = Permission.ALLOWED

            # All permissions not explicitly granted in the ACL are denied
            # TODO: Simplify this
            if bucket.AuthUsersRead == Permission.UNKNOWN:
                bucket.AuthUsersRead = Permission.DENIED

            if bucket.AuthUsersWrite == Permission.UNKNOWN:
                bucket.AuthUsersWrite = Permission.DENIED

            if bucket.AuthUsersReadACP == Permission.UNKNOWN:
                bucket.AuthUsersReadACP = Permission.DENIED

            if bucket.AuthUsersWriteACP == Permission.UNKNOWN:
                bucket.AuthUsersWriteACP = Permission.DENIED

            if bucket.AuthUsersFullControl == Permission.UNKNOWN:
                bucket.AuthUsersFullControl = Permission.DENIED

            if bucket.AllUsersRead == Permission.UNKNOWN:
                bucket.AllUsersRead = Permission.DENIED

            if bucket.AllUsersWrite == Permission.UNKNOWN:
                bucket.AllUsersWrite = Permission.DENIED

            if bucket.AllUsersReadACP == Permission.UNKNOWN:
                bucket.AllUsersReadACP = Permission.DENIED

            if bucket.AllUsersWriteACP == Permission.UNKNOWN:
                bucket.AllUsersWriteACP = Permission.DENIED

            if bucket.AllUsersFullControl == Permission.UNKNOWN:
                bucket.AllUsersFullControl = Permission.DENIED

    def validate_endpoint_url(self, use_ssl=True, verify_ssl=True, endpoint_address_style='path'):
        """
        Verify the user-supplied endpoint URL is S3-compliant by trying to list a maximum of 0 keys from a bucket which
        is extremely unlikely to exist.

        Note: Most S3-compliant services will return an error code of "NoSuchBucket". Some services which require auth
        for most operations (like Minio) will return an error code of "AccessDenied" instead

        :param bool use_ssl: Whether or not the endpoint serves HTTP over SSL
        :param bool verify_ssl: Whether or not to verify the SSL connection.
        :param str endpoint_address_style: Addressing style of endpoint. Must be either 'path' or 'vhost'
        :return: bool: Whether or not the server responded in an S3-compliant way
        """

        # We always want to verify the endpoint using no creds
        # so if the s3_client has creds configured, make a new anonymous client

        addressing_style = 'virtual' if endpoint_address_style == 'vhost' else 'path'

        validation_client = client('s3', config=Config(signature_version=UNSIGNED,
                                         s3={'addressing_style': addressing_style}, connect_timeout=3,
                                         retries={'max_attempts': 0}), endpoint_url=self.endpoint_url, use_ssl=use_ssl,
                                         verify=verify_ssl)

        non_existent_bucket = 's3scanner-' + str(datetime.datetime.now())[0:10]
        try:
            validation_client.list_objects_v2(Bucket=non_existent_bucket, MaxKeys=0)
        except ClientError as e:
            if (e.response['Error']['Code'] == 'NoSuchBucket' or e.response['Error']['Code'] == 'AccessDenied') and \
                    'BucketName' in e.response['Error']:
                return True
            return False
        except botocore.exceptions.ConnectTimeoutError:
            return False

        # If we get here, the bucket either existed (unlikely) or the server returned a 200 for some reason
        return False
