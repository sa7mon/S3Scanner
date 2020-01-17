import random
import string
from time import sleep
import boto3

TIMEOUT = 3600


def generate_random_bucket_name(length=50):
    candidates = string.ascii_lowercase + string.digits
    return ''.join(random.choice(candidates) for i in range(length))


bucket_name = generate_random_bucket_name(50)
session = boto3.Session(profile_name='privileged')
s3_client = session.client('s3')

print("Creating bucket: " + bucket_name)

# Create bucket
s3_client.create_bucket(Bucket=bucket_name, GrantWrite='uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers',
                        GrantWriteACP='uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers')

print("Waiting " + str(TIMEOUT) + " seconds to delete the bucket. Hit Ctrl-C to delete now:")

try:
    for i in range(0, TIMEOUT):
        sleep(1)
except KeyboardInterrupt:
    pass
finally:
    print("Deleting bucket...")
    s3_client.delete_bucket(Bucket=bucket_name)