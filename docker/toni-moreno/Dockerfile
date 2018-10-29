

FROM ubuntu:16.04

COPY ../ /app
RUN apt-get update && apt-get install -y locales
RUN apt-get install -y golang curl 
RUN curl -sL https://deb.nodesource.com/setup_8.x -o nodesource_setup.sh
RUN bash nodesource_setup.sh
RUN apt-get install -y nodejs
RUN npm install
RUN PATH=$(npm bin):$PATH           
RUN npm run build:prod 


FROM alpine:latest

MAINTAINER Toni Moreno <toni.moreno@gmail.com>

ADD ./docker/snmpcollector-last.tar.gz /

VOLUME ["/opt/snmpcollector/conf", "/opt/snmpcollector/log"]

EXPOSE 8090

WORKDIR /opt/snmpcollector
COPY ./docker/config.toml ./conf
COPY ./docker/start.sh /

ENTRYPOINT ["/start.sh"]
