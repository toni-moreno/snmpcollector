# Golang build container
FROM golang:1.13.4-alpine

RUN apk add --no-cache gcc g++ bash git

WORKDIR $GOPATH/src/github.com/toni-moreno/snmpcollector

COPY go.mod go.sum ./

RUN go mod verify

COPY pkg pkg
COPY .git .git
COPY build.go  build.go

RUN go run build.go  build

# Node build container
FROM node:8.10.0-alpine


WORKDIR /usr/src/app/

COPY src src
COPY package.json angular-cli.json tslint.json karma.conf.js protractor.conf.js ./
RUN npm install && \
    PATH=$(npm bin):$PATH && \
    ng build --prod

#RUN find .

# Final container
FROM alpine:3.12

LABEL maintainer="Toni Moreno <toni.moreno@gmail.com>"

ARG SNMPCOL_UID="472"
ARG SNMPCOL_GID="472"

ENV PATHS_HOME="/opt/snmpcollector/"

ENV SNMPCOL_GENERAL_DATA_DIR=$PATHS_HOME/data/
ENV SNMPCOL_GENERAL_LOG_DIR=$PATHS_HOME/log/

WORKDIR $PATHS_HOME 

RUN apk add --no-cache ca-certificates bash tzdata

COPY conf/sample.config.toml ./conf/config.toml

RUN addgroup -S -g $SNMPCOL_GID snmpcol && \
    adduser -S -u $SNMPCOL_UID -G snmpcol snmpcol && \
    chown -R snmpcol:snmpcol "$PATHS_HOME" && \
    mkdir -p "$SNMPCOL_GENERAL_DATA_DIR" && \
    mkdir -p "$SNMPCOL_GENERAL_LOG_DIR" && \
    chown -R snmpcol:snmpcol "$SNMPCOL_GENERAL_DATA_DIR" && \
    chmod -R 777 "$PATHS_HOME"


COPY --from=0 /go/src/github.com/toni-moreno/snmpcollector/bin/snmpcollector ./bin/
COPY --from=1 /usr/src/app/public ./public

EXPOSE 8090

USER snmpcol

ENTRYPOINT [ "./bin/snmpcollector" ]
