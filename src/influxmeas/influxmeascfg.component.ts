import { Component, ChangeDetectionStrategy, ViewChild} from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';
import { InfluxMeasService } from './influxmeascfg.service';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';
import { SnmpMetricService } from '../snmpmetric/snmpmetriccfg.service';
import { ValidationService } from '../common/validation.service'
import { FormArray, FormGroup, FormControl} from '@angular/forms';
import { ExportServiceCfg } from '../common/dataservice/export.service'
import { Observable } from 'rxjs/Rx';

import { GenericModal } from '../common/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { ItemsPerPageOptions } from '../common/global-constants';

import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';
import { TableListComponent } from '../common/table-list.component';
import { InfluxMeasCfgComponentConfig, TableRole, OverrideRoleActions } from './influxmeascfg.data';

declare var _:any;

@Component({
  selector: 'influxmeas',
  providers: [InfluxMeasService, SnmpMetricService],
  templateUrl: './influxmeaseditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class InfluxMeasCfgComponent {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

  selectedArray : any = [];
  public isRequesting : boolean;
  public counterItems : number = null;
  public counterErrors: any = [];
  public myTexts = {};
  public mySettings = {};

  itemsPerPageOptions : any = ItemsPerPageOptions;
  editmode: string; //list , create, modify
  influxmeas: Array<any>;
  filter: string;
  influxmeasForm: any;
  testinfluxmeas: any;
  snmpmetrics: Array<any>;
  selectmetrics: IMultiSelectOption[] = [];
  public defaultConfig : any = InfluxMeasCfgComponentConfig;
  public tableRole : any = TableRole;
  public overrideRoleActions: any = OverrideRoleActions;
  metricArray: Array<Object> = [];
  selectedMetrics: any = [];

  public reportMetricStatus: Array<Object> = [
    { value: 0, name: 'Never Report', icon: 'glyphicon glyphicon-remove-circle', class: 'text-danger' },
    { value: 1, name: 'Report', icon: 'glyphicon glyphicon-ok-circle', class: 'text-success' },
    { value: 2, name: 'Report if not zero', icon: 'glyphicon glyphicon-ban-circle', class: 'text-warning' }
  ];

  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];

  public tableAvailableActions : any;

  public page: number = 1;
  public itemsPerPage: number = 20;
  public maxSize: number = 5;
  public numPages: number = 1;
  public length: number = 0;
  myFilterValue: any;
  private builder;
  private oldID : string;
  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.defaultConfig['table-columns'] },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public influxMeasService: InfluxMeasService, public metricMeasService: SnmpMetricService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  createStaticForm() {
    this.influxmeasForm = this.builder.group({
      ID: [this.influxmeasForm ? this.influxmeasForm.value.ID : '', Validators.required],
      Name: [this.influxmeasForm ? this.influxmeasForm.value.Name : '', Validators.required],
      GetMode: [this.influxmeasForm ? this.influxmeasForm.value.GetMode : 'value', Validators.required],
      Fields: this.builder.array(this.influxmeasForm ? ((this.influxmeasForm.value.Fields) !== null ? this.influxmeasForm.value.Fields : []) : []),
      Description: [this.influxmeasForm ? this.influxmeasForm.value.Description : '']
    });
  }

  createDynamicForm(fieldsArray: any) : void {
    //Saves the actual to check later if there are shared values
    let tmpform : any;
    if (this.influxmeasForm)  tmpform = this.influxmeasForm.value;
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
      this.influxmeasForm.addControl(entry.ID, new FormControl(value, entry.Validators));
    }
  }

  setDynamicFields (field : any, override? : boolean) : void  {
    //Saves on the array all values to push into formGroup
    let controlArray : Array<any> = [];
    console.log(field);
    switch (field) {
      case 'indexed_it':
        controlArray.push({'ID': 'TagOID', 'defVal' : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required])});
      case 'indexed':
        controlArray.push({'ID': 'IndexOID', 'defVal' : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required])});
        controlArray.push({'ID': 'IndexDescrOID', 'defVal' : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required])});
        controlArray.push({'ID': 'IndexTag', 'defVal' : '', 'Validators' : Validators.required});
        controlArray.push({'ID': 'IndexTagFormat', 'defVal' : ''});
        controlArray.push({'ID': 'IndexAsValue', 'defVal' : 'false', 'Validators' : Validators.required})
        controlArray.push({'ID': 'Encoding', 'defVal' : ''});
      break;
      default:
        break;
    }
    //Reload the formGroup with new values saved on controlArray
    this.createDynamicForm(controlArray);
  }


  onChangeMetricArray(id) {

    //Create the array with ID:
    let testMetricID = this.metricArray.map( x => {return x['ID']})
    let delEntries = _.differenceWith(testMetricID,id,_.isEqual);
    let newEntries = _.differenceWith(id,testMetricID,_.isEqual);
    //Remove detected delEntries
    _.remove(this.metricArray, function(n) {
      return delEntries.indexOf(n['ID']) != -1;
    });
    //Add new entries
    for (let a of newEntries) {
      this.metricArray.push ({'ID': a, 'Report': 1});
    }
  }

  onCheckMetric(index: number, reportIndex: number) {
    this.metricArray[index]['Report'] = this.reportMetricStatus[reportIndex]['value'];
  }


  reloadData() {
    this.selectedArray = [];
    this.isRequesting = true;
    // now it's a simple subscription to the observable
    this.influxMeasService.getMeas(this.filter)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.influxmeas = data
        this.data = data;
      },
      err => console.error(err),
      () => { console.log('DONE'); }
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
        this.newMeas()
      case 'view':
        this.viewItem(action.event);
      break;
      case 'edit':
        this.editMeas(action.event);
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
      this.deleteInfluxMeas(myArray[i].ID,true);
      obsArray.push(this.deleteInfluxMeas(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.influxMeasService.checkOnDeleteInfluxMeas(id)
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

  newMeas() {
    if (this.influxmeasForm) {
      this.setDynamicFields(this.influxmeasForm.value.GetMode);
    } else {
      this.setDynamicFields(null);
    }
    this.editmode = "create";
    this.getMetricsforMeas();
  }

  editMeas(row) {
    let id = row.ID;
    this.metricArray = [];
    this.selectedMetrics = [];
    this.influxMeasService.getMeasById(id)
      .subscribe(data => {
        this.influxmeasForm = {};
        this.influxmeasForm.value = data;
        if (data.Fields) {
          for (var values of data.Fields) {
            this.metricArray.push({ ID: values.ID, Report: values.Report });
            this.selectedMetrics.push(values.ID);
          }
        }
        this.setDynamicFields(row.GetMode, false);
        this.getMetricsforMeas();
        this.oldID = data.ID
        this.editmode = "modify"
      },
      err => console.error(err)
      );
  }

  deleteInfluxMeas(id, recursive?) {
    if(!recursive) {
      this.influxMeasService.deleteMeas(id)
        .subscribe(data => { },
        err => console.error(err),
        () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
        );
    } else {
      return this.influxMeasService.deleteMeas(id,true)
      .do(
        (test) =>  { this.counterItems++},
        (err) => { this.counterErrors.push({'ID': id, 'error' : err})}
      );
    }
  }

  cancelEdit() {
    this.editmode = "list";
  }

  saveInfluxMeas() {
    this.influxmeasForm.value['Fields'] = this.metricArray;
    console.log(this.influxmeasForm.value);
    if (this.influxmeasForm.valid) {
      this.influxMeasService.addMeas(this.influxmeasForm.value)
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
      obsArray.push(this.updateInfluxMeas(true,component));
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
        obsArray.push(this.updateInfluxMeas(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateInfluxMeas(recursive?, component?) {
    if (!recursive) {
      if (this.influxmeasForm.valid) {
        var r = true;
        if (this.influxmeasForm.value.ID != this.oldID) {
          r = confirm("Changing Measurement ID from " + this.oldID + " to " + this.influxmeasForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.influxmeasForm.value['Fields'] = this.metricArray;
          this.influxMeasService.editMeas(this.influxmeasForm.value, this.oldID)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.influxMeasService.editMeas(component, component.ID, true)
      .do(
        (test) =>  { this.counterItems++ },
        (err) => { this.counterErrors.push({'ID': component['ID'], 'error' : err['_body']})}
      )
      .catch((err) => {
        return Observable.of({'ID': component.ID , 'error': err['_body']})
      })
    }
  }

  createMultiselectArray(tempArray) : any {
    let myarray = [];
    for (let entry of tempArray) {
      myarray.push({ 'id': entry.ID, 'name': entry.ID });
    }
    return myarray;
  }

  getMetricsforMeas() {
    this.metricMeasService.getMetrics(null)
      .subscribe(
      data => {
        this.snmpmetrics = data;
        this.selectmetrics = [];
        this.influxmeasForm.controls['Fields'].reset();
        this.selectmetrics = this.createMultiselectArray(data);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
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
