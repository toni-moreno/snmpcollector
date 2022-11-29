import { FormBuilder, Validators, FormArray, FormGroup, FormControl} from '@angular/forms';
import { ValidationService } from './validation.service'

export class AvailableTableActions {

  //AvailableOptions result depeding on component type
  public availableOptions : Array<any>;

  // type can be : device,...
  // data is the passed extraData when declaring AvailableTableActions on each component
  checkComponentType(type, data?) : any {
    switch (type) {
      case 'snmpdevicecfg':
        return this.getDeviceAvailableActions(data);
      case 'snmpmetriccfg':
        return this.getMetricAvailableActions();
      case 'influxcfg':
        return this.getInfluxServersAvailableActions();
      case 'kafkacfg':
        return this.getKafkaServersAvailableActions();
      case 'outputcfg':
          return this.getOutputAvailableActions();
      case 'oidconditioncfg':
        return this.getOIDConditionsAvailableActions();
      case 'measgroupcfg':
          return this.getMeasGroupsAvailableActions();
      case 'measfiltercfg':
          return this.getMeasFiltersAvailableActions();
      case 'customfiltercfg':
          return this.getCustomFiltersAvailableActions();
      case 'measurementcfg':
          return this.getMeasurementsAvailableActions();
      case 'varcatalogcfg':
          return this.getVarCatalogAvailableActions();
      default:
        return null;
      }
  }

  constructor (componentType : string, extraData? : any) {
    this.availableOptions = this.checkComponentType(componentType, extraData);
  }

  //Devices Available Acions:
  getDeviceAvailableActions (data) : any {
    let tableAvailableActions = [
    //Remove Action
      {'title': 'Remove', 'content' :
        {'type' : 'button','action' : 'RemoveAllSelected'}
      },
    //Change Property Action
      {'title': 'Change property', 'content' :
        {'type' : 'selector', 'action' : 'ChangeProperty', 'options' : [
          {'title' : 'Active', 'type':'boolean', 'options' : [
            'true','false'
            ]
          },
          {'title' : 'LogLevel', 'type':'boolean', 'options' : [
            'panic','fatal','error','warning','info','debug'
            ]
          },
          {'title' : 'ConcurrentGather', 'type':'boolean', 'options' : [
            'true','false'
            ]
          },
          {'title' : 'DisableBulk', 'type':'boolean', 'options' : [
            'true','false'
            ]
          },
          {'title' : 'SnmpDebug', 'type':'boolean', 'options' : [
            'true','false'
            ]
          },
          {'title': 'Timeout','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.compose([Validators.required,ValidationService.uintegerNotZeroValidator]))
            })
          },
          {'title': 'Retries','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.compose([Validators.required,ValidationService.uintegerNotZeroValidator]))
            })
          },
          {'title': 'Freq','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.compose([Validators.required,ValidationService.uintegerNotZeroValidator]))
            })
          },
          {'title': 'MaxRepetitions','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.compose([Validators.required,ValidationService.uinteger8NotZeroValidator]))
            })
          },
          {'title': 'UpdateFltFreq','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.compose([Validators.required, ValidationService.uintegerAndLessOneValidator]))
            })
          },
          {'title': 'DeviceTagName','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.required)
            })
          },
          {'title' : 'MeasurementGroups', 'type':'multiselector', 'options' :
            data[0]
          },
          {'title' : 'MeasFilters', 'type':'multiselector', 'options' :
            data[1]
          },
          {'title' : 'ExtraTags', 'type':'input', 'options' :
            new FormGroup({
              formControl : new FormControl('', Validators.compose([ValidationService.noWhiteSpaces, ValidationService.extraTags]))
            })
          }
        ]},
      },
      //AppendProperty
      {'title': 'AppendProperty', 'content' :
        {'type' : 'selector', 'action' : 'AppendProperty', 'options' : [
          {'title' : 'MeasurementGroups', 'type':'multiselector', 'options' :
            data[0]
          },
          {'title' : 'MeasFilters', 'type':'multiselector', 'options' :
            data[1]
          },
          {'title' : 'ExtraTags', 'type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.compose([ValidationService.noWhiteSpaces, ValidationService.extraTags]))
            })
          }
          ]
        }
      }
    ];
    return tableAvailableActions;
  }

  //Metric Available Acions:
  getMetricAvailableActions (data ? : any) : any {
    let tableAvailableActions = [
    //Remove Action
      {'title': 'Remove', 'content' :
        {'type' : 'button','action' : 'RemoveAllSelected'}
      },
    //Change Property Action
      {'title': 'Change property', 'content' :
        {'type' : 'selector', 'action' : 'ChangeProperty', 'options' : [
          {'title' : 'DataSrcType', 'type':'boolean', 'options' : [
            'INTEGER','Integer32','Gauge32','UInteger32','Unsigned32','Counter32','Counter64','TimeTicks','BITS','ENUM','OCTETSTRING','OID','IpAddress','TIMETICKS','COUNTER32','COUNTER64','COUNTERXX','HWADDR','STRINGPARSER','STRINGEVAL','CONDITIONEVAL','BITSCHK'
            ]
          },
          {'title': 'Scale','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.compose([Validators.required,ValidationService.floatValidator]))
            })
          },
          {'title': 'Shift','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.compose([Validators.required,ValidationService.floatValidator]))
            })
          },
          {'title' : 'IsTag', 'type':'boolean', 'options' : [
            'true','false'
            ]
          }
        ]},
      }
    ];
    return tableAvailableActions;
  }

  getInfluxServersAvailableActions (data ? : any) : any {
    let tableAvailableActions = [
    //Remove Action
      {'title': 'Remove', 'content' :
        {'type' : 'button','action' : 'RemoveAllSelected'}
      },
    //Change Property Action
      {'title': 'Change property', 'content' :
        {'type' : 'selector', 'action' : 'ChangeProperty', 'options' : [
          {'title' : 'Precision', 'type':'boolean', 'options' : [
            'h','m','s','ms','u','ns']
          },
          {'title': 'Retention','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.required)
            })
          },
          {'title': 'Timeout','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator]))
            })
          }
        ]},
      }
    ];
    return tableAvailableActions;
  }

  getOutputAvailableActions(data?: any): any {
    let tableAvailableActions = [
      //Remove Action
      {
        'title': 'Remove', 'content':
          { 'type': 'button', 'action': 'RemoveAllSelected' }
      },
      //Change Property Action
      {
        'title': 'Change property', 'content':
        {
          'type': 'selector', 'action': 'ChangeProperty', 'options': [
            {
              'title': 'Active', 'type': 'boolean', 'options': [
                'true', 'false'
              ]
            },
            {'title': 'BufferSize','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl(0, Validators.required)
            })
          },
          ]
        },
      }
    ];
    return tableAvailableActions;
  }

  getMeasGroupsAvailableActions (data ? : any) : any {
    let tableAvailableActions = [
    //Remove Action
      {'title': 'Remove', 'content' :
        {'type' : 'button','action' : 'RemoveAllSelected'}
      }
    ];
    return tableAvailableActions;
  }

  getKafkaServersAvailableActions (data ? : any) : any {
    let tableAvailableActions = [
    //Remove Action
      {'title': 'Remove', 'content' :
        {'type' : 'button','action' : 'RemoveAllSelected'}
      }
    ];
    return tableAvailableActions;
  }

  getCustomFiltersAvailableActions (data ? : any) : any {
    let tableAvailableActions = [
    //Remove Action
      {'title': 'Remove', 'content' :
        {'type' : 'button','action' : 'RemoveAllSelected'}
      },
    ];
    return tableAvailableActions;
  }

  getOIDConditionsAvailableActions (data ? : any) : any {
    let tableAvailableActions = [
    //Remove Action
      {'title': 'Remove', 'content' :
        {'type' : 'button','action' : 'RemoveAllSelected'}
      },
    ];
    return tableAvailableActions;
  }

  getMeasFiltersAvailableActions (data ? : any) : any {
    let tableAvailableActions = [
    //Remove Action
      {'title': 'Remove', 'content' :
        {'type' : 'button','action' : 'RemoveAllSelected'}
      },
    ];
    return tableAvailableActions;
  }

  getMeasurementsAvailableActions (data ? : any) : any {
    let tableAvailableActions = [
    //Remove Action
      {'title': 'Remove', 'content' :
        {'type' : 'button','action' : 'RemoveAllSelected'}
      },
    //Change Property Action
      {'title': 'Change property', 'content' :
        {'type' : 'selector', 'action' : 'ChangeProperty', 'options' : [
          {'title': 'IndexTag','type':'input', 'options':
            new FormGroup({
              formControl : new FormControl('', Validators.required)
            })
          },
          {'title' : 'IndexAsValue', 'type':'boolean', 'options' : [
            'true','false'
            ]
          }
        ]},
      }
    ];
    return tableAvailableActions;
  }

  getVarCatalogAvailableActions (data ? : any) : any {
    let tableAvailableActions = [
    //Remove Action
      {'title': 'Remove', 'content' :
        {'type' : 'button','action' : 'RemoveAllSelected'}
      }
    ];
    return tableAvailableActions;
  }

}
