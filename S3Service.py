"""
    This will be a service that the client program will instantiate to then call methods
    passing buckets
"""
import boto3
from s3Bucket import s3Bucket, BucketExists
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
            raise NotImplementedError("check_bucket_exists not implement for no aws creds")
            # TODO add checking method if no aws creds
