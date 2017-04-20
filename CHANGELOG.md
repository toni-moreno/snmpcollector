# v 0.7.0 (not released yet)
### New Features
* added a gonsmp fix for snmpv3 performance problems.
* updated all dependencies
* added overwrite option to the config import dialog

### fixes
* fix for #169, #172, #173, #177, #181, #182

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
