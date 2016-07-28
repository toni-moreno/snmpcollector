/*
[metrics]
[metrics."Linux_user_CPU_percent"]
  fieldname = "user"
  description = "percentage of user CPU time"
  baseOID = ".1.3.6.1.4.1.2021.11.9.0"
  Datasrctype = "INTEGER"
*/

INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift,description)  VALUES ('Linux_user_CPU_percent','user','.1.3.6.1.4.1.2021.11.9.0','INTEGER',0,0.0,0.0,'percentage of user CPU time');
/*
[metrics."Linux_system_CPU_percent"]
  fieldname = "system"
  description = "percentage of System CPU time"
  baseOID = ".1.3.6.1.4.1.2021.11.10.0"
  Datasrctype = "INTEGER"
*/
INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift,description)  VALUES ('Linux_system_CPU_percent','system','.1.3.6.1.4.1.2021.11.10.0','INTEGER',0,0.0,0.0,'percentage of system CPU time');

/*
[metrics."Linux_idle_CPU_percent"]
  fieldname = "idle"
  description = "percentage of System CPU time"
  baseOID = ".1.3.6.1.4.1.2021.11.11.0"
  Datasrctype = "INTEGER"
  */

INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift,description)  VALUES ('Linux_idle_CPU_percent','idle','.1.3.6.1.4.1.2021.11.11.0','INTEGER',0,0.0,0.0,'percentage of Idle CPU time');

/*
[metrics."ifHCInOctets"]
  fieldname = "In.bytes"
  description = "Bytes In - 64-bit Counters"
  baseOID = ".1.3.6.1.2.1.31.1.1.1.6"
  DatasrcType = "COUNTER64"
  getrate = false
*/

INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift,description)  VALUES ('ifHCInOctets','In.bytes','.1.3.6.1.2.1.31.1.1.1.6','COUNTER64',0,0.0,0.0,'Bytes In - 64-bit Counters');

/*

[metrics."ifHCOutOctets"]
  fieldname = "Out.bytes"
  description = "Bytes In - 64-bit Counters"
  baseOID = ".1.3.6.1.2.1.31.1.1.1.10"
  DatasrcType = "COUNTER64"
  getRate = false
*/
INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift,description)  VALUES ('ifHCOutOctets','Out.bytes','.1.3.6.1.2.1.31.1.1.1.10','COUNTER64',0,0.0,0.0,'Bytes Out - 64-bit Counters');

/*
[metrics."ifHCInUcastPkts"]
  fieldname = "In.packets"
  description = "Bytes In - 64-bit Counters"
  baseOID = ".1.3.6.1.2.1.31.1.1.1.7"
  DatasrcType = "COUNTER64"
  getRate = false
*/
INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift,description)  VALUES ('ifHCInUcastPkts','In.packets','.1.3.6.1.2.1.31.1.1.1.7','COUNTER64',0,0.0,0.0,'Packets In - 64-bit Counters');
/*
[metrics."ifHCOutUcastPkts"]
  fieldname = "Out.packets"
  description = "Packets - 64-bit Counters"
  baseOID = ".1.3.6.1.2.1.31.1.1.1.11"
  DatasrcType = "COUNTER64"
  getRate = false
*/
INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift,description)  VALUES ('ifHCOutUcastPkts','Out.packets','.1.3.6.1.2.1.31.1.1.1.11','COUNTER64',0,0.0,0.0,'Packets Out - 64-bit Counters');
/*
[metrics."ifInOctets"]
  fieldname = "In.bytes"
  description = "Bytes In - 32-bit Counters"
  baseOID = ".1.3.6.1.2.1.2.2.1.10"
  DatasrcType = "COUNTER32"
  getRate = false

  */

INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift,description)  VALUES ('ifInOctets','In.bytes','.1.3.6.1.2.1.2.2.1.10','COUNTER32',0,0.0,0.0,'Bytes In - 32-bit Counters');

/*
[metrics."ifOutOctets"]
  fieldname = "Out.bytes"
  description = "Bytes Out - 32-bit Counters"
  baseOID = ".1.3.6.1.2.1.2.2.1.16"
  DatasrcType = "COUNTER32"
  getRate = false

*/

INSERT INTO snmp_metric_cfg  (id, field_name, baseoid, datasrctype, getrate, scale, shift,description)  VALUES ('ifOutOctets','Out.bytes','.1.3.6.1.2.1.2.2.1.16','COUNTER32',0,0.0,0.0,'Bytes Out - 32-bit Counters');



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
INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('network_32bits','ifOctets');
INSERT INTO measurement_field_cfg( id_measurement_cfg,id_metric_cfg) VALUES ('network_32bits','ifOutOctets');
