export const SnmpDeviceCfgComponentConfig: any =
  {
    'name' : 'SNMP Devices',
    'table-columns' : [
      { title: 'ID', name: 'ID' },
      { title: 'Host', name: 'Host' },
      { title: 'Port', name: 'Port' },
      { title: 'Active', name: 'Active' },
      { title: 'Alternate System OIDs', name: 'SystemOIDs' },
      { title: 'Snmp Version', name: 'SnmpVersion' },
      { title: 'Snmp Debug', name: 'SnmpDebug' },
      { title: 'Polling Period (sec)', name: 'Freq' },
      { title: 'Update Filter (Cycles)', name: 'UpdateFltFreq' },
      { title: 'Concurrent Gather', name: 'ConcurrentGather' },
      { title: 'Influx DB', name: 'OutDB' },
      { title: 'Log Level', name: 'LogLevel' },
      { title: 'Disable Snmp Bulk Queries', name: 'DisableBulk' },
      { title: 'Timeout', name: 'Timeout' },
      { title: 'Retries', name: 'Retries' },
      { title: 'SNMP Max Repetitions', name: 'MaxRepetitions' },
      { title: 'Tag Name', name: 'DeviceTagName' },
      { title: 'Tag Value', name: 'DeviceTagValue' },
      { title: 'Extra Tags', name: 'ExtraTags' },
      { title: 'Device Variables', name: 'DeviceVars' },
      { title: 'Measurement Groups', name: 'MeasurementGroups' },
      { title: 'Measurement Filters', name: 'MeasFilters' }
    ],
    'slug' : 'devicecfg'
  }; 

  export const TableRole : string = 'fulledit';
  export const OverrideRoleActions : Array<Object> = [
    {'name':'export', 'type':'icon', 'icon' : 'glyphicon glyphicon-download-alt text-default', 'tooltip': 'Export item'},
    {'name':'test-connection', 'type':'icon', 'icon' : 'glyphicon glyphicon-flash text-info', 'tooltip': 'Test connection'},
    {'name':'view', 'types':'icon', 'icon' : 'glyphicon glyphicon-eye-open text-success', 'tooltip': 'View item'},
    {'name':'edit', 'type':'icon', 'icon' : 'glyphicon glyphicon-edit text-warning', 'tooltip': 'Edit item'},
    {'name':'remove', 'type':'icon', 'icon' : 'glyphicon glyphicon glyphicon-remove text-danger', 'tooltip': 'Remove item'}
  ]