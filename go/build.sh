#!/bin/bash

set -e

export GO111MODULE="on"

functions=(
  hello_world
  deposit
  deposit_in_provider
  stream_consumer
)


rm -rf build/*

compile_and_zip() {
  func=$1

  cd src

  echo Building $func...
  cmdout=$(GOARCH=amd64 GOOS=linux go build -o ../build/$func/bootstrap $func/main.go 2>&1)
  es=$?

  if (($es)); then
    echo -e "\n[$func] error $es:\n$cmdout\n"
    exit 255
  else
    cd ..
    zip -qjT9 build/$func/main.zip build/$func/bootstrap
    rm build/$func/bootstrap
  fi
}

export -f compile_and_zip

if [[ "$(uname)" == "Darwin" ]]; then
  CORES=$(sysctl -n hw.ncpu)
else
  CORES=$(nproc --all)
fi

printf %s\\n ${functions[@]} | xargs -n 1 -P $CORES -I {} bash -c 'compile_and_zip $@' _ {}
