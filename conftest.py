####
# Pytest Configuration
####


def pytest_addoption(parser):
    parser.addoption("--do-dangerous", action="store_true",
                     help="Run all tests, including ones where buckets are created.")


def pytest_generate_tests(metafunc):
    if "do_dangerous_test" in metafunc.fixturenames:
        do_dangerous_test = True if metafunc.config.getoption("do_dangerous") else False
        print("do_dangerous_test: " + str(do_dangerous_test))
        metafunc.parametrize("do_dangerous_test", [do_dangerous_test])