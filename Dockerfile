# To build and run with Docker:
#
#  $ docker build -t snmpcollector .
#  $ docker run -it --rm -p 3000:3000 -p 3001:3001 snmpcollector
#
FROM node:latest

RUN mkdir -p /snmpcollector /home/nodejs && \
    groupadd -r nodejs && \
    useradd -r -g nodejs -d /home/nodejs -s /sbin/nologin nodejs && \
    chown -R nodejs:nodejs /home/nodejs

WORKDIR /snmpcollector
COPY package.json typings.json /snmpcollector/
RUN npm install --unsafe-perm=true

COPY . /snmpcollector
RUN chown -R nodejs:nodejs /snmpcollector
USER nodejs

CMD npm start
