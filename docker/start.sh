#!/bin/sh

if [ "$APP_ENV" = "development" ]
then
    echo "MODE: Developement"
    if [  -f "./conf/snmpcollector.db" ]
        then 
            rm ./conf/snmpcollector.db
        fi
fi


if ! [ -x "./bin/snmpcollector" ]
then
    go run build.go build
    echo "Status: build complete"
fi



./bin/snmpcollector  > ./log/stdout.log 2> ./log/stderr.log 
