

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

