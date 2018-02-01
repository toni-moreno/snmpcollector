export const InfluxMeasCfgComponentConfig: any =
  {
    'name' : 'Influx Measurements',
    'table-columns' : [
      { title: 'ID', name: 'ID' },
      { title: 'Name', name: 'Name' },
      { title: 'GetMode', name: 'GetMode' },
      { title: 'Index OID', name: 'IndexOID' },
      { title: 'Tag OID', name: 'TagOID' },
      { title: 'Index Tag', name: 'IndexTag' },
      { title: 'Index Tag Format', name: 'IndexTagFormat' },
      { title: 'Index as Value', name: 'IndexAsValue' },
      { title: 'Metric Fields', name: 'Fields'}
    ],
    'slug' : 'measurementcfg'
  }; 

  export const TableRole : string = 'fulledit';
  export const OverrideRoleActions : Array<Object> = [
    {'name':'export', 'type':'icon', 'icon' : 'glyphicon glyphicon-download-alt text-default', 'tooltip': 'Export item'},
    {'name':'view', 'type':'icon', 'icon' : 'glyphicon glyphicon-eye-open text-success', 'tooltip': 'View item'},
    {'name':'edit', 'type':'icon', 'icon' : 'glyphicon glyphicon-edit text-warning', 'tooltip': 'Edit item'},
    {'name':'remove', 'type':'icon', 'icon' : 'glyphicon glyphicon glyphicon-remove text-danger', 'tooltip': 'Remove item'}
  ]