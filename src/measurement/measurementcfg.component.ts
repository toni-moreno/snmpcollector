import { Component, ChangeDetectionStrategy, ViewChild} from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';
import { MeasurementService } from './measurementcfg.service';
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
import { MeasurementCfgComponentConfig, TableRole, OverrideRoleActions } from './measurementcfg.data';

declare var _:any;

@Component({
  selector: 'measurement',
  providers: [MeasurementService, SnmpMetricService],
  templateUrl: './measurementeditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class MeasurementCfgComponent {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

  selectedArray : any = [];
  public isRequesting : boolean;
  public counterItems : number = null;
  public counterErrors: any = [];

  itemsPerPageOptions : any = ItemsPerPageOptions;
  editmode: string; //list , create, modify
  measurement: Array<any>;
  filter: string;
  measurementForm: any;
  testmeasurement: any;
  snmpmetrics: Array<any>;
  selectmetrics: IMultiSelectOption[] = [];
  public defaultConfig : any = MeasurementCfgComponentConfig;
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

  constructor(public measurementService: MeasurementService, public metricMeasService: SnmpMetricService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  createStaticForm() {
    this.measurementForm = this.builder.group({
      ID: [this.measurementForm ? this.measurementForm.value.ID : '', Validators.required],
      Name: [this.measurementForm ? this.measurementForm.value.Name : '', Validators.required],
      GetMode: [this.measurementForm ? this.measurementForm.value.GetMode : 'value', Validators.required],
      Freq: [this.measurementForm ? this.measurementForm.value.Freq : ''],
      UpdateFltFreq: [this.measurementForm ? this.measurementForm.value.UpdateFltFreq : ''],
      Fields: this.builder.array(this.measurementForm ? ((this.measurementForm.value.Fields) !== null ? this.measurementForm.value.Fields : []) : []),
      Description: [this.measurementForm ? this.measurementForm.value.Description : '']
    });
  }

  createDynamicForm(fieldsArray: any) : void {
    //Saves the actual to check later if there are shared values
    let tmpform : any;
    if (this.measurementForm)  tmpform = this.measurementForm.value;
    this.createStaticForm();

    for (let entry of fieldsArray) {
      let value = entry.defVal;
      //Check if there are common values from the previous selected item
      if (tmpform) {
        if (tmpform[entry.ID] && entry.override !== true) {
          value = tmpform[entry.ID];
        }
      }

      console.log(entry.ID)
      if (entry.ID == "MultiTagOID") {
        this.measurementForm.addControl(entry.ID, entry.defVal);
        // if it has already values, load them passing it to function - addMultiIndex
        if (value == tmpform[entry.ID]) {
          for (let val of value) {
            let p = this.addMultiTagOID(val)
            this.measurementForm.get("MultiTagOID").push(p)

          }
        }
        continue
      }

      //Set different controls:
      // if MultiIndex, the added control must be an special FormArray, not FormControl and we have to map its values first
      if (entry.ID == "MultiIndexCfg") {
        this.measurementForm.addControl(entry.ID, entry.defVal);
        // if it has already values, load them passing it to function - addMultiIndex
        if (value == tmpform[entry.ID]) {
          for (let val of value) {
            let p = this.addMultiIndex(val.GetMode, val)
            this.measurementForm.get("MultiIndexCfg").push(p)

          }
        }
        
        continue
      }
      this.measurementForm.addControl(entry.ID, new FormControl(value, entry.Validators));
    }
  }

  createDynamicFields(field: any, defVal : any = {} ) : any {
    let controlArray : Array<any> = [];
    console.log(field);
    switch (field) {
      case 'indexed_it':
        controlArray.push({'ID': 'TagOID', 'defVal' : defVal["TagOID"] ?  defVal["TagOID"] : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required])});
      case 'indexed':
        controlArray.push({'ID': 'UpdateFltFreq', 'defVal' : defVal["UpdateFltFreq"] ? defVal["UpdateFltFreq"] : '' });
        controlArray.push({'ID': 'IndexOID', 'defVal' : defVal["IndexOID"] ? defVal["IndexOID"] : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required])});
        controlArray.push({'ID': 'IndexTag', 'defVal' : defVal["IndexTag"] ? defVal["IndexTag"] : '', 'Validators' : Validators.required});
        controlArray.push({'ID': 'IndexTagFormat', 'defVal' : defVal["IndexTagFormat"] ? defVal["IndexTagFormat"] : ''});
        controlArray.push({'ID': 'IndexAsValue', 'defVal' : defVal["IndexAsValue"] ? defVal["IndexAsValue"] : "false", 'Validators' : Validators.required});
      break;
      case 'indexed_mit':
        controlArray.push({'ID': 'UpdateFltFreq', 'defVal' : defVal["UpdateFltFreq"] ? defVal["UpdateFltFreq"] : '' });
        controlArray.push({'ID': 'IndexOID', 'defVal' : defVal["IndexOID"] ? defVal["IndexOID"] : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required])});
        controlArray.push({'ID': 'MultiTagOID', 'defVal' : this.builder.array([]), 'Validators': Validators.compose([ValidationService.notEmpty, Validators.required])});
        controlArray.push({'ID': 'IndexTag', 'defVal' : defVal["IndexTag"] ? defVal["IndexTag"] : '', 'Validators' : Validators.required});
        controlArray.push({'ID': 'IndexTagFormat', 'defVal' : defVal["IndexTagFormat"] ? defVal["IndexTagFormat"] : ''});
        controlArray.push({'ID': 'IndexAsValue', 'defVal' : defVal["IndexAsValue"] ? defVal["IndexAsValue"] : "false", 'Validators' : Validators.required});
        break
      case 'indexed_multiple':
        controlArray.push({'ID': 'UpdateFltFreq', 'defVal' : defVal["UpdateFltFreq"] ? defVal["UpdateFltFreq"] : ''});
        controlArray.push({'ID': 'MultiIndexResult', 'defVal': defVal["MultiIndexResult"], 'Validators': Validators.required});
        controlArray.push({'ID': 'MultiIndexCfg', 'defVal' : this.builder.array([])});
        break;
      case 'value':
      default:
        break;
    }
    return controlArray
  }

  setDynamicFields (field : any, override?: boolean) : void  {
    //Saves on the array all values to push into formGroup
    let controlArray = this.createDynamicFields(field)
    //Reload the formGroup with new values saved on controlArray
    this.createDynamicForm(controlArray);
  }

  get MultiIndexCfg(): FormArray {
    return this.measurementForm.get("MultiIndexCfg") as FormArray
  }

  get MultiTagOID(): FormArray {
    console.log("this.measurementform.get(MultiTagOID)", this.measurementForm.get("MultiTagOID"))
    return this.measurementForm.get("MultiTagOID") as FormArray
  }

  MIMultiTagOID(i : string): FormArray {
    console.log("this.measurementform.get(MultiTagOID)", this.measurementForm.get("MultiTagOID"))
    return this.MultiIndexCfg.controls[i].get("MultiTagOID") as FormArray
  }


  addMIMultiTagOID(i: number, fieldArray?) {

      //let p = this.createDynamicFields(getMode, fieldArray)
      let bb = this.builder.group({})

      // for (let entry of p) {
      //   bb.addControl(entry.ID, new FormControl(entry.defVal, entry.Validators))
      // }
  
      //Add special fields, label and description:
      bb.addControl("TagOID", new FormControl(fieldArray ? fieldArray.TagOID : '',Validators.compose([ValidationService.OIDValidator, Validators.required])));
      bb.addControl("IndexFormat", new FormControl(fieldArray ? fieldArray.IndexFormat : ''));
      
      // if (fieldArray) {
      //   return bb
      // }
      console.log()
      this.measurementForm.get("MultiIndexCfg").controls[i].get("MultiTagOID").push(bb)
  }

  getMIMultiTagOID(i: number) {
    return this.measurementForm.get("MultiIndexCfg").controls[i].get("MultiTagOID")
  }



  // MULTI TAG OID
  addMultiTagOID(fieldArray?) {
    //let p = this.createDynamicFields(getMode, fieldArray)
    let bb = this.builder.group({})

    // for (let entry of p) {
    //   bb.addControl(entry.ID, new FormControl(entry.defVal, entry.Validators))
    // }

    //Add special fields, label and description:
    bb.addControl("TagOID", new FormControl(fieldArray ? fieldArray.TagOID : '',Validators.compose([ValidationService.OIDValidator, Validators.required])));
    bb.addControl("IndexFormat", new FormControl(fieldArray ? fieldArray.IndexFormat : ''));
    
    if (fieldArray) {
      return bb
    }
    this.measurementForm.get("MultiTagOID").push(bb);
  }

  removeTagOID(i: number) {
    this.MultiTagOID.removeAt(i);
  }

  promoteTagOID(i: number) {
    let p = this.MultiTagOID.at(i)
    this.removeTagOID(i)
    this.MultiTagOID.insert(i - 1, p)
  }

  demoteTagOID(i: number) {
    let p = this.MultiTagOID.at(i)
    this.removeTagOID(i)
    this.MultiTagOID.insert(i + 1, p)
  }

  // MULTI INDEX - MULTI TAG OID
  removeMITagOID(i: number, j: number) {
    this.getMIMultiTagOID(i).removeAt(j);
  }

  promoteMITagOID(i: number, j:number) {
    let p = this.getMIMultiTagOID(i).at(j)
    this.removeMITagOID(i,j)
    this.getMIMultiTagOID(i).insert(j - 1, p)
  }

  demoteMITagOID(i: number, j:number) {
    let p = this.getMIMultiTagOID(i).at(j)
    this.removeMITagOID(i,j)
    this.getMIMultiTagOID(i).insert(j + 1, p)
  }
  


  addMultiIndex(getMode: string, fieldArray?: any) {
    console.log("GOT GETMODE", getMode)
    let p = this.createDynamicFields(getMode, fieldArray)
    let bb = this.builder.group({})

    for (let entry of p) {
      if (entry.ID == "MultiTagOID") {
        bb.addControl(entry.ID, entry.defVal)
        if (fieldArray) {
        if (fieldArray.MultiTagOID.length > 0) {
        for (let fa of fieldArray.MultiTagOID) {
          let kk = this.builder.group({})
    
          //Add special fields, label and description:
          kk.addControl("TagOID", new FormControl(fa ? fa.TagOID : '',Validators.compose([ValidationService.OIDValidator, Validators.required])));
          kk.addControl("IndexFormat", new FormControl(fa ? fa.IndexFormat : ''));
          bb.controls.MultiTagOID.push(kk)

        }
      }
    } else {
        let kk = this.builder.group({})
        //Add special fields, label and description:
        kk.addControl("TagOID", new FormControl(fieldArray ? fieldArray.TagOID : '',Validators.compose([ValidationService.OIDValidator, Validators.required])));
        kk.addControl("IndexFormat", new FormControl(fieldArray ? fieldArray.IndexFormat : ''));
        bb.controls.MultiTagOID.push(kk)
      }
        continue
      }
      bb.addControl(entry.ID, new FormControl(entry.defVal, entry.Validators))
    }
    //Add special fields, label and description:
    bb.addControl("GetMode", new FormControl(getMode))
    bb.addControl("Label", new FormControl(fieldArray ? fieldArray.Label : 'Label', Validators.required));
    bb.addControl("Description", new FormControl(fieldArray ? fieldArray.Description : 'Description'));
    bb.addControl("Dependency", new FormControl(fieldArray ? fieldArray.Dependency : ''))

    if (fieldArray) {
      return bb
    }
    console.log(bb)
    this.measurementForm.get("MultiIndexCfg").push(bb);
  }

  removeMeas(i: number) {
    this.MultiIndexCfg.removeAt(i);
  }

  promoteMeas(i: number) {
    let p = this.MultiIndexCfg.at(i)
    this.removeMeas(i)
    this.MultiIndexCfg.insert(i - 1, p)
  }

  demoteMeas(i: number) {
    let p = this.MultiIndexCfg.at(i)
    this.removeMeas(i)
    this.MultiIndexCfg.insert(i + 1, p)
  }

  getMultiLabels() {
    let p = this.MultiIndexCfg.getRawValue()
    let kk = []
    for (let k in p) {
      kk.push({ 'id': k, 'name': k + '|' + p[k].Label })
    }
    console.log(kk)
    return kk
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
    this.measurementService.getMeas(this.filter)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.measurement = data
        this.data = data;
        // console.log(data)
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

  tableCellParser (data: any, type: string) {
    if (type === "multi") {
      var test: any = '<table>';
      if (data['multi']) {
      for (var i in data['multi']) {
          // declare val as value
          let val = data['multi'][i]
          // if empty, set up as "--"
          if (val === ""){
            val = "--"
          } 
          // if boolean, set up as icon
          if (typeof val === "boolean") {
            val = (val === true ? '<i class="glyphicon glyphicon-ok"></i>' : '<i class="glyphicon glyphicon-remove"></i>')
          }
          test +="<tr class='marg'>"
          test +="<td><span class=\"badge badge-multi\">"+ i +"</span></td><td>"+ val;
          test += "</td></tr>";
      }
      test += "</table>"
      return test
      }
    }
    if (type === "metrics") {        
       test = '<table>'
        for (var metric in data) {
            test += '<tr><td style="padding-right: 20px"><i class="padd ' + this.reportMetricStatus[data[metric]['Report']]['icon'] + ' ' + this.reportMetricStatus[data[metric]['Report']]['class'] + ' displayinline"></td><td>'+ data[metric]['ID'] +'</td>'
        }
        test += "</table>";
        return test
      }
      return data
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
      this.deleteMeasurement(myArray[i].ID,true);
      obsArray.push(this.deleteMeasurement(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.measurementService.checkOnDeleteMeasurement(id)
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
    if (this.measurementForm) {
      this.setDynamicFields(this.measurementForm.value.GetMode);
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
    this.measurementService.getMeasById(id)
      .subscribe(data => {
        this.measurementForm = {};
        this.measurementForm.value = data;
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

  deleteMeasurement(id, recursive?) {
    if(!recursive) {
      this.measurementService.deleteMeas(id)
        .subscribe(data => { },
        err => console.error(err),
        () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
        );
    } else {
      return this.measurementService.deleteMeas(id,true)
      .do(
        (test) =>  { this.counterItems++},
        (err) => { this.counterErrors.push({'ID': id, 'error' : err})}
      );
    }
  }

  cancelEdit() {
    this.editmode = "list";
  }

  saveMeasurement() {
    this.measurementForm.value['Fields'] = this.metricArray;
    console.log(this.measurementForm.value);
    if (this.measurementForm.valid) {
      this.measurementService.addMeas(this.measurementForm.value)
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
      obsArray.push(this.updateMeasurement(true,component));
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
        obsArray.push(this.updateMeasurement(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateMeasurement(recursive?, component?) {
    if (!recursive) {
      if (this.measurementForm.valid) {
        var r = true;
        if (this.measurementForm.value.ID != this.oldID) {
          r = confirm("Changing Measurement ID from " + this.oldID + " to " + this.measurementForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.measurementForm.value['Fields'] = this.metricArray;
          this.measurementService.editMeas(this.measurementForm.value, this.oldID)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.measurementService.editMeas(component, component.ID, true)
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
        this.measurementForm.controls['Fields'].reset();
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
