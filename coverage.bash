#!/bin/bash -e

export GOPATH="$GOPATH:`pwd`"

TEST_MODULES="`find ./src -name '*_test.go' | sed 's#^\./src/##' | sed 's#/[^/]\+$##' | sort | uniq | while read p; do echo -n \"$p \"; done`"
./cover.bash $TEST_MODULES
