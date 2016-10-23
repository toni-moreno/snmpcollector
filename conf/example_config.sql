
DELETE from snmp_metric_cfg;
DELETE from influx_measurement_cfg;
DELETE from measurement_field_cfg;
DELETE from m_groups_cfg;
DELETE from m_groups_measurements;
DELETE from influx_cfg;
DELETE from snmp_device_cfg;
DELETE from snmp_dev_m_groups;
DELETE from meas_filter_cfg;
DELETE from snmp_dev_filters;
/*
=====================
metrics
======================
# [metrics."id"] => id for other structs
# fielname : name  measurement
# Description : little description for user easy understand
# BaseOID : the real OID we need to query
# Datasrctype : any of GAUGE,INTEGER,COUNTER32,COUNTER64
# - GAUGE:  with this type of data this tool will send  direct value( float )
# - INTEGER:  with this type of data this tool will send  direct value( int)
# - COUNTER32:  with this , this tool will compute the difference ( it takes care of  posible counter overflow)
# - COUNTER64:  with this , this tool will compute the difference ( it takes care of  posible counter overflow)

#http://www.debianadmin.com/linux-snmp-oids-for-cpumemory-and-disk-statistics.html
#http://publib.boulder.ibm.com/tividd/td/ITMFNP/SC31-6364-00/en_US/HTML/MNPADMST84.htm

*/

/*
[metrics]
[metrics."Linux_user_CPU_percent"]
  fieldname = "user"
  description = "percentage of user CPU time"
  baseOID = ".1.3.6.1.4.1.2021.11.9.0"
  Datasrctype = "INTEGER"
  IsTag = false
*/

INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift, istag, description)  VALUES ('Linux_user_CPU_percent','user','.1.3.6.1.4.1.2021.11.9.0','INTEGER',0,0.0,0.0,0,'percentage of user CPU time');
/*
[metrics."Linux_system_CPU_percent"]
  fieldname = "system"
  description = "percentage of System CPU time"
  baseOID = ".1.3.6.1.4.1.2021.11.10.0"
  Datasrctype = "INTEGER"
  IsTag = false
*/
INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift, istag, description)  VALUES ('Linux_system_CPU_percent','system','.1.3.6.1.4.1.2021.11.10.0','INTEGER',0,0.0,0.0,0,'percentage of system CPU time');

/*
[metrics."Linux_idle_CPU_percent"]
  fieldname = "idle"
  description = "percentage of System CPU time"
  baseOID = ".1.3.6.1.4.1.2021.11.11.0"
  Datasrctype = "INTEGER"
  IsTag = false
  */

INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift, istag, description)  VALUES ('Linux_idle_CPU_percent','idle','.1.3.6.1.4.1.2021.11.11.0','INTEGER',0,0.0,0.0,0,'percentage of Idle CPU time');

/*
[metrics."ifHCInOctets"]
  fieldname = "In.bytes"
  description = "Bytes In - 64-bit Counters"
  baseOID = ".1.3.6.1.2.1.31.1.1.1.6"
  DatasrcType = "COUNTER64"
  getrate = false
  IsTag = false
*/

INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift, istag, description)  VALUES ('ifHCInOctets','In.bytes','.1.3.6.1.2.1.31.1.1.1.6','COUNTER64',0,0.0,0.0,0,'Bytes In - 64-bit Counters');

/*

[metrics."ifHCOutOctets"]
  fieldname = "Out.bytes"
  description = "Bytes In - 64-bit Counters"
  baseOID = ".1.3.6.1.2.1.31.1.1.1.10"
  DatasrcType = "COUNTER64"
  getRate = false
  IsTag = false
*/
INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift, istag, description)  VALUES ('ifHCOutOctets','Out.bytes','.1.3.6.1.2.1.31.1.1.1.10','COUNTER64',0,0.0,0.0,0,'Bytes Out - 64-bit Counters');

/*
[metrics."ifHCInUcastPkts"]
  fieldname = "In.packets"
  description = "Bytes In - 64-bit Counters"
  baseOID = ".1.3.6.1.2.1.31.1.1.1.7"
  DatasrcType = "COUNTER64"
  getRate = false
  IsTag = false

*/
INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift,istag, description)  VALUES ('ifHCInUcastPkts','In.packets','.1.3.6.1.2.1.31.1.1.1.7','COUNTER64',0,0.0,0.0,0,'Packets In - 64-bit Counters');
/*
[metrics."ifHCOutUcastPkts"]
  fieldname = "Out.packets"
  description = "Packets - 64-bit Counters"
  baseOID = ".1.3.6.1.2.1.31.1.1.1.11"
  DatasrcType = "COUNTER64"
  getRate = false
  IsTag = false
*/
INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift, istag, description)  VALUES ('ifHCOutUcastPkts','Out.packets','.1.3.6.1.2.1.31.1.1.1.11','COUNTER64',0,0.0,0.0,0,'Packets Out - 64-bit Counters');
/*
[metrics."ifInOctets"]
  fieldname = "In.bytes"
  description = "Bytes In - 32-bit Counters"
  baseOID = ".1.3.6.1.2.1.2.2.1.10"
  DatasrcType = "COUNTER32"
  getRate = false
  IsTag = false
  */

INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift, istag, description)  VALUES ('ifInOctets','In.bytes','.1.3.6.1.2.1.2.2.1.10','COUNTER32',0,0.0,0.0,0,'Bytes In - 32-bit Counters');

/*
[metrics."ifOutOctets"]
  fieldname = "Out.bytes"
  description = "Bytes Out - 32-bit Counters"
  baseOID = ".1.3.6.1.2.1.2.2.1.16"
  DatasrcType = "COUNTER32"
  getRate = false
  IsTag = false
*/

INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift, istag, description)  VALUES ('ifOutOctets','Out.bytes','.1.3.6.1.2.1.2.2.1.16','COUNTER32',0,0.0,0.0,0,'Bytes Out - 32-bit Counters');

/*
[metrics."ifName"]
  fieldname = "IfName"
  description = "Tag for Name "
  baseOID = ".1.3.6.1.2.1.31.1.1.1.1"
  DatasrcType = "STRING"
  getRate = false
  IsTag = true 
*/

INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift, istag, description)  VALUES ('ifName','IfName','.1.3.6.1.2.1.31.1.1.1.1','STRING',0,0.0,0.0,1,'Tags');


/*
===========================
measurements
===========================

# [measurements."id"] => id unique for this measurement.
# name 		: name  for the measurement
# fields 	: array of metric id's to join as fields in the measurement
# getMode 	:  posible values [ "value", "indexed" ]
# 	* value: will send direct the value from these metrics.
#	* indexed: will send points for each found indexed values ( like intefaces ) => needs indexOID/indexTag parameters
# indexOID: the OID from we will get all real OID's to query data.
# indexTAG: the tag name that will be sent after data gathered.

*/

/*
[measurements]

[measurements."linux_cpu"]
 name = "linux.cpu"
 fields = [ 	"Linux_user_CPU_percent",
		"Linux_system_CPU_percent",
		"Linux_idle_CPU_percent",
	]
 getmode = "value"
*/
INSERT INTO influx_measurement_cfg (id, name, getmode, indexoid, indextag) VALUES ('linux_cpu','linux.cpu','value','','');
INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_cpu','Linux_user_CPU_percent');
INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_cpu','Linux_system_CPU_percent');
INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_cpu','Linux_idle_CPU_percent');

/*
[measurements."linux_ports"]
 name = "linux.ports"
 fields = [ 	"ifHCInOctets",
		"ifHCOutOctets",
		"ifHCInUcastPkts",
		"ifHCOutUcastPkts"
 ]
 getmode = "indexed"
 indexOID = ".1.3.6.1.2.1.31.1.1.1.1" # ifName => needed to be "STRING"
 IndexTAG = "portName"
 */
 INSERT INTO influx_measurement_cfg (id, name, getmode, indexoid, indextag) VALUES ('linux_ports','linux.ports','indexed','.1.3.6.1.2.1.31.1.1.1.1','portName');
 INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_ports','ifHCInOctets');
 INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_ports','ifHCOutOctets');
 INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_ports','ifHCInUcastPkts');
 INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_ports','ifHCOutUcastPkts');
 INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_ports','ifName');



/*Test to check issue */
INSERT INTO influx_measurement_cfg (id, name, getmode, indexoid, indextag) VALUES ('linux_ports_by_index','linux.ports','indexed','.1.3.6.1.2.1.2.2.1.1','ifindex');
INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_ports_by_index','ifHCInOctets');
INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_ports_by_index','ifHCOutOctets');
INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_ports_by_index','ifHCInUcastPkts');
INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('linux_ports_by_index','ifHCOutUcastPkts');


 /*

[measurements."network_32bits"]
 name = "32bit-ports"
 fields = [ 	"ifInOctets",
		"ifOutOctets"
 ]
 getmode = "indexed"
 indexOID = ".1.3.6.1.2.1.31.1.1.1.1" # ifName => needed to be "STRING"
 IndexTAG = "portName"
*/

INSERT INTO influx_measurement_cfg (id, name, getmode, indexoid, indextag) VALUES ('network_32bits','32bits-ports','indexed','.1.3.6.1.2.1.31.1.1.1.1','portName');
INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('network_32bits','ifInOctets');
INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('network_32bits','ifOutOctets');

/*====================================
Measurement Groups
===================================*/

/*

[getgroups."Linux"]

measurements = 	[
		"linux_cpu",
		"linux_ports",
		"network_32bits"
		]
*/

INSERT INTO m_groups_cfg (id ) VALUES ('Linux');
INSERT INTO m_groups_measurements (id_mgroup_cfg,id_measurement_cfg) VALUES ('Linux','linux_cpu');
INSERT INTO m_groups_measurements (id_mgroup_cfg,id_measurement_cfg) VALUES ('Linux','linux_ports');

INSERT INTO m_groups_cfg (id ) VALUES ('Issue_index');
INSERT INTO m_groups_measurements (id_mgroup_cfg,id_measurement_cfg) VALUES ('Issue_index','linux_cpu');
INSERT INTO m_groups_measurements (id_mgroup_cfg,id_measurement_cfg) VALUES ('Issue_index','linux_ports_by_index');
/*INSERT INTO m_groups_measurements (id_mgroup_cfg,id_measurement_cfg) VALUES ('Linux','network_32bits');*/


/*=======================================
Measurement filters
*=======================================

	"id"                 : Text identificator
	"id_measurement_cfg" : related to a measurement
  "filter_type"       : could be  "file" or "OIDCondition"
  "file_name"          : a valid filename contained in the conf/ directory (only valid if filter_type = "file")
  "enabla_alias"       : "1" if enabled "0" if not ( allow get measurement tag from alias in the file)
  "cond_oid"          : a valid OID only valid for filter_type = "OIDCondition"
  "cond_type"    : Valid conditions "eq","lt","gt","ge","le"
  "cond_value"   : the value to get the condition

*/

INSERT INTO meas_filter_cfg (id,id_measurement_cfg,filter_type,file_name,enable_alias) VALUES ('filter_ports_file_a','linux_ports','file','lp_file_filter_a.txt',1);
INSERT INTO meas_filter_cfg (id,id_measurement_cfg,filter_type,file_name,enable_alias) VALUES ('filter_ports_file_b','linux_ports','file','lp_file_filter_b.txt',1);
INSERT INTO meas_filter_cfg (id,id_measurement_cfg,filter_type,cond_oid,cond_type,cond_value) VALUES ('filter_port_if_status_up','linux_ports','OIDCondition','.1.3.6.1.2.1.2.2.1.8','neq','1');
INSERT INTO meas_filter_cfg (id,id_measurement_cfg,filter_type,cond_oid,cond_type,cond_value) VALUES ('filter_port_if_name_match_eth','linux_ports','OIDCondition','.1.3.6.1.2.1.31.1.1.1.1','match','eth.*');


/*=======================================
Influx databases configuration
========================================*/

/*

[influxdb."*"]
 host = "127.0.0.1"
 port = 8086
 db = "snmp"
 user = "snmpuser"
 password = "snmppass"
 retention = "autogen"
 */

INSERT INTO influx_cfg (id,host,port,db,user,password,retention) VALUES ('default','127.0.0.1',8086,'snmp','snmpuser','snmppass','autogen');


/*======================================
  SNMP devices
  =====================================

  #--------------------------
  #  CONETION PARAMETERS
  #--------------------------


  # host : hostname or IP (no default)
  # port : conection port (no default)
  # timeout : Timeout is the timeout for the SNMP Query in seconds
  # retries : Set the number of retries to attempt within timeout.
  # repeat : (Not used yet)

  # NOTE any device has an  any unique indentificator "ID" string in the [snmpdevice."id"] field

  #--------------------------------
  #   SNMP VERSION/AUTH PARAMETERS
  #--------------------------------

  # snmpversion 	: posible values values [ "1", "2c" , "3"] ( no default set)
  # community 	: (por snmpv2 only) commumity for authentication (no default set)
  #
  # -- for SNMP v3 only --
  #
  # v3seclevel	:  posible values [ "NoAuthNoPriv" , "AuthNoPriv" , "AuthPriv" ] ( no default set)
  # v3authuser 	:  Authentication user
  # v3authpass	:  Autentication password
  # v3authprot 	:  Autehtication Protocol =>  values  [ "MD5", "SHA" ] ( default none)
  #
  # -- only needed if AuthPriv mode enabled ---
  #
  # v3privpass 	: Privacy Password
  # v3privprot 	: Privacy Protocol => values [ "DES", "AES"] ( default none)

  # ---------------------------
  # RUNTIME OPTIONS
  # --------------------------
  #
  # freq          : Frequency of polling in seconds ( default none)
  # devicetagname : The tag name that will identify the device (default : "device")
  # devicetagvalue: the tag valude that will identify the device ( default : "id" ) [ posible values "id" --hostsnmpv2_1--  or "host" --127.0.0.1-- ]
  # extratags     : Array with extratags in format key = value will be sent to the influxdb database format [ "" , "" , ""]
  # snmpdebug     :
  # filters       : Array containing filter file per measurement.

*/


/*

  [snmpdevice.hostsnmpv3a]

   host = "hostsnmpv3a"
   port = 161
   timeout = 20
   retries = 5
   repeat = 0
   snmpversion = "3"
   v3seclevel = "NoAuthNoPriv"
   #only needed if Auth v3 enabled
   v3authuser = "v3usernoauth"
   freq = 60
   devicetagname = "router"
   devicetagvalue = "id"
   extratags = [ "tagA=4","tagB=5" ,"tagC=6" ]
   loglevel = "debug"
   snmpdebug  = false
   metricgroups = [ "Linux.*" ]
   measfilters = [
  		[ "linux_ports" , "file" , "lp_file_filter_a.txt" ,"EnableAlias"],
  	]

    */
INSERT INTO snmp_device_cfg (id,host,port,retries,timeout,repeat,snmpversion,v3seclevel,v3authuser,freq,devicetagname,devicetagvalue,'extra-tags',loglevel,snmpdebug) VALUES ('hostsnmpv3a','hostsnmpv3a',161,5,20,0,'3','NoAuthNoPriv','v3usernoauth',60,'router','id','[ "tagA=4","tagB=5","tagC=6" ]','debug',0);
INSERT INTO snmp_dev_m_groups (id_snmpdev,id_mgroup_cfg) VALUES ('hostsnmpv3a','Issue_index');
/*INSERT INTO snmp_dev_m_groups (id_snmpdev,id_mgroup_cfg) VALUES ('hostsnmpv3a','Linux.*');
INSERT INTO snmp_dev_filters (id_snmpdev,id_filter) VALUES ('hostsnmpv3a','filter_ports_file_a');*/


    /*

  [snmpdevice.hostsnmpv3b]
   #conection parameters
   host = "hostsnmpv3b"
   port = 161
   timeout = 20
   retries = 5
   repeat = 0
   #snmp version /auth params
   snmpversion = "3"
   v3seclevel = "AuthNoPriv"
   #only needed if Auth v3 enabled
   v3authuser = "v3userauth"
   v3authpass = "v3passauth"
   v3authprot = "MD5"
   # runtime parameters
   freq = 30
   devicetagname = "router"
   devicetagvalue = "id"
   #extratags = [ "tagA=4","tagB=5" ,"tagC=6" ]
   loglevel = "debug"
   snmpdebug  = false
   metricgroups = [ ".*ux" ]
   measfilters = [
  		[ "linux_ports" , "file", "lp_file_filter_b.txt","EnableAlias" ],
  	]
    */

    INSERT INTO snmp_device_cfg (id,host,port,retries,timeout,repeat,snmpversion,v3seclevel,v3authuser,v3authpass,v3authprot,freq,devicetagname,devicetagvalue,'extra-tags',loglevel,snmpdebug) VALUES  ('hostsnmpv3b','hostsnmpv3b',161,5,20,0,'3','AuthNoPriv','v3userauth','v3passauth','MD5',30,'router','id','[ "tagA=14","tagB=15","tagC=16" ]','debug',0);
    INSERT INTO snmp_dev_m_groups (id_snmpdev,id_mgroup_cfg) VALUES ('hostsnmpv3b','Linux');
    INSERT INTO snmp_dev_filters (id_snmpdev,id_filter) VALUES ('hostsnmpv3b','filter_ports_file_b');

    /*

  [snmpdevice.hostsnmpv3c]
   #conection parameters
   host = "hostsnmpv3c"
   port = 161
   timeout = 20
   retries = 5
   repeat = 0
   snmpversion = "3"
   v3seclevel = "AuthPriv"
   v3authuser = "v3userpriv"
   v3authpass = "v3privpass"
   v3authprot = "MD5"
   v3privpass = "v3passenc"
   v3privprot = "DES"
   freq = 30
   devicetagname = "router"
   devicetagvalue = "id"
   extratags = [ "tagA=4","tagB=5" ,"tagC=6" ]
   loglevel = "debug"
   snmpdebug  = false
   metricgroups = [ "Lin.*"]

*/

INSERT INTO snmp_device_cfg (id,host,port,retries,timeout,repeat,snmpversion,v3seclevel,v3authuser,v3authpass,v3authprot,freq,devicetagname,devicetagvalue,'extra-tags',loglevel,snmpdebug) VALUES ('hostsnmpv3c','hostsnmpv3c',161,5,20,0,'3','AuthNoPriv','v3userauth','v3passauth','MD5',30,'router','id','[ "tagA=14","tagB=15","tagC=16"]','debug',0);
INSERT INTO snmp_dev_m_groups (id_snmpdev,id_mgroup_cfg) VALUES ('hostsnmpv3c','Linux');
/*INSERT INTO snmp_dev_filters (id_snmpdev,id_filter) VALUES ('hostsnmpv3c','filter_port_if_status_up');*/
INSERT INTO snmp_dev_filters (id_snmpdev,id_filter) VALUES ('hostsnmpv3c','filter_port_if_name_match_eth');
