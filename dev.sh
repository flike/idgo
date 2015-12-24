#!/bin/bash

export QINGTOP=$(pwd)
export QINGROOT="${QINGROOT:-${QINGTOP/\/src\/github.com\/flike\/idgo/}}"
# QINGTOP sanity check
if [[ "$QINGTOP" == "${QINGTOP/\/src\/github.com\/flike\/idgo/}" ]]; then
    echo "WARNING: QINGTOP($QINGTOP) does not contain src/github.com/flike/idgo"
    exit 1
fi

function add_path()
{
  # $1 path variable
  # $2 path to add
  if [ -d "$2" ] && [[ ":$1:" != *":$2:"* ]]; then
    echo "$1:$2"
  else
    echo "$1"
  fi
}

export GOBIN=$QINGTOP/bin

godep path > /dev/null 2>&1
if [ "$?" = 0 ]; then
    echo "GO=godep go" > build_config.mk
    export GOPATH=`godep path`
#    godep restore
else
    echo "GO=go" > build_config.mk
fi

export GOPATH="$QINGROOT"