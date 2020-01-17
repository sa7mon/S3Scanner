import random
import string
from time import sleep
import boto3

TIMEOUT = 15


def generate_random_bucket_name(length=50):
    candidates = string.ascii_lowercase + string.digits
    return ''.join(random.choice(candidates) for i in range(length))


def delete_bucket():
    global bucket_name
    global s3_client
    s3_client.delete_bucket(Bucket=bucket_name)


bucket_name = generate_random_bucket_name(50)
session = boto3.Session(profile_name='privileged')
s3_client = session.client('s3')

print("Creating bucket: " + bucket_name)

# Create bucket
s3_client.create_bucket(Bucket=bucket_name, GrantWrite='uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers',
                        GrantWriteACP='uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers')
export 

print("Waiting " + str(TIMEOUT) + " seconds to delete the bucket. Hit Ctrl-C to delete now:")
sleep(TIMEOUT)
delete_bucket()