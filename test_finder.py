import s3utils as s3
import sh


def test_getBucketSize():

    # Scenario: A non-existent bucket
    # Expected: awscli returns 255 code
    try:
        s3.getBucketSize('example-this-hopefully-wont-exist-123123123')
    except sh.ErrorReturnCode_255:
        assert True

    # Scenario: Bucket name exists, region is wrong
        # http://amazon.com.s3-us-west-1.amazonaws.com/
    # Scenario: Bucket exists, region correct,
    # Scenario: Bucket exists, region correct, listing denied
