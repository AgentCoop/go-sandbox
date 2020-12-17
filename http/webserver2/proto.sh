#!/usr/bin/env bash

set -e

SRCDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJECT=$(realpath $SRCDIR)
PROTO_SRC=$PROJECT/api
PROTO_OUT=$PROJECT/build/gen

#
# Compile proto files
#

compile() {
  for file in $(find $PROTO_SRC -name '*.proto' -type f);
  do
    echo Compiling ...$(basename $file)
    protoc -I=$PROTO_SRC \
           # --go-grpc_out=$PROTO_OUT \
            --js_out=$PROTO_OUT $file
  done
}

#protoc --js_out=import_style=commonjs,library=foo,binary:build/gen api/service/chat.proto
protoc \
  --js_out=import_style=commonjs,binary:build/gen \
  api/type/chat.proto

protoc -I=./api/ --go-grpc_out=/home/pihpah/go/src/ --go_out=/home/pihpah/go/src/  api/service/chat.proto

protoc -I=./api/ --js_out=import_style=commonjs,binary:build api/service/chat.proto

#protoc \
#  -I=$PROTO_SRC \
#  --go-grpc_out=~/go/src \
#  --go_out=~/go/src  api/service/chat.proto

#compile