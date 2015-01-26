#!/bin/bash -e

source ./colors.bash

export GOPATH="$GOPATH:`pwd`"
export PATH="$PATH:`pwd`/bin"

echo "${bold}${magenta}Building Go code...${normal}"
BINARIES=("gopnikrender" "gopnikdispatcher" "gopnikbench" "gopnikperf")
BINARIES+=("gopnikprerender" "gopnikprerenderslave" "gopnikprerenderimport")
for target in ${BINARIES[@]}
do
	echo -n " -- $target"
	go install "$target"
	echo "  ${green}[done]${normal}"
done

echo "${bold}${magenta}Building C++ code...${normal}"
cd slave_build
make
make install
cd - > /dev/null

echo "${bold}${magenta}Running Go test code...${normal}"
TEST_MODULES="`find ./src -name '*_test.go' | sed 's#^\./src/##' | sed 's#/[^/]\+$##' | sort | uniq | while read p; do echo -n \"$p \"; done`"
go test $TEST_MODULES

echo "${bold}${magenta}Running C++ test code...${normal}"
cd slave_build
make test
cd - > /dev/null

echo "${bold}${green}Done${normal}"
