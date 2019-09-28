import s3utils as s3
import os
import sys
import shutil
import time
import logging
import subprocess


pyVersion = sys.version_info  # pyVersion[0] can be 2 or 3


s3scannerLocation = "./"
testingFolder = "./test/"

setupRan = False


def test_setup():
    """ Setup code to run before we run all tests. """
    global setupRan

    if setupRan:    # We only need to run this once per test-run
        return

    # Check if AWS creds are configured
    s3.AWS_CREDS_CONFIGURED = s3.checkAwsCreds()

    print("--> AWS credentials configured: " + str(s3.AWS_CREDS_CONFIGURED))

    # Create testingFolder if it doesn't exist
    if not os.path.exists(testingFolder) or not os.path.isdir(testingFolder):
        os.makedirs(testingFolder)

    setupRan = True


def test_arguments():
    """
    Scenario mainargs.1: No args supplied
    Scenario mainargs.2: --out-file
    Scenario mainargs.3: --list
    Scenario mainargs.4: --dump
    """

    test_setup()

    # mainargs.1
    a = subprocess.run(['python3', s3scannerLocation + 's3scanner.py'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert a.stderr == b'usage: s3scanner [-h] [-o OUTFILE] [-d] [-l] [--version] buckets\ns3scanner: error: the following arguments are required: buckets\n'

    # mainargs.2

    # Put one bucket into a new file
    with open(testingFolder + "mainargs.2_input.txt", "w") as f:
        f.write('s3scanner-bucketsize\n')

    try:
        a = subprocess.run(['python3', s3scannerLocation + 's3scanner.py', '--out-file', testingFolder + 'mainargs.2_output.txt',
                        testingFolder + 'mainargs.2_input.txt'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        with open(testingFolder + "mainargs.2_output.txt") as f:
            line = f.readline().strip()

        assert line == 's3scanner-bucketsize'

    finally:
        # No matter what happens with the test, clean up the test files at the end
        try:
            os.remove(testingFolder + 'mainargs.2_output.txt')
            os.remove(testingFolder + 'mainargs.2_input.txt')
        except OSError:
            pass

    # mainargs.3
    # mainargs.4


def test_checkAcl():
    """
    Scenario checkAcl.1 - ACL listing enabled
        Expected:
            found = True
            acls = {'allUsers': ['READ', 'READ_ACP'], 'authUsers': ['READ', 'READ_ACP']}
    Scenario checkAcl.2 - AccessDenied for ACL listing
        Expected:
            found = True
            acls = 'AccessDenied'
    Scenario checkAcl.3 - Bucket access is disabled
        Expected:
            found = True
            acls = "AllAccessDisabled"
    Scenario checkAcl.4 - Bucket doesn't exist
        Expected:
            found = False
            acls = {}
    """
    test_setup()

    if not s3.AWS_CREDS_CONFIGURED:  # Don't run tests if AWS creds aren't configured
        return

    # checkAcl.1
    r1 = s3.checkAcl('aneta')
    assert r1["found"] is True
    assert r1["acls"] == {'allUsers': ['READ', 'READ_ACP'], 'authUsers': ['READ', 'READ_ACP']}

    # checkAcl.2
    result = s3.checkAcl('flaws.cloud')
    assert result["found"] is True
    assert result["acls"] == "AccessDenied"

    # checkAcl.3
    result = s3.checkAcl('amazon.com')
    assert result["found"] is True
    assert result["acls"] == "AllAccessDisabled"

    # checkAcl.4
    result = s3.checkAcl('hopethisdoesntexist1234asdf')
    assert result["found"] is False
    assert result["acls"] == {}


def test_checkAwsCreds():
    """
    Scenario checkAwsCreds.1 - Output of checkAwsCreds() matches a more intense check for creds
    """
    test_setup()

    # Check more thoroughly for creds being set.
    vars = os.environ

    keyid = vars.get("AWS_ACCESS_KEY_ID")
    key = vars.get("AWS_SECRET_ACCESS_KEY")
    credsFile = os.path.expanduser("~") + "/.aws/credentials"

    if keyid is not None and len(keyid) == 20:
        if key is not None and len(key) == 40:
            credsActuallyConfigured = True
        else:
            credsActuallyConfigured = False
    else:
        credsActuallyConfigured = False

    if os.path.exists(credsFile):
        print("credsFile path exists")
        if not credsActuallyConfigured:
            keyIdSet = None
            keySet = None

            # Check the ~/.aws/credentials file
            with open(credsFile, "r") as f:
                for line in f:
                    line = line.strip()
                    if line[0:17].lower() == 'aws_access_key_id':
                        if len(line) >= 38:  # key + value = length of at least 38 if no spaces around equals
                            keyIdSet = True
                        else:
                            keyIdSet = False

                    if line[0:21].lower() == 'aws_secret_access_key':
                        if len(line) >= 62:
                            keySet = True
                        else:
                            keySet = False

            if keyIdSet and keySet:
                credsActuallyConfigured = True

    # checkAwsCreds.1
    assert s3.checkAwsCreds() == credsActuallyConfigured

def test_checkBucket():
    """
    checkBucket.1 - Bucket name
    checkBucket.2 - Domain name
    checkBucket.3 - Full s3 url
    checkBucket.4 - bucket:region
    """
    
    test_setup()
    
    testFile = './test/test_checkBucket.txt'
    
    # Create file logger
    flog = logging.getLogger('s3scanner-file')
    flog.setLevel(logging.DEBUG)              # Set log level for logger object
    
    # Create file handler which logs even debug messages
    fh = logging.FileHandler(testFile)
    fh.setLevel(logging.DEBUG)
    
    # Add the handler to logger
    flog.addHandler(fh)
    
    # Create secondary logger for logging to screen
    slog = logging.getLogger('s3scanner-screen')
    slog.setLevel(logging.CRITICAL)
    
    try:
        # checkBucket.1
        s3.checkBucket("flaws.cloud", slog, flog, False, False)
        
        # checkBucket.2
        s3.checkBucket("flaws.cloud.s3-us-west-2.amazonaws.com", slog, flog, False, False)
        
        # checkBucket.3
        s3.checkBucket("flaws.cloud:us-west-2", slog, flog, False, False)
        
        # Read in test loggin file and assert
        f = open(testFile, 'r')
        results = f.readlines()
        f.close()
        
        assert results[0].rstrip() == "flaws.cloud"
        assert results[1].rstrip() == "flaws.cloud"
        assert results[2].rstrip() == "flaws.cloud"
        
    finally:
        # Delete test file
        os.remove(testFile)

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
    test_setup()

    # checkBucketName.1
    result = s3.checkBucketName('ab')
    assert result is False

    # checkBucketName.2
    tooLong = "asdfasdf12834092834nMSdfnasjdfhu23y49u2y4jsdkfjbasdfbasdmn4asfasdf23423423423423"  # 80 characters
    result = s3.checkBucketName(tooLong)
    assert result is False

    # checkBucketName.3
    badBucket = "mycoolbucket:dev"
    assert s3.checkBucketName(badBucket) is False

    # checkBucketName.4
    assert s3.checkBucketName('') is False

    # checkBucketName.5
    assert s3.checkBucketName('arathergoodname') is True


def test_checkBucketWithoutCreds():
    """
    Scenario checkBucketwc.1 - Non-existent bucket
    Scenario checkBucketwc.2 - Good bucket
    Scenario checkBucketwc.3 - No public read perm
    """
    test_setup()

    if s3.AWS_CREDS_CONFIGURED:
        return

    # checkBucketwc.1
    assert s3.checkBucketWithoutCreds('ireallyhopethisbucketdoesntexist') is False

    # checkBucketwc.2
    assert s3.checkBucketWithoutCreds('flaws.cloud') is True

    # checkBucketwc.3
    assert s3.checkBucketWithoutCreds('blog') is True


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
    test_setup()

    # dumpBucket.1

    s3.dumpBucket("flaws.cloud")

    dumpDir = './buckets/flaws.cloud/'  # Folder to look for the files in

    # Expected sizes of each file
    expectedFiles = {'hint1.html': 2575, 'hint2.html': 1707, 'hint3.html': 1101, 'index.html': 3082,
                     'robots.txt': 46, 'secret-dd02c7c.html': 1051, 'logo.png': 15979}

    try:
        # Assert number of files in the folder
        assert len(os.listdir(dumpDir)) == len(expectedFiles)

        # For each file, assert the size
        for file, size in expectedFiles.items():
            assert os.path.getsize(dumpDir + file) == size
    finally:
        # No matter what happens with the asserts, cleanup after the test by deleting the flaws.cloud directory
        shutil.rmtree(dumpDir)

    # dumpBucket.2
    assert s3.dumpBucket('s3scanner-private') is s3.AWS_CREDS_CONFIGURED
    assert os.path.exists('./buckets/s3scanner-private') is False

    # dumpBucket.3
    assert s3.dumpBucket('s3scanner-auth') is s3.AWS_CREDS_CONFIGURED  # Asserts should both follow whether or not creds are set
    assert os.path.exists('./buckets/s3scanner-auth') is False


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
    test_setup()

    # getBucketSize.1
    assert s3.getBucketSize('s3scanner-bucketsize') == "43 bytes"

    # getBucketSize.2
    assert s3.getBucketSize('app-dev') == "AccessDenied"

    # getBucketSize.3
    assert s3.getBucketSize('thiswillprobablynotexistihope') == "NoSuchBucket"

    # getBucketSize.4
    assert s3.getBucketSize('s3scanner-long') == "4000 bytes"


def test_getBucketSizeTimeout():
    """
    Scenario getBucketSize.1 - Too many files to list so it times out
        Expected: The function returns a timeout error after the specified wait time
        Note: Verify that getBucketSize returns an unknown size and doesn't take longer
        than sizeCheckTimeout set in s3utils
    """
    test_setup()

    # s3.AWS_CREDS_CONFIGURED = False
    # s3.SIZE_CHECK_TIMEOUT = 2  # In case we have a fast connection

    # output = s3.getBucketSize("s3scanner-long")

    # # Assert that the size check timed out
    # assert output == "Unknown Size - timeout"

    print("!! Notes: test_getBucketSizeTimeout temporarily disabled.")


def test_listBucket():
    """
    Scenario listBucket.1 - Public read enabled
        Expected: Listing bucket flaws.cloud will create the directory, create flaws.cloud.txt, and write the listing to file
    Scenario listBucket.2 - Public read disabled
    Scenario listBucket.3 - Public read enabled, long listing

    """
    test_setup()

    # listBucket.1

    listFile = './list-buckets/s3scanner-bucketsize.txt'

    s3.listBucket('s3scanner-bucketsize')

    assert os.path.exists(listFile)                         # Assert file was created in the correct location

    lines = []
    with open(listFile, 'r') as g:
        for line in g:
            lines.append(line)

    assert lines[0].rstrip().endswith('test-file.txt')      # Assert the first line is correct
    assert len(lines) == 1                                  # Assert number of lines in the file is correct

    # listBucket.2
    if s3.AWS_CREDS_CONFIGURED:
        assert s3.listBucket('s3scanner-private') == None
    else:
        assert s3.listBucket('s3scanner-private') == "AccessDenied"

    # listBucket.3
    longFile = './list-buckets/s3scanner-long.txt'
    s3.listBucket('s3scanner-long')
    assert os.path.exists(longFile)

    lines = []
    with open(longFile, 'r') as f:
        for line in f:
            lines.append(f)
    assert len(lines) == 3501