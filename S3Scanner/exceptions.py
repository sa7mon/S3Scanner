class AccessDeniedException(Exception):
    def __init__(self, message):
        pass
        # Call the base class constructor
        # super().__init__(message, None)

        # Now custom code
        # self.errors = errors


class InvalidEndpointException(Exception):
    def __init__(self, message):
        self.message = message


class BucketMightNotExistException(Exception):
    def __init__(self):
        pass
