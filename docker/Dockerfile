### compile snmpcollector from github, master branch
FROM golang:1.11-alpine as builder

RUN  apk add alpine-sdk git nodejs npm
RUN  go get -d github.com/toni-moreno/snmpcollector/...

WORKDIR  $GOPATH/src/github.com/toni-moreno/snmpcollector

RUN  go run build.go setup
RUN  godep restore -v
RUN  npm install
RUN  npm run build:static
RUN  go run build.go pkg-min-tar
RUN  mkdir -p /data/dist
RUN  cp $GOPATH/src/github.com/toni-moreno/snmpcollector/dist/* /data/dist/


### build snmpcollector container to run.
FROM alpine:latest

COPY --from=builder /data/dist/* /
RUN  tar xfz /*.tar.gz
RUN  rm /*.tar.gz
RUN  cp /opt/snmpcollector/conf/sample.config.toml /opt/snmpcollector/conf/config.toml

WORKDIR /opt/snmpcollector

RUN  echo "#!/bin/sh" > /start.sh
RUN  echo "./bin/snmpcollector  > ./log/stdout.log 2> ./log/stderr.log" >> /start.sh
RUN  chmod +x /start.sh

VOLUME ["/opt/snmpcollector/conf", "/opt/snmpcollector/log"]

EXPOSE 8090

ENTRYPOINT ["/start.sh"]
