# v 0.10.1 ( unreleased )

### New Features

* Added swagger api doc
* Now SNMPCollector URL's use hash(#) for all UI accesses ie: `http(s)://my_snmpcollector_engine/#/home`

### Fixes

* Fix #499
* Fix blocking API requests
* fix #495
* security fix for CVE-2020-12666
* security fix for CVE-2021-23337

### breaking changes

# v 0.10.0 ( 2021-03-06 )
### New Features
* added support for PostgreSQL as config database.
* added option to redirect influxdb HTTP writes over HTTP(S)_PROXY variables.
* added new MAC trasformation on  IndexTagFormat Definition [transformations](https://github.com/toni-moreno/snmpcollector/wiki/Component:-Measurements#defined-transformations)
* added support for spaces in tag values.
* Improved device stats , Added new tags "device_active", "device_connected" true/false and "active_value/connected_value" as fields with values  0/1, with this new fields/tags we can easily count % devices connected 
* Added  /api/rt/agent/shutdown as fast way to reload config when existing external tools to restart the snmpcollector instance (docker --restart=always) by example.
* added support to tags converted from int/float/bool types when is_tag is true.
* Added support for SNMPv3  AES192,AES192C,AES256,AES256C Privacy Protocols and SHA224,SHA256,SHA384,SHA512 Auth protocols.
* updated compiler platform to golang 1.16 



### fixes
* Fixed invalid field when using MULTISTRINGPARSER and void capture group. Now the value will be ommited and it won't be written on InfluxDB
* #383, #463 , #427

### breaking changes

# v 0.9.0 ( 2021-01-04 )
### New Features
* Added snmpmetric unit tests
* Updated to last "gosnmp" v1.28.0 release
* Added Mock SnmpServer and measurements unit tests
* Added [trim](https://github.com/toni-moreno/snmpcollector/wiki/Component:-SNMP-Metrics#about-octetstringhex-string) functions in octetstring based metrics (#405)
* added HTTPS support 
* Added new measurement types: multi-index and indirect multi tag to retrieve complex MIBs like [QoS](https://github.com/toni-moreno/snmpcollector/wiki/Example:-Cisco-QoS)

### fixes
* Fixed  #446
* Fixed  #405
* Fixed  #398

### breaking changes
* Measurement TagName property (only retrieved on runtime API) changed from `string` to `[]string`

# v 0.8.1 ( 2020-10-06 ) 
### New Features
* Upgraded dependencies from dep to gomodules
* Added config through environment vars ( docker friendly ),implements #420
* Added new HTTP option "listen" as substitute for "port" ( still running but deprecated ).
* Added new GENERAL option "log_mode" (file/console) to enable only main agent logs and http_access.log to be written in console. This option does not affect to the device runtime/debug logs that will be written in a file in the logdir with filename as the id of the device.
* Added new DATABASE option "log_mode" (none/file/console) to enable or disable SQL logs and choose where to write (file/console)
* New improved Dockerfile (now running as non root user) and now docker image could be built from Makefile.docker

### fixes
* Fix #379,#409,#410,#442
* Fixed lodash vulnerabily, bump to 4.17.19
* Fixed jquery vulnerability, Bump from 3.0.0 to 3.5.0

### breaking changes

* no longer supported "-httpPort" command line parameter use "-httpListen" instead (still supported but deprecated Port option at config file) 
* DOCKER IMAGE upgrade needs for previous ownership change to UID/GID=472 on all its persistent files/volumes 

# v 0.8.0 ( 2018-11-05) 
### New Features
* Added new capabilities to add , modify and delete devices online ( avoids restart the gathering ocess  on the other devices) 
* Go 1.9.8 to 1.11 binaries 
* Added a new Conversión parameter to most of the snmpmetric types.
* Implement HexString to INTEGER conversion , implements #310

### fixes
* Fix CVE-2018-3721 vulnerability

### breaking changes
* IMPORTANT!!!: OCTECSTRING  SnmpMetric type  should be manualy updated with the following SQL in the database after the upgrade, "update snmp_metric_cfg set Conversion=3 where datasrctype='OCTETSTRING';" an then restart proces or reload config.
* STRINGPARSER SnmpMetric Type no longer will be Scaled with Scale/shift values.


# v 0.7.7 (2018-05-28)
### New Features
* Added SIGTERM handler to stop gracefully all device gathering gourutines and flush pending data to its defined output db.
* Added SIGHUP handler to reload complete configuration and restart gathering process (as in the 'Reload Config' WebUI option)
* Added new Metric Type - ENUM to map values adding description to save it as STRING (thanks to https://github.com/simnv)
* Added new OID condition type - `nin` that allows multiple numerical values on conditions
* Added new internal metrics - fields_sent and fields_sent_max that retrieves the number of fields sent to InfluxDB on "selfmon_outdb_stats" measurement:

field | description
------|------------
fields_sent | number of fields sent to the DB on the last period
fields_sent_max | max number of fields sent to the DB on the last period

# v 0.7.6 (2018-01-17)
### New Features
* Go 1.8.2 to 1.9.2 binaries
* Added new value on NFR (Non Filtered Rows)  with total index length before filter applied, for STRINGEVAL metric types.
* decreased memory usage per metric in memory.
* Fix for gathering metrics from eval metrics based on oidconditionevals
* Added new statistics measurement for each output db "selfmon_outdb_stats" , these statisctics are collected on each selfmon.freq period secods and  it will send with the following fields and tags (selfmon.extratags will be also added):

field | description
------|------------
write_sent | number of HTTP writes sent to the DB (each write sends a batchPoint) on the last period
write_error | number of HTTP write errors on the period
points_sent | number of Points sent on each Write (on each BatchPoint) on the last period
points_sent_max| max number of points sent on all writes on the last period
points_sent_avg | (only if write_sent > 0) averaged points sent for all writes on the last period
write_time | sum of all HTTP response times on  all writes on the last period
write_time_max | max HTTP response time in all writes on the last period
write_time_avg | (only if write_sent > 0) average response time for all writes on the last period
		
tags | description
-----|------------
outdb| the ID for each Output DB

* Improved  selfmon_gvm statistcs added the following fields

field | description
------|------------
mem.sys| Sys is the total bytes of memory obtained from the OS.
mem.heapSys|  HeapSys is bytes of heap memory obtained from the OS.
mem.heapIdle|  HeapIdle is bytes in idle (unused) spans.
mem.heapInUse|  HeapInuse is bytes in in-use spans.
mem.heapReleased| HeapReleased is bytes of physical memory returned to the OS.
mem.heapObjects|HeapObjects is the number of allocated heap objects.
mem.stackInuse| StackInuse is bytes in stack spans.
mem.mSpanInuse| MSpanInuse is bytes of allocated mspan structures.
mem.mCacheInuse| MCacheInuse is bytes of allocated mcache structures.

* Added New MULTISTRINGPARSER metric type

### fixes

* fix for #303, #304, #307, #312

### breaking changes

* fixed typos in field names on the selfmon_device_stats measurement (thanks to https://github.com/jensenja)

 Old | New		  
  ----|----
  cicle_gather_duration| cycle_gather_duration
  cicle_gather_start_time|cycle_gather_start_time


# v 0.7.5 (2017-12-16)
### New Features
* Added Hard snmp reset option to remap all measurements again when some problem happens when trying to do snmp connection  init on first attemps. (fix #271)
* Added a new Runtime option to Force Gather data even if device is not active ( useful for configuration testing )(implements #275)If active the current gathering period won't be changed, and this will be an extra point.
* Added New Variable Catalog let us to define variable names ( and its default values if not redefined before) to use on STRINGEVAL metric types, These variables could be redefined on each device maintaining the metric formula definition across devices.(implements #99).
* Added Version Info to the Login Page and new  SnmpCollector Logo
* Pretty Web User Interface forms.
* UI migration to Angular4.
* Improved Runtime table viewer , now shows metric information on the header tooltip and only important data in each metric tooltip.
* marked as deprecated the IndexAsValue boolean configuration  parameter

### fixes

* fix for #268, #271, #272, #275, #278, #285, #292, #293

### breaking changes

* Fixed table  name field snmp_device_cfg.system_o_i_ds  to snmp_device_cfg.systemoids , to migrate data you should only execute
update snmp_device_cfg set systemoids = system_o_i_ds ;

# v 0.7.4 (2017-09-23)
### New Features
* added new Alternate SystemOID's map to extend snmpcollector to non MIB-2 based devices (as by example Proxy Squid), is important to get system description and also to check connectivity with the device.
* added new SNMPv3 parameters ContextEngineID and ContextName
* added Table Actions on InfluxDB Servers, Measurements, Measurement Filters, Measurement Groups, Custom Filters and OID Conditions
* Added Timestamp in snmp debug logs
* Updated GoSNMP Library with improved superdebug !! (https://github.com/soniah/gosnmp/pull/114)

### fixes

* fix for #247, #249, #254, #258, #256 (PR), #260, #263, #259

### breaking changes

# v 0.7.3 (2017-06-26)
### New Features
* added new BITS snmp SMI type to send named array strings as Fields or tags.
* added new BITS Check (BITSCHK) cooked type to check only a specific position.
* Add MeasFilters, MeasGroups and Extradata as available options on table actions.
* Added Table Actions on SNMP Devices, SNMPMetrics. This will allow the user to Remove several items at once and change some properties
* Added default items per page as 20
* complete runtime view refactor, a main view with a statistic and basic operation table, you can now filter also by sysDescription, and tag map. The detail view with detailed info on measuremet and field errors.
* updated gosnmp base library with our CiscoAXA and peformance fixes
* Added Github Issue Template.

### fixes

* fix for #229, #233, #238, #243

### breaking changes

# v 0.7.2 (2017-05-27)
### New Features
* added new ndif (numeric different) oid condition comparison
* added new comunication Bus, improves webui to devices message send in unicast and broadcast mode.
* added import validations.
* updated golang version to 1.8.3.
* Addedd new OID snmpmetric type.
* Runtime web-ui shows now Measurement Names and FieldNames instead of ID's.


### fixes
* fix for #211,

### breaking changes


# v 0.7.1 (2017-05-17)
### New Features
* Improved self monitoring process, renamed and added new device statistics metrics.
* Added backgound colours on runtime data to validate or invalidate data.

### fixes
* fix for #197, #203, #204, #206, #208, #209

### breaking changes
* device measurements and field names have been renamed

__measurement__

Old| New
---|-------
selfmon_rt | selfmon_device_stats

__fields__

Old | New
----|----
process_t | cicle_gather_duration
getsent | snmp_oid_get_processed
geterror | snmp_oid_get_errors

# v 0.7.0 (2017-04-29)
### New Features
* added a gonsmp fix for snmpv3 performance problems.
* updated all dependencies
* added overwrite option to the config import dialog

### fixes
* fix for #169, #172, #173, #177, #181, #182 , #184, #186, #188, #190, #193, #194, #195

### breaking changes

# v 0.6.6
### New Features
* Compilation with Go1.8
* added Docker image to hub tonimoreno/snmpcollector
* added timestamp precision parameter to the influxdb Config (changed default precision from ns to seconds)
* added influxdb connection check (influx ping) to the influxdb configuration forms
* added MaxConnections device settins for SNMP BULKWALK queries
* Added Concurrent SNMP Gathering, one GOROUTINE per measurement
* Updated gosnmp library with performance improvements
* Improved Tag Format with a new ${VAR|SELECTOR|TRANSFORMATOR} definition
* passwords now will be hidden( Thanx to @TeraHz)

### fixes
* fix for #157, #158, #161

### breaking changes

# v 0.6.5
### New Features
* Refactored and Improved form Validation options,to allow dynamic validators  on some parameters changes
* Added new Multiple OID condition Filter and filter check improvements.
* Added Import/Export methods to API REST
* Added Import/Export WebUI forms
* Added CustomFilter to Filter type REST API and WebUI

### fixes
* fix for #126, #140, #142, #145

### breaking changes


# v 0.6.4
### New Features
* Measurement Filters refactor , added CustomFilter.
* Added OID condition as new SNMP Metric Type
* Migrated OID conditions from Measurement Filter tables to its rigth place on OID condition Table (breaking change)

### fixes
* fix for #105, #107, #115, #119, #120. #123

### breaking changes
* OID Contions now are stored in a separate table in the configuration DB , data migration should be done before install this version.

```sql
-- table creation
CREATE TABLE `oid_condition_cfg` (`id` TEXT NULL, `cond_oid` TEXT NULL, `cond_type` TEXT NULL, `cond_value` TEXT NULL, `description` TEXT NULL);
CREATE UNIQUE INDEX `UQE_oid_condition_cfg_id` ON `oid_condition_cfg` (`id`);
-- oid contition data migration from meas_filter_cfg
insert into oid_condition_cfg select id,cond_oid,cond_type,cond_value,description  from meas_filter_cfg where filter_type = 'OIDCondition';
-- old table reestructuration
ALTER TABLE meas_filter_cfg  rename to meas_filter_cfg_old;
CREATE TABLE `meas_filter_cfg` (`id` TEXT NULL, `id_measurement_cfg` TEXT NULL, `filter_type` TEXT NULL,`filter_name` TEXT NULL, `enable_alias` INTEGER NULL, `description` TEXT NULL);
-- old table migration to new depending on the type
INSERT INTO meas_filter_cfg select id,id_measurement_cfg,filter_type,id,enable_alias,description from meas_filter_cfg_old where filter_type == 'OIDCondition';
INSERT INTO meas_filter_cfg select id,id_measurement_cfg,filter_type,file_name,enable_alias,description from meas_filter_cfg_old where filter_type == 'file';
INSERT INTO meas_filter_cfg select id,id_measurement_cfg,filter_type,customid,enable_alias,description from meas_filter_cfg_old where filter_type == 'CustomFilter';
DROP TABLE meas_filter_cfg_old;
CREATE UNIQUE INDEX `UQE_meas_filter_cfg_id` ON `meas_filter_cfg` (`id`);
```

# v 0.6.3
* this version have been bypassed for technical reasons

# v 0.6.2
### New Features
* Metric Type standarization according to RFC2578 SMIv2.
* new IndexTagFormat to the measurement enabling custom Tag names
* Go code big refactor and reorganization
* Added conditional send "On non zero"

### fixes
* fix for #91, #97, #100

### breaking changes

* Database measurement type changes standarized to the RFC2578 (https://tools.ietf.org/html/rfc2578#section-7.1)
```sql
update snmp_metric_cfg set datasrctype='Gauge32' where datasrctype = 'GAUGE32';
update snmp_metric_cfg set datasrctype='Gauge32' where datasrctype = 'GAUGE';
update snmp_metric_cfg set datasrctype='Integer32' where datasrctype  = 'INTEGER32';
update snmp_metric_cfg set datasrctype='OCTETSTRING' where datasrctype  = 'STRING';
update snmp_metric_cfg set datasrctype='IpAddress' where datasrctype  = 'IPADDR';
alter table influx_measurement_cfg rename to measurement_cfg;
```

# v 0.6.1
### New Features
* upgraded to angular 2.4.1/router 3.4.1/ng2-bootstrap 1.1.16-9/angular-cli 1.0.0-beta.24/ zone.js 0.7.4 /rxjs 5.0.1
* new bundle system based on angular-cli and npm
* added new indexed with indirec tag measurement type , implements #67
* added --data --pidfile --home as agent parameters
* Added  deb and rpm packaging files and option to the building process
* Default agent log set to $LOGDIR/snmpcollector.log
* Default HTTP log  set to $LOGDIR/http_access.log

### fixes
* fix for issue #81, #83 #85, #87, #90

### breaking changes

# v 0.6.0
### New Features
* new metric types based on SNMP ANS1 types
* new snmp runtime console
* improved form validations
* new string eval metric type (computed metric)
* added new "report metric" option allowing get data to operate them but not for send to the Output DB

### fixes
* fix for isue #66, #69, #61

### breaking changes

# v 0.5.6
### New features.
* UI Enhanced login aspect
* added DisableBulk option to devices with problems in bulk queries like some IBM devices
* added device process time in the selfmon metrics as selfmon_rt measurement #25
* added new "nomatch" filter condicion #55
* support for OctetString Indexes #54
* added new metric type "strigParser" #51
* added pprof options to enable debug
* added new HTTP wrapper to the WebUI.
* fixed race conditions on reload config
* removed angular2-jwt unneeded dependency

### fixes
* fix for issue #54, #45, #56

### breaking changes


# v 0.5.5
### New features.
* Online Reload Configuration
* New runtime WebUI option which enables user inspect online current gathered snmp values for all measurements on all devices, also allow interact to change logging and debug options and also deactivate device.

### fixes
* fix for issue #38, #40, #42, #46, #47

### breaking changes
* no longer needed options flags: freq, repeat, verbose ; since all this features can be changed on the WebUI

# v 0.5.4
### New features.
* added UpdateFltFreq option to periodically update the status of the tables and filters
* added Security to the Collector API
* added Scale/Shift option to all numeric metric types
* improved selfmonitor metrics in only one measurement

### fixes
* fix for issue #35, #30, #33, #31

### breaking changes
* none

# v 0.5.3
### New features.
* WebUI now shows data in tables and can be filtered.


# v 0.5.2
### New features.
* Added Runtime Version to the snmpcollector API
* Estandarized Description for all objects.
* Added timeout/ UserAgent to the influxclient

### fixes
* fix for issue #22

### breaking changes
* none


# v 0.5.1
### New features.
* Define field metric as Tag (IsTag = true) => type STRING,HWADDR,IP
* Defined measurment Indexed  "index" as value on special devices with no special Index OID.
* Device initialization  is now faster. It is done  concurrently.
* Added Active device option to disable / enable runtime Initializacion and gather data.


### fixes
* device logs now in its own log filepath
* Added missing extra-tags input on the device add form

### breaking changes
* none

# v 0.5.0
### New features.
* new WebIU with all forms to insert , update , delete objects.
* Major internal snmpdeice/measurement refactor
* added internal Influxdummy conection . This object enable work with the collector without any influxdb server installed

### fixes
* fix issues related to snmp version1
* fix issue #4
* fix issue #12

### breaking changes

* none
