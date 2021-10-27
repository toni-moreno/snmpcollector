export const RuntimeComponentConfig: any =
  {
    'name' : 'Runtime',
    'table-columns' : [
      { title: 'ID', name: 'ID' },
      { title: 'TagMap', name: 'TagMap', tooltip: 'Num Measurements configured' },
      { title: 'SysDesc', name: 'SysDescription' },
      { title: '#Meas', name: 'NumMeasurements', tooltip: 'Num Measurements configured' },
      { title: '#Metrics', name: 'NumMetrics' },
      { title: 'Get.Errs', name: 'Counter7', tooltip: 'SnmpOIDGetErrors:number of  oid with errors for all measurements ' },
      { title: 'M.Errs', name: 'Counter14', tooltip: 'MeasurementSentErrors: number of measuremenets  formatted with errors ' },
      { title: 'G.Time', name: 'Counter16', tooltip: 'CycleGatherDuration time: elapsed time taken to get all measurement info', transform: 'elapsedseconds' },
      { title: 'F.Time', name: 'Counter18', tooltip: 'CycleGatherDuration time: elapsed time taken to compute all applicable filters on the device', transform: 'elapsedseconds' }
    ],
  }; 

export const TableRole : string = 'runtime';

export const ExtraActions: Object = {
  data: [
    { 
      title: 'SetActive', content: [
        { 
          type: 'boolean',
          enabled: '<i class="glyphicon glyphicon-pause"></i>',
          disabled: '<i class="glyphicon glyphicon-play"></i>',
          property: 'Active',
          action: "SetActive"
        }
      ] 
    },
    { 
      title: 'SnmpReset', content: [
        {
          type: 'button', 
          text: '<span>Reset</span>', 
          action: 'SnmpReset'
        }
      ] 
    }
  ],
  "position": "last"
};

export const DeviceCounterDef: CounterType[] = [
  /*0*/    { show: false, id: "SnmpGetQueries", label: "SnmpGet Queries", type: "counter", tooltip: "number of snmp queries" },
  /*1*/    { show: false, id: "SnmpWalkQueries", label: "SnmpWalk Queries", type: "counter", tooltip: "number of snmp walks" },
  /*2*/    { show: false, id: "SnmpGetErrors", label: "SnmpGet Errors", type: "counter", tooltip: "Get Error" },
  /*3*/    { show: false, id: "SnmpWalkErrors", label: "SnmpWalk Errors", type: "counter", tooltip: "Walk Error" },
  /*4*/    { show: false, id: "SnmpQueryTimeouts", label: "Snmp Errors by Timeout", type: "counter", tooltip: "number of registered errors by timeouts" },
  /*5*/    { show: true, id: "SnmpOIDGetAll", label: "OID Gets ALL", type: "counter", tooltip: "All Gathered snmp metrics ( sum of snmpget oid's and all received oid's in snmpwalk queries)" },
  /*6*/    { show: true, id: "SnmpOIDGetProcessed", label: "OID Processed", type: "counter", tooltip: "Gathered and processed snmp metrics after filters are applied ( not always sent to the backend it depens on the report flag)" },
  /*7*/    { show: true, id: "SnmpOIDGetErrors", label: "OID With Errors", type: "counter", tooltip: "number of  oid with errors for all measurements" },
  /*8*/    { show: false, id: "EvalMetricsAll", label: "Evaluated Metrics", type: "counter", tooltip: "number of evaluated metrics" },
  /*9*/    { show: false, id: "EvalMetricsOk", label: "Evaluated Metrics", type: "counter", tooltip: "number of evaluated metrics without errors." },
  /*10*/    { show: false, id: "EvalMetricsErrors", label: "Evaluated Metrics", type: "counter", tooltip: "number of evaluated metrics with some kind of error." },
  /*11*/    { show: true, id: "MetricSent", label: "Metric Sent", type: "counter", tooltip: "number of metrics sent (taken as fields) for all measurements" },
  /*12*/    { show: true, id: "MetricSentErrors", label: "Metric Sent Errors", type: "counter", tooltip: "number of metrics  (taken as fields) with errors forall measurements" },
  /*13*/    { show: true, id: "MeasurementSent", label: "Series sent", type: "counter", tooltip: "(number of  measurements build to send as a sigle request sent to the backend)" },
  /*14*/    { show: true, id: "MeasurementSentErrors", label: "Series sent Errors", type: "counter", tooltip: "(number of measuremenets  formatted with errors )" },
  /*15*/    { show: false, id: "CycleGatherStartTime", label: "Cycle Gather Start Time", type: "time", tooltip: "Last gather time " },
  /*16*/    { show: true, id: "CycleGatherDuration", label: "Cycle Gather Duration", type: "duration", tooltip: "elapsed time taken to get all measurement info" },
  /*17*/    { show: false, id: "FilterStartTime", label: "Filter update Start Time", type: "time", tooltip: "Last Filter time" },
  /*18*/    { show: true, id: "FilterDuration", label: "Filter update Duration", type: "duration", tooltip: "elapsed time taken to compute all applicable filters on the device" },
  /*19*/    { show: false, id: "BackEndSentStartTime", label: "BackEnd (influxdb) Sent Start Time", type: "time", tooltip: "Last sent time" },
  /*20*/    { show: false, id: "BackEndSentDuration", label: "BackEnd (influxdb) Sent Duration", type: "duration", tooltip: "elapsed time taken to send data to the db backend" },
];

export const MeasurementCounterDef: CounterType[] = [
  /*0*/    { show: false, id: "SnmpGetQueries", label: "SnmpGet Queries", type: "counter", tooltip: "number of snmp queries" },
  /*1*/    { show: false, id: "SnmpWalkQueries", label: "SnmpWalk Queries", type: "counter", tooltip: "number of snmp walks" },
  /*2*/    { show: false, id: "SnmpGetErrors", label: "SnmpGet Errors", type: "counter", tooltip: "Get Error" },
  /*3*/    { show: false, id: "SnmpWalkErrors", label: "SnmpWalk Errors", type: "counter", tooltip: "Walk Error" },
  /*4*/    { show: false, id: "SnmpQueryTimeouts", label: "Snmp Errors by Timeout", type: "counter", tooltip: "number of registered errors by timeouts" },
  /*5*/    { show: true, id: "SnmpOIDGetAll", label: "OID Gets ALL", type: "counter", tooltip: "All Gathered snmp metrics ( sum of snmpget oid's and all received oid's in snmpwalk queries)" },
  /*6*/    { show: true, id: "SnmpOIDGetProcessed", label: "OID Processed", type: "counter", tooltip: "Gathered and processed snmp metrics after filters are applied ( not always sent to the backend it depens on the report flag)" },
  /*7*/    { show: true, id: "SnmpOIDGetErrors", label: "OID With Errors", type: "counter", tooltip: "number of  oid with errors for all measurements" },
  /*8*/    { show: false, id: "EvalMetricsAll", label: "Evaluated Metrics", type: "counter", tooltip: "number of evaluated metrics" },
  /*9*/    { show: false, id: "EvalMetricsOk", label: "Evaluated Metrics", type: "counter", tooltip: "number of evaluated metrics without errors." },
  /*10*/    { show: false, id: "EvalMetricsErrors", label: "Evaluated Metrics", type: "counter", tooltip: "number of evaluated metrics with some kind of error." },
  /*11*/    { show: true, id: "MetricSent", label: "Metric Sent", type: "counter", tooltip: "number of metrics sent (taken as fields) for all measurements" },
  /*12*/    { show: true, id: "MetricSentErrors", label: "Metric Sent Errors", type: "counter", tooltip: "number of metrics  (taken as fields) with errors forall measurements" },
  /*13*/    { show: true, id: "MeasurementSent", label: "Series sent", type: "counter", tooltip: "(number of  measurements build to send as a sigle request sent to the backend)" },
  /*14*/    { show: true, id: "MeasurementSentErrors", label: "Series sent Errors", type: "counter", tooltip: "(number of measuremenets  formatted with errors )" },
  /*15*/    { show: false, id: "CycleGatherStartTime", label: "Cycle Gather Start Time", type: "time", tooltip: "Last gather time " },
  /*16*/    { show: true, id: "CycleGatherDuration", label: "Cycle Gather Duration", type: "duration", tooltip: "elapsed time taken to get all measurement info" },
  /*17*/    { show: false, id: "FilterStartTime", label: "Filter update Start Time", type: "time", tooltip: "Last Filter time" },
  /*18*/    { show: true, id: "FilterDuration", label: "Filter update Duration", type: "duration", tooltip: "elapsed time taken to compute all applicable filters on the device" },
  /*19*/    { show: false, id: "BackEndSentStartTime", label: "BackEnd (influxdb) Sent Start Time", type: "time", tooltip: "Last sent time" },
  /*20*/    { show: true, id: "BackEndSentDuration", label: "BackEnd (influxdb) Sent Duration", type: "duration", tooltip: "elapsed time taken to send data to the db backend" },
];