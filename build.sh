#!/bin/sh
export GOPATH=$(pwd)
export GOBIN=$GOPATH/bin

# This script will build the entire project using standardized
# GoLang tooling. Make sure you've sourced the gopath first or
# this will not worky.
go fmt 7factor.io/...
go install 7factor.io/...
go test --short --cover 7factor.io/...

echo "Generating API documentation..."
if command -v aglio >/dev/null; then
    aglio -i texter.apib -o docs/api.html
else
    echo "Cannot generate documentation!"
    echo "Program aglio not found. Please install by running:"
    echo "npm -g install aglio"
fi