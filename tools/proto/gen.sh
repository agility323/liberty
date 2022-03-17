#!/bin/sh

# variable
impl="gogofaster" # opts: gofast, gogofast, gogofaster, gogoslick
binary="protoc-gen-$impl"
protoc="./protoc-3.14.0-linux-x86_64"
protoc_plugins="plugins=grpc:"	# for service/rpc generation
export_path="../../lbtproto/"

# path
GOPATH=`go env | grep GOPATH= | sed 's/^GOPATH=//g' | awk '{print substr($0,2,length-2)}'`
GOBIN=`go env | grep GOBIN= | sed 's/^GOBIN=//g' | awk '{print substr($0,2,length-2)}'`
if [ "$GOBIN" = "" ]; then
	GOBIN="$GOPATH/bin"
fi
echo "GOPATH is: $GOPATH"
echo "GOBIN is: $GOBIN"
PATH=$PATH:$GOBIN

# install gen binaries
#go get github.com/gogo/protobuf/proto
#go install github.com/gogo/protobuf/${binary}
#go install github.com/gogo/protobuf/gogoproto

# generate
files=`find . -name "*.proto"`
for file in $files; do
	echo "generating proto: $file"
	./$protoc -I=. --${impl}_out=${protoc_plugins}. $file
	echo "proto done: $file"
done
echo "gen exit ..."

# cp
mv *.go $export_path
