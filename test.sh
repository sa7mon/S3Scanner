#!/bin/bash

# Test 1 - No arguments.
#   Expected: Exit code 2

test1=$(python s3finder.py)
test1_status=$?

if [ ${test1_status} -eq 2 ]
then
  echo "Test 1 succeeded"
else
  echo "Test 1 failed. Script call had exit code of: "${test1_status} >&2
  exit 1
fi