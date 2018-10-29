# SnmpCollector [![Go Report Card](https://goreportcard.com/badge/snmpcollector)](https://goreportcard.com/report/snmpcollector)

SnmpCollector is a full featured Generic SNMP data collector with Web Administration Interface Open Source tool which has as main goal simplify  the configuration for getting data from any  device witch snmp protocol support and send resulting data to an influxdb backend.

For complete information on installation from  binary package and configuration you could  read the [snmpcollector wiki](https://snmpcollector/wiki).

If you wish to compile from source code you can follow the next steps

## Run from master
If you want to build a package yourself, or contribute. Here is a guide for how to do that.

### Dependencies

- Go 1.5
- NodeJS >=6.2.1

### Get Code

```bash
go get -d snmpcollector/...
```

### Building the backend


```bash
cd $GOPATH/src/snmpcollector
go run build.go setup            # only needed once to install godep
godep restore                    # will pull down all golang lib dependencies in your current GOPATH
```

### Building frontend and backend in production mode

```bash
npm install
PATH=$(npm bin):$PATH            # or export PATH=$(npm bin):$PATH depending on your shell
npm run build:prod               # will build fronted and backend
```

### Creating minimal package tar.gz

After building frontend and backend you wil do

```bash
npm run postbuild #will build fronted and backend
```

### Creating rpm and deb packages

you  will need previously installed the fpm/rpm and deb packaging tools.
After building frontend and backend  you will do.

```bash
go run build.go latest
```

### Running first time
To execute without any configuration you need a minimal config.toml file on the conf directory.

```bash
cp conf/sample.config.toml conf/config.toml
./bin/snmpcollector
```

This will create a default user with username *adm1* and password *adm1pass* (don't forget to change them!).

### Recompile backend on source change (only for developers)

To rebuild on source change (requires that you executed godep restore)
```bash
go get github.com/Unknwon/bra
npm start
```
will init a change autodetect webserver with angular-cli (ng serve) and also a autodetect and recompile process with bra for the backend


#### Online config

Now you wil be able to configure metrics/measuremnets and devices from the builting web server at  http://localhost:8090 or http://localhost:4200 if working in development mode (npm start)

### Offline configuration.

You will be able also insert data directly on the sqlite db that snmpcollector has been created at first execution on config/snmpcollector.db examples on example_config.sql

```
cat conf/example_config.sql |sqlite3 conf/snmpcollector.db
```
