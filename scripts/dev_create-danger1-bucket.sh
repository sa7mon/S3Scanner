#!/usr/bin/env bash

# Time to wait before deleting the dangerous bucket (in seconds)
TIMEOUT=10

# Generate a random alphanumeric (lower-case) 50 character string for a new bucket name
BUCKET_NAME=$(tr -dc 'a-z0-9' < /dev/urandom | fold -w 50 | head -n 1)

# Create the bucket, giving Write and WriteACP privileges to AuthenticatedUsers
aws s3api create-bucket --profile privileged --bucket "$BUCKET_NAME" --grant-write uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers \
--grant-write-acp uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers

# Pause the script for x seconds
echo "Bucket will be deleted in $TIMEOUT seconds. Press enter to delete now: "
read -t $TIMEOUT -r answer

echo "Deleting bucket..."
aws s3api delete-bucket --profile privileged --bucket "$BUCKET_NAME"