export const CustomFilterCfgComponentConfig: any =
  {
    'name' : 'Custom Filters',
    'table-columns' : [
      { title: 'ID', name: 'ID' },
      { title: 'RelatedDev', name: 'RelatedDev' },
      { title: 'RelatedMeas', name: 'RelatedMeas' },
      { title: 'Items', name: 'Items' },
    ],
    'slug' : 'customfiltercfg'
  }; 

  export const TableRole : string = 'fulledit';
  export const OverrideRoleActions : Array<Object> = [
    {'name':'export', 'type':'icon', 'icon' : 'glyphicon glyphicon-download-alt text-default', 'tooltip': 'Export item'},
    {'name':'view', 'type':'icon', 'icon' : 'glyphicon glyphicon-eye-open text-success', 'tooltip': 'View item'},
    {'name':'edit', 'type':'icon', 'icon' : 'glyphicon glyphicon-edit text-warning', 'tooltip': 'Edit item'},
    {'name':'remove', 'type':'icon', 'icon' : 'glyphicon glyphicon glyphicon-remove text-danger', 'tooltip': 'Remove item'}
  ]