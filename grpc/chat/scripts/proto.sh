#!/usr/bin/env bash

set -e

SRCDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJECT=$(realpath $SRCDIR/..)
PROTO_SRC=$PROJECT/service
PROTO_OUT=~/go/src

#
# Compile proto files
#

compile() {
  for file in $(find $PROTO_SRC -name '*.proto' -type f);
  do
    echo Compiling ...$(basename $file)
    protoc -I=$PROTO_SRC \
            --go-grpc_out=$PROTO_OUT \
            --go_out=$PROTO_OUT $file
  done
}

rm -rf $PROJECT/protobuf/*
compile

