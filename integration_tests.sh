#!/bin/bash

# Rebuild our containers
docker-compose build

# This is the least invasive way to execute our integration tests. We do not want to
# put anything in the container that smells like test scripts.
docker-compose run -w "/go" --entrypoint "/bin/bash -c \"/go/bin/server & go test --cover -run .*Integration credomobile.com/texter/...\"" texter