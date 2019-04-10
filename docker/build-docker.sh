#!/bin/bash

#note you need to run this from the base directory, not the docker directory

set -e -x

COMMIT=`curl --silent https://github.com/toni-moreno/snmpcollector |grep -A1 commit-tease-sha | grep -v '<' |awk -F ' ' '{print $1}'`
VERSION=`curl -L --silent https://github.com/toni-moreno/snmpcollector/raw/master/package.json | grep version | awk -F':' '{print $2}'| tr -d "\", "`

cd docker

sudo docker build --label version="${VERSION}" --label commitid="${COMMIT}" -t tonimoreno/snmpcollector:${VERSION} -t tonimoreno/snmpcollector:latest .
