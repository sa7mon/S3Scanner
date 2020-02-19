import random
import string
from time import sleep
import boto3
import os

TIMEOUT = 3600


def generate_random_bucket_name(length=50):
    candidates = string.ascii_lowercase + string.digits
    return ''.join(random.choice(candidates) for i in range(length))


def delete_bucket():
    global bucket_name
    global s3_client
    s3_client.delete_bucket(Bucket=bucket_name)
    print("Bucket "+bucket_name+" deleted.")
    os.environ.pop(bucket_name)


bucket_name = generate_random_bucket_name(50)
session = boto3.Session(profile_name='privileged')
s3_client = session.client('s3')

print("Creating bucket: " + bucket_name)

# Create bucket
s3_client.create_bucket(Bucket=bucket_name, GrantWrite='uri=http://acs.amazonaws.com/groups/global/AllUsers',
                        GrantWriteACP='uri=http://acs.amazonaws.com/groups/global/AllUsers')

os.environ['bucket_danger_2'] = bucket_name

print("Environment variable 'bucket_danger_2' set.")
print("Waiting " + str(TIMEOUT) + " seconds to delete the bucket. Hit Ctrl-C to delete now:")
sleep(TIMEOUT)
delete_bucket()