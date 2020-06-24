
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


def test_checkBucket():
    """
    checkBucket.1 - Bucket name
    checkBucket.2 - Domain name
    checkBucket.3 - Full s3 url
    checkBucket.4 - bucket:region
    """

    print("test_checkBucket temporarily disabled.")

    # test_setup()
    #
    # testFile = './test/test_checkBucket.txt'
    #
    # # Create file logger
    # flog = logging.getLogger('s3scanner-file')
    # flog.setLevel(logging.DEBUG)              # Set log level for logger object
    #
    # # Create file handler which logs even debug messages
    # fh = logging.FileHandler(testFile)
    # fh.setLevel(logging.DEBUG)
    #
    # # Add the handler to logger
    # flog.addHandler(fh)
    #
    # # Create secondary logger for logging to screen
    # slog = logging.getLogger('s3scanner-screen')
    # slog.setLevel(logging.CRITICAL)
    #
    # try:
    #     # checkBucket.1
    #     s3.checkBucket("flaws.cloud", slog, flog, False, False)
    #
    #     # checkBucket.2
    #     s3.checkBucket("flaws.cloud.s3-us-west-2.amazonaws.com", slog, flog, False, False)
    #
    #     # checkBucket.3
    #     s3.checkBucket("flaws.cloud:us-west-2", slog, flog, False, False)
    #
    #     # Read in test loggin file and assert
    #     f = open(testFile, 'r')
    #     results = f.readlines()
    #     f.close()
    #
    #     assert results[0].rstrip() == "flaws.cloud"
    #     assert results[1].rstrip() == "flaws.cloud"
    #     assert results[2].rstrip() == "flaws.cloud"
    #
    # finally:
    #     # Delete test file
    #     os.remove(testFile)


def test_checkBucketName():
    """
    Scenario checkBucketName.1 - Under length requirements
        Expected: False
    Scenario checkBucketName.2 - Over length requirements
        Expected: False
    Scenario checkBucketName.3 - Contains forbidden characters
        Expected: False
    Scenario checkBucketName.4 - Blank name
        Expected: False
    Scenario checkBucketName.5 - Good name
        Expected: True
    """

    print("test_checkBucketName temporarily disabled.")

    # test_setup()
    #
    # # checkBucketName.1
    # result = s3.checkBucketName('ab')
    # assert result is False
    #
    # # checkBucketName.2
    # tooLong = "asdfasdf12834092834nMSdfnasjdfhu23y49u2y4jsdkfjbasdfbasdmn4asfasdf23423423423423"  # 80 characters
    # result = s3.checkBucketName(tooLong)
    # assert result is False
    #
    # # checkBucketName.3
    # badBucket = "mycoolbucket:dev"
    # assert s3.checkBucketName(badBucket) is False
    #
    # # checkBucketName.4
    # assert s3.checkBucketName('') is False
    #
    # # checkBucketName.5
    # assert s3.checkBucketName('arathergoodname') is True

def test_getBucketSize():
    """
    Scenario getBucketSize.1 - Public read enabled
        Expected: The s3scanner-bucketsize bucket returns size: 43 bytes
    Scenario getBucketSize.2 - Public read disabled
        Expected: app-dev bucket has public read permissions disabled
    Scenario getBucketSize.3 - Bucket doesn't exist
        Expected: We should get back "NoSuchBucket"
    Scenario getBucketSize.4 - Public read enabled, more than 1,000 objects
        Expected: The s3scanner-long bucket returns size: 3900 bytes
    """
    print("test_getBucketSize temporarily disabled.")

    # test_setup()
    # # getBucketSize.1
    # assert s3.getBucketSize('s3scanner-bucketsize') == "43 bytes"
    #
    # # getBucketSize.2
    # assert s3.getBucketSize('s3scanner-private') == "AccessDenied"
    #
    # # getBucketSize.3
    # assert s3.getBucketSize('thiswillprobablynotexistihope') == "NoSuchBucket"
    #
    # # getBucketSize.4
    # assert s3.getBucketSize('s3scanner-long') == "4000 bytes"


def test_getBucketSizeTimeout():
    """
    Scenario getBucketSize.1 - Too many files to list so it times out
        Expected: The function returns a timeout error after the specified wait time
        Note: Verify that getBucketSize returns an unknown size and doesn't take longer
        than sizeCheckTimeout set in s3utils
    """
    print("test_getBucketSizeTimeout temporarily disabled.")

    # test_setup()

    # s3.AWS_CREDS_CONFIGURED = False
    # s3.SIZE_CHECK_TIMEOUT = 2  # In case we have a fast connection

    # output = s3.getBucketSize("s3scanner-long")

    # # Assert that the size check timed out
    # assert output == "Unknown Size - timeout"


def test_listBucket():
    """
    Scenario listBucket.1 - Public read enabled
        Expected: Listing bucket flaws.cloud will create the directory, create flaws.cloud.txt, and write the listing to file
    Scenario listBucket.2 - Public read disabled
    Scenario listBucket.3 - Public read enabled, long listing

    """
    print("test_listBucket temporarily disabled.")

    # test_setup()
    #
    # # listBucket.1
    #
    # listFile = './list-buckets/s3scanner-bucketsize.txt'
    #
    # s3.listBucket('s3scanner-bucketsize')
    #
    # assert os.path.exists(listFile)                         # Assert file was created in the correct location
    #
    # lines = []
    # with open(listFile, 'r') as g:
    #     for line in g:
    #         lines.append(line)
    #
    # assert lines[0].rstrip().endswith('test-file.txt')      # Assert the first line is correct
    # assert len(lines) == 1                                  # Assert number of lines in the file is correct
    #
    # # listBucket.2
    # if s3.AWS_CREDS_CONFIGURED:
    #     assert s3.listBucket('s3scanner-private') == None
    # else:
    #     assert s3.listBucket('s3scanner-private') == "AccessDenied"
    #
    # # listBucket.3
    # longFile = './list-buckets/s3scanner-long.txt'
    # s3.listBucket('s3scanner-long')
    # assert os.path.exists(longFile)
    #
    # lines = []
    # with open(longFile, 'r') as f:
    #     for line in f:
    #         lines.append(f)
    # assert len(lines) == 3501


def test_dumpBucket():
    """
    Scenario dumpBucket.1 - Public read permission enabled
        Expected: Dumping the bucket "flaws.cloud" should result in 6 files being downloaded into the buckets folder.
                    The expected file sizes of each file are listed in the 'expectedFiles' dictionary.
    Scenario dumpBucket.2 - Public read objects disabled
        Expected: The function returns false and the bucket directory doesn't exist
    Scenario dumpBucket.3 - Authenticated users read enabled, public users read disabled
        Expected: The function returns true and the bucket directory exists. Opposite for if no aws creds are set
    """
    print("test_dumpBucket temporarily disabled.")

    # test_setup()

    # # dumpBucket.1
    #
    # s3.dumpBucket("flaws.cloud")
    #
    # dumpDir = './buckets/flaws.cloud/'  # Folder to look for the files in
    #
    # # Expected sizes of each file
    # expectedFiles = {'hint1.html': 2575, 'hint2.html': 1707, 'hint3.html': 1101, 'index.html': 3082,
    #                  'robots.txt': 46, 'secret-dd02c7c.html': 1051, 'logo.png': 15979}
    #
    # try:
    #     # Assert number of files in the folder
    #     assert len(os.listdir(dumpDir)) == len(expectedFiles)
    #
    #     # For each file, assert the size
    #     for file, size in expectedFiles.items():
    #         assert os.path.getsize(dumpDir + file) == size
    # finally:
    #     # No matter what happens with the asserts, cleanup after the test by deleting the flaws.cloud directory
    #     shutil.rmtree(dumpDir)
    #
    # # dumpBucket.2
    # assert s3.dumpBucket('s3scanner-private') is s3.AWS_CREDS_CONFIGURED
    # assert os.path.exists('./buckets/s3scanner-private') is False
    #
    # # dumpBucket.3
    # assert s3.dumpBucket('s3scanner-auth') is s3.AWS_CREDS_CONFIGURED  # Asserts should both follow whether or not creds are set
    # assert os.path.exists('./buckets/s3scanner-auth') is False

