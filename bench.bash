#!/bin/bash -e

export GOPATH="$GOPATH:`pwd`"
export PATH="$PATH:`pwd`/bin"

TEST_MODULES="`find ./src -name '*_test.go' | sed 's#^\./src/##' | sed 's#/[^/]\+$##' | sort | uniq | while read p; do echo -n \"$p \"; done`"
go test -bench '.*' -benchtime 15s $TEST_MODULES

cd slave_build
make bench
cd - > /dev/null
