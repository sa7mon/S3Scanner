import s3finder as s3


def test_getBucketSize():
    assert s3.getBucketSize('example-this-probably-wont-exist-123123123') == "n"

