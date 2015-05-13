#!/bin/bash -e

source ./colors.bash

PLUGINS=()

printHelp() {
	echo "------------------"
	echo "  bootstrap.bash  "
	echo "------------------"
	echo
	echo "Options:"
	echo
	echo "  --version \"v0.0.0\" - Vesion"
	echo "  --plugin \"pnlugin_package\" - Add plugin to build"
	echo "  -h | --help - Show this help"
	echo
	exit 1
}

while [[ $# > 0 ]]
do
	key="$1"
	shift
	case $key in
		--version)
			VERSION="$1"
			shift
		;;
		--plugin)
			PLUGINS+=("$1")
			shift
		;;
		-h|--help)
			printHelp
		;;
		*)
			# unknown option
		;;
	esac
done

echo "${bold}${magenta}Generating Go code...${normal}"
protoc --go_out=src/tilerender slave/proto/*.proto
for file in `ls thrift/*.thrift`
do
	thrift -gen go -I thrift -out src $file
done
go-bindata -o src/gopnikwebstatic/bindata.go  -pkg "gopnikwebstatic" -prefix "src/gopnikwebstatic/" src/gopnikwebstatic/public/fonts/ src/gopnikwebstatic/public/css/ src/gopnikwebstatic/public/css/images src/gopnikwebstatic/public/js/
go-bindata -o src/gopnikprerender/bindata.go  -prefix "src/gopnikprerender/" src/gopnikprerender/templates/
go-bindata -o src/gopnikperf/bindata.go  -prefix "src/gopnikperf/" src/gopnikperf/templates/
go-bindata -o src/sampledata/bindata.go -pkg "sampledata"  -prefix "sampledata_tiles" sampledata_tiles/

cat << EOF > src/sampledata/env.go
package sampledata

const Stylesheet = "`pwd`/sampledata/stylesheet.xml"
const MapnikInputPlugins = "`mapnik-config --input-plugins`"
var SlaveCmd = []string{"`pwd`/bin/gopnikslave",
		"-stylesheet", Stylesheet,
		"-pluginsPath", MapnikInputPlugins}
EOF

echo "${bold}${magenta}Configuring plugins...${normal}"
for p in `ls -d ./src/defplugins/* | egrep -o 'defplugins/[a-z]+'`
do
	PLUGINS+=("$p")
done
PLUGINS_CONFIG="src/plugins_enabled/config.go"
PLUGINS_TEST_CONFIG="src/plugins_enabled/config_test.go"
cat << EOF > $PLUGINS_CONFIG
package plugins_enabled

import (
EOF
for p in ${PLUGINS[@]}; do
	echo -e "\t_ \"$p\"" >> $PLUGINS_CONFIG
done

echo ')' >> $PLUGINS_CONFIG

echo '[plugins]'

cat << EOF > $PLUGINS_TEST_CONFIG
package plugins_enabled

import (
	. "gopkg.in/check.v1"
)

EOF

for p in src/defplugins/*; do
 TEST_FILE="$p/test.json"
 if [ -f "$TEST_FILE" ]
 then
 PNAME=`jq --raw-output '.Plugin' "$TEST_FILE"`
cat << EOF >> $PLUGINS_TEST_CONFIG
func (s *StorePluginSuite) Test$PNAME(c *C) {
	cfg := \``cat "$TEST_FILE"`\`
	s.GetSetTest(c, cfg)
}
EOF
 fi
done

echo '[plugin tests]'

echo "${bold}${magenta}Setup version...${normal}"
VERSION_CONFIG="src/program_version/version.go"
if [[ "x$VERSION" != "x" ]]; then
cat << EOF > $VERSION_CONFIG
package program_version

func init() {
	version = "$VERSION"
	publishVersion()
}
EOF
fi

echo "${bold}${magenta}Configuring C++ code...${normal}"
[ -d slave_build ] || mkdir slave_build
cd slave_build
cmake ../slave
cd - > /dev/null

echo "${bold}${green}Done${normal}"
