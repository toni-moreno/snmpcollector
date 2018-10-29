
# Create snmpcollector-last.tar.gz, which includes the binary 

FROM ubuntu:16.04 as build
WORKDIR /app/src/snmpcollector
RUN apt update -y && apt upgrade -y
RUN apt-get install -y locales
RUN apt-get install -y  curl git gcc
RUN curl -sL https://deb.nodesource.com/setup_8.x -o nodesource_setup.sh
RUN bash nodesource_setup.sh
RUN apt-get install -y nodejs
RUN curl -O https://storage.googleapis.com/golang/go1.9.1.linux-amd64.tar.gz
RUN tar -xvf go1.9.1.linux-amd64.tar.gz -C /usr/local
ENV PATH=$PATH:/usr/local/go/bin
ENV GOROOT=/usr/local/go
ENV GOPATH=/app
ENV PATH=$PATH:$GOROOT/bin
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH


COPY . /app/src/snmpcollector/


RUN go run build.go setup 
# RUN go get github.com/tools/godep
# RUN godep restore

# RUN npm install
# RUN PATH=$(npm bin):$PATH
# RUN npm run build:prod

# RUN npm run build:static
RUN go run build.go build-static

RUN go run build.go pkg-min-tar



# Create the Dockerfile for the SnmpCollector

FROM alpine:latest

COPY --from=build /app/src/snmpcollector/dist/snmpcollector.tar.gz  /

RUN  tar zxvf snmpcollector.tar.gz -C .


WORKDIR /opt/snmpcollector

COPY ./docker/config.toml ./conf
COPY ./docker/start.sh /

# RUN  mkdir ./log

ENTRYPOINT ["/start.sh"]