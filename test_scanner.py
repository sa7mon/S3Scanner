

def test_arguments():
    """
    Scenario mainargs.1: No args supplied
    Scenario mainargs.2: --out-file
    Scenario mainargs.3: --list
    Scenario mainargs.4: --dump
    """

    print("test_checkAcl temporarily disabled.")

    # test_setup()
    #
    # # mainargs.1
    # a = subprocess.run(['python3', s3scannerLocation + 's3scanner.py'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    # assert a.stderr == b'usage: s3scanner [-h] [-o OUTFILE] [-d] [-l] [--version] buckets\ns3scanner: error: the following arguments are required: buckets\n'
    #
    # # mainargs.2
    #
    # # Put one bucket into a new file
    # with open(testingFolder + "mainargs.2_input.txt", "w") as f:
    #     f.write('s3scanner-bucketsize\n')
    #
    # try:
    #     a = subprocess.run(['python3', s3scannerLocation + 's3scanner.py', '--out-file', testingFolder + 'mainargs.2_output.txt',
    #                     testingFolder + 'mainargs.2_input.txt'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    #
    #     with open(testingFolder + "mainargs.2_output.txt") as f:
    #         line = f.readline().strip()
    #
    #     assert line == 's3scanner-bucketsize'
    #
    # finally:
    #     # No matter what happens with the test, clean up the test files at the end
    #     try:
    #         os.remove(testingFolder + 'mainargs.2_output.txt')
    #         os.remove(testingFolder + 'mainargs.2_input.txt')
    #     except OSError:
    #         pass

    # mainargs.3
    # mainargs.4


def test_check_aws_creds():
    """
    Scenario checkAwsCreds.1 - Output of checkAwsCreds() matches a more intense check for creds
    """
    print("test_checkAwsCreds temporarily disabled.")

    # test_setup()
    #
    # # Check more thoroughly for creds being set.
    # vars = os.environ
    #
    # keyid = vars.get("AWS_ACCESS_KEY_ID")
    # key = vars.get("AWS_SECRET_ACCESS_KEY")
    # credsFile = os.path.expanduser("~") + "/.aws/credentials"
    #
    # if keyid is not None and len(keyid) == 20:
    #     if key is not None and len(key) == 40:
    #         credsActuallyConfigured = True
    #     else:
    #         credsActuallyConfigured = False
    # else:
    #     credsActuallyConfigured = False
    #
    # if os.path.exists(credsFile):
    #     print("credsFile path exists")
    #     if not credsActuallyConfigured:
    #         keyIdSet = None
    #         keySet = None
    #
    #         # Check the ~/.aws/credentials file
    #         with open(credsFile, "r") as f:
    #             for line in f:
    #                 line = line.strip()
    #                 if line[0:17].lower() == 'aws_access_key_id':
    #                     if len(line) >= 38:  # key + value = length of at least 38 if no spaces around equals
    #                         keyIdSet = True
    #                     else:
    #                         keyIdSet = False
    #
    #                 if line[0:21].lower() == 'aws_secret_access_key':
    #                     if len(line) >= 62:
    #                         keySet = True
    #                     else:
    #                         keySet = False
    #
    #         if keyIdSet and keySet:
    #             credsActuallyConfigured = True
    #
    # # checkAwsCreds.1
    # assert s3.checkAwsCreds() == credsActuallyConfigured

