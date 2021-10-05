export const PollerLocationCfgComponentConfig: any =
  {
    'name' : 'Locations',
    'table-columns' : [
      { title: 'Instance ID', name: 'Instance_ID' },
      { title: 'Location', name: 'Location' },
      { title: 'Active', name: 'Active' },
      { title: 'Hostname', name: 'Hostname' },
      { title: 'IP', name: 'IP' },
      { title: 'Description', name: 'Description' }
    ],
    'slug' : 'pollerlocationcfg'
  }; 

  export const TableRole : string = 'fulledit';
  export const OverrideRoleActions : Array<Object> = [
    {'name':'export', 'type':'icon', 'icon' : 'glyphicon glyphicon-download-alt text-default', 'tooltip': 'Export item'},
    {'name':'view', 'type':'icon', 'icon' : 'glyphicon glyphicon-eye-open text-success', 'tooltip': 'View item'},
    {'name':'edit', 'type':'icon', 'icon' : 'glyphicon glyphicon-edit text-warning', 'tooltip': 'Edit item'},
    {'name':'remove', 'type':'icon', 'icon' : 'glyphicon glyphicon glyphicon-remove text-danger', 'tooltip': 'Remove item'}
  ]