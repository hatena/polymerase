#!/bin/bash

RED="\033[0;31m"
GREEN="\033[0;32m"
NC="\033[0m"

# kill and remove any running containers
cleanup () {
  docker-compose -f docker-compose.test.yml -p ci kill
  docker-compose -f docker-compose.test.yml -p ci rm -f --all
}

# TODO: Change project root dir

# catch unexpected failures, do cleanup and output an error message
trap 'cleanup ; printf "${RED}Tests Failed For Unexpected Reasons${NC}\n"' \
  HUP INT QUIT PIPE TERM

# build and run the composed services
if docker-compose -f docker-compose.test.yml -p ci build; then
  echo -e "${GREEN}Docker Compose Build Passed${NC}"
else
  echo -e "${RED}Docker Compose Build Failed${NC}"
  exit 1
fi

if docker-compose -f docker-compose.test.yml -p ci up -d; then
  echo -e "${GREEN}Docker Compose Up Passed${NC}"
else
  echo -e "${RED}Docker Compose Up Failed${NC}"
  exit 1
fi

# wait for the test service to complete and grab the exit code
TEST_EXIT_CODE=$(docker wait ci_sut_1)

# output the logs for the test (for clarity)
docker logs ci_sut_1

# inspect the output of the test and display respective message
if [ -z "${TEST_EXIT_CODE+x}" ] || [ "$TEST_EXIT_CODE" -ne 0 ] ; then
  echo -e "${RED}Tests Failed${NC} - Exit Code: $TEST_EXIT_CODE"
else
  echo -e "${GREEN}Tests Passed${NC}"
fi

# call the cleanup function
cleanup

# exit the script with the same code as the test service code
exit "${TEST_EXIT_CODE}"
