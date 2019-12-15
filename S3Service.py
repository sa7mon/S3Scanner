"""
    This will be a service that the client program will instantiate to then call methods
    passing buckets
"""
import boto3
from s3Bucket import s3Bucket, BucketExists, Permission, s3BucketObject
from botocore.exceptions import ClientError
import botocore.session


class S3Service:
    def __init__(self):
        self.s3_client = boto3.client('s3')

        # Check for AWS credentials
        session = botocore.session.get_session()
        if session.get_credentials() is None or session.get_credentials().access_key is None:
            self.aws_creds_configured = False
        else:
            self.aws_creds_configured = True

    def check_bucket_exists(self, bucket):
        if self.aws_creds_configured:
            if not isinstance(bucket, s3Bucket):
                raise ValueError("Passed object was not type s3Bucket")

            bucket_exists = True

            try:
                self.s3_client.head_bucket(Bucket=bucket.name)
            except ClientError as e:
                if e.response['Error']['Code'] == '404':
                    bucket_exists = False

            bucket.exists = BucketExists.YES if bucket_exists else BucketExists.NO
        else:
            raise NotImplementedError("check_bucket_exists not implemented for no aws creds")
            # TODO add checking method if no aws creds

    def check_perm_read_acl(self, bucket):
        if not self.aws_creds_configured:
            raise NotImplementedError("check_perm_list_bucket not implemented for no aws creds")
            # TODO add method if no aws creds
        read_acl_perm_allowed = True
        try:
            self.s3_client.get_bucket_acl(Bucket=bucket.name)
        except ClientError as e:
            if e.response['Error']['Code'] == "AccessDenied":
                read_acl_perm_allowed = False
            else:
                raise e
        # TODO: If we can read ACLs, we know the rest of the permissions
        bucket.PermGetBucketAcl = Permission.ALLOWED if read_acl_perm_allowed is True else Permission.DENIED

    def check_perm_list_bucket(self, bucket):
        if not self.aws_creds_configured:
            raise NotImplementedError("check_perm_list_bucket not implemented for no aws creds")
            # TODO add method if no aws creds
        list_bucket_perm_allowed = True
        try:
            self.s3_client.list_objects_v2(Bucket=bucket.name, MaxKeys=0)
        except ClientError as e:
            if e.response['Error']['Code'] == "AccessDenied":
                list_bucket_perm_allowed = False
            else:
                raise e
        bucket.PermListBucket = Permission.ALLOWED if list_bucket_perm_allowed else Permission.DENIED

    def enumerate_bucket_objects(self, bucket):
        for page in self.s3_client.get_paginator("list_objects_v2").paginate(Bucket=bucket.name):
            if 'Contents' not in page:  # No items in this bucket
                bucket.objects_enumerated = True
                return
            for item in page['Contents']:
                obj = s3BucketObject(key=item['Key'], last_modified=item['LastModified'], size=item['Size'])
                bucket.addObject(obj)
        bucket.objects_enumerated = True
