export const OidConditionCfgComponentConfig: any =
  {
    'name' : 'OID Conditions',
    'table-columns' : [
      { title: 'ID', name: 'ID' },
      { title: 'Is Multiple', name: 'IsMultiple' },
      { title: 'OIDCond', name: 'OIDCond' },
      { title: 'CondType', name: 'CondType' },
      { title: 'CondValue', name: 'CondValue' }
    ],
    'slug' : 'oidconditioncfg'
  }; 

  export const TableRole : string = 'fulledit';
  export const OverrideRoleActions : Array<Object> = [
    {'name':'export', 'type':'icon', 'icon' : 'glyphicon glyphicon-download-alt text-default', 'tooltip': 'Export item'},
    {'name':'view', 'type':'icon', 'icon' : 'glyphicon glyphicon-eye-open text-success', 'tooltip': 'View item'},
    {'name':'edit', 'type':'icon', 'icon' : 'glyphicon glyphicon-edit text-warning', 'tooltip': 'Edit item'},
    {'name':'remove', 'type':'icon', 'icon' : 'glyphicon glyphicon glyphicon-remove text-danger', 'tooltip': 'Remove item'}
  ]