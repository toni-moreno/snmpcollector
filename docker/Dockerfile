FROM alpine:latest
MAINTAINER Toni Moreno <toni.moreno@gmail.com>

ADD ./snmpcollector-last.tar.gz /

VOLUME ["/opt/snmpcollector/conf", "/opt/snmpcollector/log"]

EXPOSE 8090

WORKDIR /opt/snmpcollector
COPY ./config.toml ./conf/
COPY ./start.sh /

ENTRYPOINT ["/start.sh"]
