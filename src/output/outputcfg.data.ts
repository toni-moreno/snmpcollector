export const OutputCfgComponentConfig: any =
  {
    'name' : 'Outputs',
    'table-columns' : [
      { title: 'ID', name: 'ID' },
      { title: 'Active', name: 'Active' },
      { title: 'Enq. On Error', name: 'EnqueueOnError' },
      { title: 'BackendType', name: 'BackendType' },
      { title: 'BufferSize', name: 'BufferSize' },
      { title: 'MetricBatchSize', name: 'MetricBatchSize' },
      { title: 'FlushInterval', name: 'FlushInterval' },
      { title: 'Backend', name: 'Backend'}
    ],
    'slug' : 'outputcfg'
  }; 

  export const TableRole : string = 'fulledit';
  export const OverrideRoleActions : Array<Object> = [
    {'name':'export', 'type':'icon', 'icon' : 'glyphicon glyphicon-download-alt text-default', 'tooltip': 'Export item'},
    {'name':'view', 'type':'icon', 'icon' : 'glyphicon glyphicon-eye-open text-success', 'tooltip': 'View item'},
    {'name':'edit', 'type':'icon', 'icon' : 'glyphicon glyphicon-edit text-warning', 'tooltip': 'Edit item'},
    {'name':'remove', 'type':'icon', 'icon' : 'glyphicon glyphicon glyphicon-remove text-danger', 'tooltip': 'Remove item'}
  ]