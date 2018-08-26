import { Component, ChangeDetectionStrategy, ViewChild  } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';
import { SnmpMetricService } from './snmpmetriccfg.service';
import { OidConditionService } from '../oidcondition/oidconditioncfg.service';
import { ControlMessagesComponent } from '../common/control-messages.component'
import { ValidationService } from '../common/validation.service'
import { FormArray, FormGroup, FormControl} from '@angular/forms';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { Observable } from 'rxjs/Rx';

import { GenericModal } from '../common/generic-modal';
import { ExportServiceCfg } from '../common/dataservice/export.service'
import { ItemsPerPageOptions } from '../common/global-constants';

import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

import { TableListComponent } from '../common/table-list.component';
import { SnmpMetricCfgComponentConfig, TableRole, OverrideRoleActions } from './snmpmetriccfg.data';

declare var _:any;

@Component({
  selector: 'snmpmetrics',
  providers: [SnmpMetricService,OidConditionService],
  templateUrl: './snmpmetriceditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class SnmpMetricCfgComponent {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

  selectedArray : any = [];

  public isRequesting : boolean;
  public counterItems : number = null;
  public counterErrors: any = [];
  public conversionModes: Array<any>;

  itemsPerPageOptions : any = ItemsPerPageOptions;
  editmode: string; //list , create, modify
  snmpmetrics: Array<any>;
  filter: string;
  snmpmetForm: any;
  myFilterValue: any;
  //OID selector
  oidconditions: Array<any>;
  selectoidcond: IMultiSelectOption[] = [];
  private mySettings: IMultiSelectSettings = {
       singleSelect: true,
 };

  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];

  public defaultConfig : any = SnmpMetricCfgComponentConfig;
  public tableRole : any = TableRole;
  public overrideRoleActions: any = OverrideRoleActions;

  public tableAvailableActions : any;
  public page: number = 1;
  public itemsPerPage: number = 20;
  public maxSize: number = 5;
  public numPages: number = 1;
  public length: number = 0;
  private builder;
  private oldID : string;
  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.defaultConfig['table-columns'] },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public snmpMetricService: SnmpMetricService, public oidCondService: OidConditionService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  createStaticForm() {
    this.snmpmetForm = this.builder.group({
      ID: [this.snmpmetForm ? this.snmpmetForm.value.ID : '', Validators.required],
      FieldName: [this.snmpmetForm ? this.snmpmetForm.value.FieldName : '', Validators.required],
      DataSrcType: [this.snmpmetForm ? this.snmpmetForm.value.DataSrcType : 'Gauge32', Validators.required],
      Description: [this.snmpmetForm ? this.snmpmetForm.value.Description : ''],
      Conversion: [this.snmpmetForm ? this.snmpmetForm.value.Conversion : 0]
    });
  }


  createDynamicForm(fieldsArray: any) : void {
    //Saves the actual to check later if there are shared values
    let tmpform : any;
    if (this.snmpmetForm)  tmpform = this.snmpmetForm.value;
    this.createStaticForm();

    for (let entry of fieldsArray) {
      let value = entry.defVal;
      //Check if there are common values from the previous selected item
      if (tmpform) {
        if (tmpform[entry.ID] && entry.override !== true) {
          value = tmpform[entry.ID];
        }
      }
      //Set different controls:
      this.snmpmetForm.addControl(entry.ID, new FormControl(value, entry.Validators));
    }
  }

  setDynamicFields (field : any, override? : boolean) : void  {
    //Saves on the array all values to push into formGroup
    let controlArray : Array<any> = [];
    let value : any;

    if (this.snmpmetForm ) {
      //need a clone of the object 
      value = { ...this.snmpmetForm.value}
      //force required fields to be not null to get Conversion modes on the cloned object
      value['ID']='test'
      value['FieldName']='test'
    } else {
      value = { ID: 'test', FieldName: 'test', DataSrcType: 'Gauge32' };
    }
    this.snmpMetricService.getConversionModes(value)
    .subscribe(data => {
      this.conversionModes = data;
      console.log(this.conversionModes)
      },
      err => console.error(err)
    );
 

    switch (field) {
      case 'BITS':
      case 'BITSCHK':
      case 'ENUM':
        controlArray.push({'ID': 'ExtraData', 'defVal' : '', 'Validators' : Validators.required, 'override' : override });
        controlArray.push({'ID': 'IsTag', 'defVal' : 'false', 'Validators' : Validators.required, 'override' : override });
      case 'OCTETSTRING':
      case 'HWADDR':
      case 'IpAddress':
        controlArray.push({'ID': 'BaseOID', 'defVal' : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required]) })
        controlArray.push({'ID': 'IsTag', 'defVal' : 'false', 'Validators' : Validators.required, 'override' : override });
        controlArray.push({'ID': 'Conversion', 'defVal' : 3, 'Validators' : Validators.required, 'override' : override })
        break;
      case 'CONDITIONEVAL':
        this.getOidCond();
        controlArray.push({'ID': 'ExtraData', 'defVal' : '', 'Validators' : Validators.required, 'override' : override });
        controlArray.push({'ID': 'IsTag', 'defVal' : 'false', 'Validators' : Validators.required, 'override' : override });        
        controlArray.push({'ID': 'Conversion', 'defVal' : 1, 'Validators' : Validators.required, 'override' : override })
        break;
      case 'STRINGPARSER':
        controlArray.push({'ID': 'BaseOID', 'defVal' : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required]) })
        controlArray.push({'ID': 'ExtraData', 'defVal' : '', 'Validators' : Validators.required, 'override' : override });
        controlArray.push({'ID': 'IsTag', 'defVal' : 'false', 'Validators' : Validators.required, 'override' : override });
        controlArray.push({'ID': 'Scale', 'defVal' : '0', 'Validators' : Validators.compose([Validators.required, ValidationService.floatValidator]) })
        controlArray.push({'ID': 'Shift', 'defVal' : '0', 'Validators' : Validators.compose([Validators.required, ValidationService.floatValidator]) })
        controlArray.push({'ID': 'Conversion', 'defVal' : 0, 'Validators' : Validators.required, 'override' : override })
        break;
      case 'MULTISTRINGPARSER':
        controlArray.push({'ID': 'BaseOID', 'defVal' : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required]) })
        controlArray.push({'ID': 'ExtraData', 'defVal' : '', 'Validators' : Validators.required, 'override' : override });
        break;
      case 'STRINGEVAL':
        controlArray.push({'ID': 'IsTag', 'defVal' : 'false', 'Validators' : Validators.required, 'override' : override });
        controlArray.push({'ID': 'ExtraData', 'defVal' : '', 'Validators' : Validators.required, 'override' : override });
        controlArray.push({'ID': 'Scale', 'defVal' : '0', 'Validators' : Validators.compose([Validators.required, ValidationService.floatValidator]) })
        controlArray.push({'ID': 'Shift', 'defVal' : '0', 'Validators' : Validators.compose([Validators.required, ValidationService.floatValidator]) })
        controlArray.push({'ID': 'Conversion', 'defVal' : 0, 'Validators' : Validators.required, 'override' : override })
        break;
      case 'COUNTER32':
      case 'COUNTER64':
      case 'COUNTERXX':
        controlArray.push({'ID': 'GetRate', 'defVal' : 'false', 'Validators' : Validators.required});
        controlArray.push({'ID': 'IsTag', 'defVal' : 'false', 'Validators' : Validators.required, 'override' : override });
      default: //Gauge32
        controlArray.push({'ID': 'BaseOID', 'defVal' : '', 'Validators' :Validators.compose([ValidationService.OIDValidator, Validators.required]) })
        controlArray.push({'ID': 'Scale', 'defVal' : '0', 'Validators' : Validators.compose([Validators.required, ValidationService.floatValidator]) })
        controlArray.push({'ID': 'Shift', 'defVal' : '0', 'Validators' : Validators.compose([Validators.required, ValidationService.floatValidator]) })
        controlArray.push({'ID': 'IsTag', 'defVal' : 'false', 'Validators' : Validators.required, 'override' : override });
        controlArray.push({'ID': 'Conversion', 'defVal' : 1, 'Validators' : Validators.required, 'override' : override })
        break;
    }
    //Reload the formGroup with new values saved on controlArray
    this.createDynamicForm(controlArray);
  }

  getOidCond() {
    this.oidCondService.getConditions(null)
      .subscribe(
      data => {
        this.oidconditions = data;
        this.selectoidcond = [];
        for (let entry of data) {
          console.log(entry)
          this.selectoidcond.push({ 'id': entry.ID, 'name': entry.ID });
        }
      },
      err => console.error(err),
      () => { console.log('DONE') }
      );
  }

  reloadData() {
    //In order to avoid issues with the array we clean it every time we refresh data
    this.selectedArray = [];
    this.isRequesting = true;
    // now it's a simple subscription to the observable
    this.snmpMetricService.getMetrics(null)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.snmpmetrics = data;
        this.data = data;
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  applyAction(test : any, data? : Array<any>) : void {
    this.selectedArray = data || [];
    switch(test.action) {
       case "RemoveAllSelected": {
          this.removeAllSelectedItems(this.selectedArray);
          break;
       }
       case "ChangeProperty": {
          this.updateAllSelectedItems(this.selectedArray,test.field,test.value)
          break;
       }
       case "AppendProperty": {
         this.updateAllSelectedItems(this.selectedArray,test.field,test.value,true);
       }
       default: {
          break;
       }
    }
  }

  customActions(action : any) {
    switch (action.option) {
      case 'export' : 
        this.exportItem(action.event);
      break;
      case 'new' :
        this.newMetric()
      case 'view':
        this.viewItem(action.event);
      break;
      case 'edit':
        this.editMetric(action.event);
      break;
      case 'remove':
        this.removeItem(action.event);
      break;
      case 'tableaction':
        this.applyAction(action.event, action.data);
      break;
    }
  }

  viewItem(id) {
    console.log('view', id);
    this.viewModal.parseObject(id);
  }

  exportItem(item : any) : void {
    this.exportFileModal.initExportModal(item);
  }

  removeAllSelectedItems(myArray) {
    let obsArray = [];
    this.counterItems = 0;
    this.isRequesting = true;
    for (let i in myArray) {
      console.log("Removing ",myArray[i].ID)
      this.deleteSNMPMetric(myArray[i].ID,true);
      obsArray.push(this.deleteSNMPMetric(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.snmpMetricService.checkOnDeleteMetric(id)
      .subscribe(
      data => {
        console.log(data);
        let temp = data;
        this.viewModalDelete.parseObject(temp)
      },
      err => console.error(err),
      () => { }
      );
  }
  newMetric() {
    if (this.snmpmetForm) {
      this.setDynamicFields(this.snmpmetForm.value.DataSrcType);
    } else {
      this.setDynamicFields(null);
    }
    this.editmode = "create";
  }

  editMetric(row) {
    let id = row.ID;
    this.snmpMetricService.getMetricsById(id)
      .subscribe(data => {
        this.snmpmetForm = {};
        this.snmpmetForm.value = data;
        this.setDynamicFields(row.DataSrcType, false);
        this.oldID = data.ID
        this.editmode = "modify"
       },
      err => console.error(err)
      );
  }
  deleteSNMPMetric(id, recursive?) {
    if (!recursive) {
      this.snmpMetricService.deleteMetric(id)
        .subscribe(data => { },
        err => console.error(err),
        () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
        );
    } else {
      return this.snmpMetricService.deleteMetric(id, true)
      .do(
        (test) =>  { this.counterItems++},
        (err) => { this.counterErrors.push({'ID': id, 'error' : err})}
      );
    }
  }

  cancelEdit() {
    this.editmode = "list";
  }

  saveSnmpMet() {
    console.log(this.snmpmetForm);
    if (this.snmpmetForm.valid) {
      this.snmpMetricService.addMetric(this.snmpmetForm.value)
        .subscribe(data => { console.log(data) },
        err => console.error(err),
        () => { this.editmode = "list"; this.reloadData() }
        );
    }
  }

  updateAllSelectedItems(mySelectedArray,field,value, append?) {
    let obsArray = [];
    this.counterItems = 0;
    this.isRequesting = true;
    if (!append)
    for (let component of mySelectedArray) {
      component[field] = value;
      obsArray.push(this.updateSnmpMet(true,component));
    } else {
      let tmpArray = [];
      if(!Array.isArray(value)) value = value.split(',');
      console.log(value);
      for (let component of mySelectedArray) {
        console.log(value);
        //check if there is some new object to append
        let newEntries = _.differenceWith(value,component[field],_.isEqual);
        tmpArray = newEntries.concat(component[field])
        console.log(tmpArray);
        component[field] = tmpArray;
        obsArray.push(this.updateSnmpMet(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateSnmpMet(recursive?, component?) {
    if(!recursive) {
      if (this.snmpmetForm.valid) {
        var r = true;
        if (this.snmpmetForm.value.ID != this.oldID) {
          r = confirm("Changing Metric ID from " + this.oldID + " to " + this.snmpmetForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.snmpMetricService.editMetric(this.snmpmetForm.value, this.oldID)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.snmpMetricService.editMetric(component, component.ID, true)
      .do(
        (test) =>  { this.counterItems++ },
        (err) => { this.counterErrors.push({'ID': component['ID'], 'error' : err['_body']})}
      )
      .catch((err) => {
        return Observable.of({'ID': component.ID , 'error': err['_body']})
      })
    }
  }

  genericForkJoin(obsArray: any) {
    Observable.forkJoin(obsArray)
              .subscribe(
                data => {
                  this.selectedArray = [];
                  this.reloadData()
                },
                err => console.error(err),
              );
  }
}
