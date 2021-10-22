export const RuntimeComponentConfig: any =
  {
    'name' : 'Runtime Output',
    'table-columns' : [
      { title: 'Output ID', name: 'ID' },
      { title: 'FieldSent', name: 'FieldSent', tooltip: 'Num of fields sent' },
      { title: 'FieldSentMax', name: 'FieldSentMax', tooltip: 'Max num of fields sent' },
      { title: 'BufferPercentUsed (%)', name: 'BufferPercentUsed', tooltip: 'Buffer % used', transform: 'percent' },
      { title: 'PSent', name: 'PSent', tooltip: 'Num of points sent' },
      { title: 'PSentMax', name: 'PSentMax', tooltip: 'Max num of points sent' },
      { title: 'WriteErrors', name: 'WriteErrors', tooltip: 'Num of writes errors' },
      { title: 'WriteSent', name: 'WriteSent', tooltip: 'Num of writes sent' },
      { title: 'WriteTime', name: 'WriteTime', tooltip: 'Write time', transform: 'elapsednanoseconds' },
      { title: 'WriteTimeMax', name: 'WriteTimeMax', tooltip: 'Max write time', transform: 'elapsednanoseconds' }
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
      title: 'Set Enqueue On Error', content: [
        { 
          type: 'boolean',
          enabled: '<span>Not enqueue</span>',
          disabled: '<span>Enqueue</span>',
          property: 'EnqueueOnWriteError',
          action: "SetEnqueuePolicy"
        }
      ] 
    },
    { 
      title: 'Flush Buffer', content: [
        {
          type: 'button', 
          text: '<span>Flush</span>',
          tooltip: 'Flushes current points on buffer to Output',
          action: 'FlushBuffer'
        }
      ] 
    },
    { 
      title: 'Reset Buffer', content: [
        {
          type: 'button',
          color: 'warning',
          text: '<span>Reset</span>',
          tooltip: 'Removes current points on buffer',
          action: 'ResetBuffer'
        }
      ] 
    }
  ],
  "position": "last"
};
