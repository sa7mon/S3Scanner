
"""
s3Bucket.py methods to test:

- s3BucketObject
    - init()
    - getHumanReadableSize()
- s3Bucket
    - init()
        - Test with invalid name
        - Verify all permissions default to Unknown
        - Verify name gets set
    - __checkBucketName()
        - Test all naming options
    - addObject()
        - Verify that the length of the bucket objects increases
        - Verify that the bucketSize increases
    - getHumanReadableSize()
    - getHumanReadablePermissions()
"""
