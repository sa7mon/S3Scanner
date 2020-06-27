import random
import string
import boto3


class TestBucketService:
    def __init__(self):
        self.session = boto3.Session(profile_name='privileged')
        self.s3_client = self.session.client('s3')

    @staticmethod
    def generate_random_bucket_name(length=40):
        candidates = string.ascii_lowercase + string.digits
        return 's3scanner-' + ''.join(random.choice(candidates) for i in range(length))

    def delete_bucket(self, bucket_name):
        self.s3_client.delete_bucket(Bucket=bucket_name)

    def create_bucket(self, danger_bucket):
        bucket_name = self.generate_random_bucket_name()

        # For type descriptions, refer to: https://github.com/sa7mon/S3Scanner/wiki/Test-Buckets
        if danger_bucket == 1:
            self.s3_client.create_bucket(Bucket=bucket_name,
                                         GrantWrite='uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers')
            self.s3_client.put_bucket_acl(Bucket=bucket_name,
                                          GrantWrite='uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers',
                                          GrantWriteACP='uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers')
        elif danger_bucket == 2:
            self.s3_client.create_bucket(Bucket=bucket_name,
                                         GrantWrite='uri=http://acs.amazonaws.com/groups/global/AllUsers',
                                         GrantWriteACP='uri=http://acs.amazonaws.com/groups/global/AllUsers')
        elif danger_bucket == 3:
            self.s3_client.create_bucket(Bucket=bucket_name,
                                         GrantRead='uri=http://acs.amazonaws.com/groups/global/AllUsers',
                                         GrantWrite='uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers',
                                         GrantWriteACP='uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers')
        elif danger_bucket == 4:
            self.s3_client.create_bucket(Bucket=bucket_name,
                                         GrantWrite='uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers,'
                                                    'uri=http://acs.amazonaws.com/groups/global/AllUsers')
        elif danger_bucket == 5:
            self.s3_client.create_bucket(Bucket=bucket_name,
                                         GrantWriteACP='uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers,'
                                                       'uri=http://acs.amazonaws.com/groups/global/AllUsers')
        else:
            raise Exception("Unknown danger bucket type")

        return bucket_name
