import { Component, ChangeDetectionStrategy, ViewChild } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { FormArray, FormGroup, FormControl} from '@angular/forms';

import { KafkaServerService } from '../kafkaserver/kafkaservercfg.service';
import { InfluxServerService } from '../influxserver/influxservercfg.service';


import { OutputService } from './outputcfg.service';
import { ValidationService } from '../common/validation.service'
import { ExportServiceCfg } from '../common/dataservice/export.service'

import { GenericModal } from '../common/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { Observable } from 'rxjs/Rx';

import { ItemsPerPageOptions } from '../common/global-constants';
import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

import { TableListComponent } from '../common/table-list.component';
import { OutputCfgComponentConfig, TableRole, OverrideRoleActions } from './outputcfg.data';

import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';

declare var _:any;

@Component({
  selector: 'outputs',
  providers: [OutputService, ValidationService, KafkaServerService, InfluxServerService],
  templateUrl: './outputeditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class OutputCfgComponent {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

  itemsPerPageOptions : any = ItemsPerPageOptions;
  editmode: string; //list , create, modify
  outputs: Array<any>;
  filter: string;
  outputForm: any;
  myFilterValue: any;
  alertHandler : any = null;

  private mySettings: IMultiSelectSettings = {
    singleSelect: true,
    returnOption: true,
    uniqueSelect: true,
};


  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public tableAvailableActions : any;

  selectedArray : any = [];
  public defaultConfig : any = OutputCfgComponentConfig;
  public tableRole : any = TableRole;
  public overrideRoleActions: any = OverrideRoleActions;
  public isRequesting : boolean;
  public counterItems : number = null;
  public counterErrors: any = [];

  public page: number = 1;
  public itemsPerPage: number = 20;
  public maxSize: number = 5;
  public numPages: number = 1;
  public length: number = 0;
  private builder;
  private oldID : string;
  
  backends: Array<any>;
  selectbackends: IMultiSelectOption[] = [];


  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.defaultConfig['table-columns'] },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public outputService: OutputService, public influxServerService: InfluxServerService, public kafkaServerService: KafkaServerService,  public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  createStaticForm() {
    this.outputForm = this.builder.group({
      ID: [this.outputForm ? this.outputForm.value.ID : '', Validators.required],
      BackendType: [this.outputForm ? this.outputForm.value.BackendType : '', Validators.required],
      Active: [this.outputForm ? this.outputForm.value.Active : 'true', Validators.required],
      EnqueueOnError: [this.outputForm ? this.outputForm.value.EnqueueOnError : 'true', Validators.required],
      BufferSize: [this.outputForm ? this.outputForm.value.BufferSize : 131070, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      MetricBatchSize: [this.outputForm ? this.outputForm.value.MetricBatchSize : 15000, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      FlushInterval: [this.outputForm ? this.outputForm.value.FlushInterval : 60, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      Backend: [this.outputForm ? this.outputForm.value.Backend : '', Validators.required],
      Description: [this.outputForm ? this.outputForm.value.Description : '']
    });
  }

  reloadData() {
    // now it's a simple subscription to the observable
    this.alertHandler = null;
    this.outputService.getOutput(null)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.outputs = data
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
        this.newOutput()
      case 'view':
        this.viewItem(action.event);
      break;
      case 'edit':
        this.editOutput(action.event);
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
      this.deleteOutput(myArray[i].ID,true);
      obsArray.push(this.deleteOutput(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.outputService.checkOnDeleteOutput(id)
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
  newOutput() {
    //No hidden fields, so create fixed Form
    this.createStaticForm();
    this.getBackendsforOutput();
    this.editmode = "create";
  }

  
  editOutput(row) {
    let id = row.ID;
    this.getBackendsforOutput()
    this.outputService.getOutputById(id)
      .subscribe(data => {
        this.outputForm = {};
        this.outputForm.value = data;
        this.oldID = data.ID
        this.createStaticForm();
        this.editmode = "modify";
      },
      err => console.error(err)
      );
 	}

  deleteOutput(id, recursive?) {
    if (!recursive) {
    this.outputService.deleteOutput(id)
      .subscribe(data => { },
      err => console.error(err),
      () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
      );
    } else {
      return this.outputService.deleteOutput(id, true)
      .do(
        (test) =>  { this.counterItems++},
        (err) => { this.counterErrors.push({'ID': id, 'error' : err})}
      );
    }
  }

  cancelEdit() {
    this.editmode = "list";
    this.reloadData();
  }

  saveOutput() {
    if (this.outputForm.valid) {
      console.log(this.outputForm)
      this.outputService.addOutput(this.outputForm.value)
        .subscribe(data => { console.log(data) },
        err => {
          console.log(err);
        },
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
      obsArray.push(this.updateOutput(true,component));
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
        obsArray.push(this.updateOutput(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateOutput(recursive?, component?) {
    if(!recursive) {
      if (this.outputForm.valid) {
        var r = true;
        if (this.outputForm.value.ID != this.oldID) {
          r = confirm("Changing Output ID from " + this.oldID + " to " + this.outputForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.outputService.editOutput(this.outputForm.value, this.oldID, true)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.outputService.editOutput(component, component.ID)
      .do(
        (test) =>  { this.counterItems++ },
        (err) => { this.counterErrors.push({'ID': component['ID'], 'error' : err['_body']})}
      )
      .catch((err) => {
        return Observable.of({'ID': component.ID , 'error': err['_body']})
      })
    }
  }

  selectBackend(event: any) {
    let composename = event.split("..")
    if (event != "" && composename.length > 1) {
      for (let entry of this.selectbackends) {
        if (composename[0] == entry["id"] && composename[1] == entry["badge"]) {
          this.outputForm.controls["BackendType"].setValue(composename[1])
        }
      }
    }
  }

  getBackendsforOutput() {
    this.selectbackends = [];
    this.backends = [];

    this.influxServerService.getInfluxServer(null)
      .subscribe(
      data => {
        this.backends.push(data);
        for (let entry of data) {
          console.log(entry)
          this.selectbackends.push({ 'id': entry.ID, 'name': entry.ID, 'badge': "influxdb", 'parent': true });
        }
       },
      err => console.error(err),
      () => console.log('DONE')
      );
    this.kafkaServerService.getKafkaServer(null)
    .subscribe(
      data => {
        this.backends.push(data);
        for (let entry of data) {
          console.log(entry)
          this.selectbackends.push({ 'id': entry.ID, 'name': entry.ID, 'badge': "kafka", 'parent': true });
        }
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
