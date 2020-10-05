#!/bin/sh

rm ./conf/snmpcollector.db
go run build.go build



./bin/snmpcollector  > ./log/stdout.log 2> ./log/stderr.log 
