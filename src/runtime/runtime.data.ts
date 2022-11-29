export const RuntimeComponentConfig: any =
  {
    'name' : 'Runtime',
    'table-columns' : [
      { title: 'ID', name: 'ID' },
      { title: 'TagMap', name: 'TagMap', tooltip: 'List of device tags' },
      { title: 'SysDesc', name: 'SysDescr', tooltip: 'Response of SysDescription OID query' },
      { title: 'Next Statistic Update', name: 'GatherNextTime', tooltip: 'Expected time where statistics will be updated', transform: 'toDate' },
      { title: 'Statistics update Freq', name: 'GatherFreq', tooltip: ' Frequency of statistics gathering from measurements in seconds', },
      { title: '#Meas', name: 'NumMeasurements', tooltip: 'Number of measurements collected on the device in the last statistics period ' },
      { title: '#Metrics', name: 'NumMetrics', tooltip: 'Number of metrics created on device' },
      { title: 'Get.Errs', name: 'Counter7', tooltip: 'SnmpOIDGetErrors: number of  oid with errors for all measurements on last statistics period' },
      { title: 'M.Errs', name: 'Counter14', tooltip: 'MeasurementSentErrors: number of measurements formatted with errors on last statistics period' },
      { title: 'G.Time', name: 'Counter16', tooltip: 'CycleGatherDuration time: elapsed time taken to get all measurement info on last statistics period', transform: 'elapsedseconds' },
      { title: 'F.Time', name: 'Counter18', tooltip: 'CycleGatherDuration time: elapsed time taken to compute all applicable filters on the device on last statistics period', transform: 'elapsedseconds' }
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
  { show: false, source: "counters", id: "SnmpGetQueries", idx:0,  label: "SnmpGet Queries", type: "counter", tooltip: "Number of snmp queries" },
  { show: false, source: "counters", id: "SnmpWalkQueries", idx:1, label: "SnmpWalk Queries", type: "counter", tooltip: "Number of snmp walks" },
  { show: false, source: "counters", id: "SnmpGetErrors", idx:2, label: "SnmpGet Errors", type: "counter", tooltip: "Number of walk errors" },
  { show: false, source: "counters", id: "SnmpWalkErrors", idx:3, label: "SnmpWalk Errors", type: "counter", tooltip: "Walk Error" },
  { show: false, source: "counters", id: "SnmpQueryTimeouts", idx:4, label: "Snmp Errors by Timeout", type: "counter", tooltip: "Number of registered errors by timeouts" },
  { show: true, source: "counters", id: "SnmpOIDGetAll", idx:5, label: "OID Gets ALL", type: "counter", tooltip: "All Gathered snmp metrics (sum of SNMPGET OID's and all received OID's in SNMPWALK queries)" },
  { show: true, source: "counters", id: "SnmpOIDGetProcessed", idx:6, label: "OID Processed", type: "counter", tooltip: "Gathered and processed snmp metrics after filters are applied ( not always sent to the backend it depens on the report flag)" },
  { show: true, source: "counters", id: "SnmpOIDGetErrors", idx:7, label: "OID With Errors", type: "counter", tooltip: "Number of OIDs with errors for all measurements" },
  { show: false, source: "counters", id: "EvalMetricsAll", idx:8, label: "Evaluated Metrics", type: "counter", tooltip: "Number of evaluated metrics" },
  { show: false, source: "counters", id: "EvalMetricsOk", idx:9, label: "Evaluated Metrics", type: "counter", tooltip: "Number of evaluated metrics without errors." },
  { show: false, source: "counters", id: "EvalMetricsErrors", idx:10,  label: "Evaluated Metrics", type: "counter", tooltip: "Number of evaluated metrics with some kind of error." },
  { show: true, source: "counters", id: "MetricSent", idx:11, label: "Metric Sent", type: "counter", tooltip: "Number of metrics sent (taken as fields) for all measurements" },
  { show: true, source: "counters", id: "MetricSentErrors", idx:12, label: "Metric Sent Errors", type: "counter", tooltip: "Number of metrics (taken as fields) with errors for all measurements" },
  { show: true, source: "counters", id: "MeasurementSent", idx:13, label: "Series sent", type: "counter", tooltip: "Number of series build to send as a sigle request sent to the backend" },
  { show: true, source: "counters", id: "MeasurementSentErrors", idx:14, label: "Series sent Errors", type: "counter", tooltip: "Number of series formatted with errors" },
  { show: false, source: "counters", id: "CycleGatherStartTime", idx:15, label: "Cycle Gather Start Time", type: "time", tooltip: "Last gather time " },
  { show: true, source: "counters", id: "CycleGatherDuration", idx:16, label: "Cycle Gather Duration", type: "duration", tooltip: "Elapsed time taken to get all measurement info" },
  { show: false, source: "counters", id: "FilterStartTime", idx:17, label: "Filter update Start Time", type: "time", tooltip: "Last Filter time" },
  { show: true, source: "counters", id: "FilterDuration", idx:18, label: "Filter update Duration", type: "duration", tooltip: "Elapsed time taken to compute all applicable filters on the device" },
  { show: false, source: "counters", id: "BackEndSentStartTime", idx:19, label: "BackEnd DB Sent Start Time", type: "time", tooltip: "Last sent time" },
  { show: false, source: "counters", id: "BackEndSentDuration", idx:20, label: "BackEnd DB Sent Duration", type: "duration", tooltip: "Elapsed time taken to send data to the db backend" },
];

export const MeasurementCounterDef: CounterType[] = [
  { show: false, source: "counters", id: "SnmpGetQueries", idx: 0, label: "SnmpGet Queries", type: "counter", tooltip: "Number of snmp queries" },
  { show: false, source: "counters", id: "SnmpWalkQueries", idx: 1, label: "SnmpWalk Queries", type: "counter", tooltip: "Number of snmp walks" },
  { show: false, source: "counters", id: "SnmpGetErrors", idx: 2, label: "SnmpGet Errors", type: "counter", tooltip: "Number of get errors" },
  { show: false, source: "counters", id: "SnmpWalkErrors", idx: 3, label: "SnmpWalk Errors", type: "counter", tooltip: "Number of walk errors" },
  { show: false, source: "counters", id: "SnmpQueryTimeouts", idx: 4, label: "Snmp Errors by Timeout", type: "counter", tooltip: "Number of registered errors by timeouts" },
  { show: true, source: "counters", id: "SnmpOIDGetAll", idx: 5, label: "OID Gets ALL", type: "counter", tooltip: "All Gathered snmp metrics (sum of snmpget oid's and all received oid's in snmpwalk queries)" },
  { show: true, source: "counters", id: "SnmpOIDGetProcessed", idx: 6, label: "OID Processed", type: "counter", tooltip: "Gathered and processed snmp metrics after filters are applied ( not always sent to the backend it depens on the report flag)" },
  { show: true, source: "counters", id: "SnmpOIDGetErrors", idx: 7, label: "OID With Errors", type: "counter", tooltip: "Number of OID with errors for all measurements" },
  { show: false, source: "counters", id: "EvalMetricsAll", idx: 8, label: "Evaluated Metrics", type: "counter", tooltip: "Number of evaluated metrics" },
  { show: false, source: "counters", id: "EvalMetricsOk", idx: 9, label: "Evaluated Metrics", type: "counter", tooltip: "Number of evaluated metrics without errors." },
  { show: false, source: "counters", id: "EvalMetricsErrors", idx: 10, label: "Evaluated Metrics", type: "counter", tooltip: "Number of evaluated metrics with some kind of error." },
  { show: true, source: "counters", id: "MetricSent", idx: 11, label: "Metric Sent", type: "counter", tooltip: "Number of metrics sent (taken as fields) for all measurements" },
  { show: true, source: "counters", id: "MetricSentErrors", idx: 12, label: "Metric Sent Errors", type: "counter", tooltip: "Number of metrics  (taken as fields) with errors forall measurements" },
  { show: true, source: "counters", id: "MeasurementSent", idx: 13, label: "Series sent", type: "counter", tooltip: "Number of series build to send as a sigle request sent to the backend" },
  { show: true, source: "counters", id: "MeasurementSentErrors", idx: 14, label: "Series sent Errors", type: "counter", tooltip: "Number of series formatted with errors" },
  { show: false, source: "counters", id: "FilterStartTime", idx: 17, label: "Filter update Start Time", type: "time", tooltip: "Last Filter time" },
  { show: true, source: "counters", id: "FilterDuration", idx: 18, label: "Filter update Duration", type: "duration", tooltip: "Elapsed time taken to compute all applicable filters on the device" },
  { show: false, source: "counters", id: "BackEndSentStartTime", idx: 19, label: "BackEnd DB Sent Start Time", type: "time", tooltip: "Last sent time" },
  { show: true, source: "counters", id: "MeasurementDropped", idx:23, label: "Series dropped", type: "counter", tooltip: "Series dropped due to full buffer" },
  { show: false, source: "counters", id: "BackEndSentDuration", idx:20, label: "BackEnd DB Sent Duration", type: "duration", tooltip: "Elapsed time taken to send data to the db backend" },
  { show: true, source: "stats", id: "GatherFreq", label: "Gather Frequency", type: "duration", tooltip: "Gather frequency" },
  { show: true, source: "counters", id: "CycleGatherStartTime", idx: 15, label: "Cycle Gather Start Time", type: "time", tooltip: "Last gather time" },
  { show: true, source: "counters", id: "CycleGatherDuration", idx: 16, label: "Cycle Gather Duration", type: "duration", tooltip: "Elapsed time taken to get all measurement info" },
  { show: true, source: "stats", id: "GatherNextTime", label: "Gather Next Time", type: "time", tooltip: "Next planned gather time" },
  { show: true, source: "stats", id: "FilterNextTime", label: "Filter Next Time", type: "time", tooltip: "Next planned filter update time" },
];