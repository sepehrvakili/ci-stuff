#!/bin/bash
export GOPATH=$(pwd)
export GOBIN=$GOPATH/bin

go test -short --cover 7factor.io/...