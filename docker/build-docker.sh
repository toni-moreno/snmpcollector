#!/bin/bash

set -e -x

go tool dist env > /tmp/goenv.tmp
. /tmp/goenv.tmp

VERSION=`cat package.json| grep version | awk -F':' '{print $2}'| tr -d "\", "`
COMMIT=`git rev-parse --short HEAD`


if [ ! -f dist/snmpcollector-${VERSION}-${COMMIT}_${GOOS:-linux}_${GOARCH:-amd64}.tar.gz ]
then
    echo "building binary...."
    npm run build:static
    go run build.go pkg-min-tar
else
    echo "skiping build..."
fi

export VERSION
export COMMIT

cp dist/snmpcollector-${VERSION}-${COMMIT}_${GOOS:-linux}_${GOARCH:-amd64}.tar.gz docker/snmpcollector-last.tar.gz
cp conf/sample.config.toml docker/config.toml

cd docker

sudo docker build --label version="${VERSION}" --label commitid="${COMMIT}" -t tonimoreno/snmpcollector:${VERSION} -t tonimoreno/snmpcollector:latest .
rm snmpcollector-last.tar.gz
rm config.toml
rm /tmp/goenv.tmp
