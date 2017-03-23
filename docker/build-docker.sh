#!/bin/bash

set -e -x

VERSION=`cat ../package.json| grep version | awk -F':' '{print $2}'| tr -d "\", "`
COMMIT=`git rev-parse --short HEAD`
cd ..
npm run build:static
go run build.go pkg-min-tar

export VERSION
export COMMIT

cp dist/snmpcollector-${VERSION}-${COMMIT}.tar.gz docker/snmpcollector-last.tar.gz
cp conf/sample.config.toml docker/config.toml

cd docker

sudo docker build --label version="${VERSION}" --label commitid="${COMMIT}" -t tonimoreno/snmpcollector:${VERSION} -t tonimoreno/snmpcollector:latest .
rm snmpcollector-last.tar.gz
rm config.toml
