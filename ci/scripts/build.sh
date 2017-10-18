#!/bin/bash
export GOPATH=$(pwd)
export GOBIN=$GOPATH/bin

curl https://glide.sh/get | sh

cd src/ && glide install
go fmt 7factor.io/...
go install 7factor.io/...