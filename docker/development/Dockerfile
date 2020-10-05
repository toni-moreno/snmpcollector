FROM ubuntu:16.04 as build
WORKDIR /app/src/snmpcollector

RUN apt update -y && apt upgrade -y
RUN apt-get install -y locales
RUN apt-get install -y  curl git gcc
RUN curl -sL https://deb.nodesource.com/setup_8.x -o nodesource_setup.sh
RUN bash nodesource_setup.sh
RUN apt-get install -y nodejs
RUN curl -O https://storage.googleapis.com/golang/go1.12.1.linux-amd64.tar.gz
RUN tar -xvf go1.12.1.linux-amd64.tar.gz -C /usr/local
RUN rm go1.12.1.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
ENV GOROOT=/usr/local/go
ENV GOPATH=/app
ENV PATH=$PATH:$GOROOT/bin
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH
COPY ./setup.sh .
RUN bash setup.sh 
COPY  . .
RUN mv docker/start-dev.sh /

ENTRYPOINT [ "/start-dev.sh" ]

