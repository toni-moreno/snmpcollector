import { Component, ChangeDetectionStrategy, ViewChild } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { FormArray, FormGroup, FormControl} from '@angular/forms';

import { InfluxServerService } from './influxservercfg.service';
import { ValidationService } from '../common/validation.service'
import { ExportServiceCfg } from '../common/dataservice/export.service'

import { GenericModal } from '../common/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { Observable } from 'rxjs/Rx';

import { ItemsPerPageOptions } from '../common/global-constants';
import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

import { TableListComponent } from '../common/table-list.component';
import { InfluxServerCfgComponentConfig, TableRole, OverrideRoleActions } from './influxservercfg.data';

declare var _:any;

@Component({
  selector: 'influxservers',
  providers: [InfluxServerService, ValidationService],
  templateUrl: './influxservereditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class InfluxServerCfgComponent {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

  itemsPerPageOptions : any = ItemsPerPageOptions;
  editmode: string; //list , create, modify
  influxservers: Array<any>;
  filter: string;
  influxserverForm: any;
  myFilterValue: any;
  alertHandler : any = null;


  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public tableAvailableActions : any;

  selectedArray : any = [];
  public defaultConfig : any = InfluxServerCfgComponentConfig;
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
  
  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.defaultConfig['table-columns'] },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public influxServerService: InfluxServerService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  createStaticForm() {
    this.influxserverForm = this.builder.group({
      ID: [this.influxserverForm ? this.influxserverForm.value.ID : '', Validators.required],
      Host: [this.influxserverForm ? this.influxserverForm.value.Host : '', Validators.required],
      Port: [this.influxserverForm ? this.influxserverForm.value.Port : '', Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      DB: [this.influxserverForm ? this.influxserverForm.value.DB : '', Validators.required],
      User: [this.influxserverForm ? this.influxserverForm.value.User : '', Validators.required],
      Password: [this.influxserverForm ? this.influxserverForm.value.Password : '', Validators.required],
      Retention: [this.influxserverForm ? this.influxserverForm.value.Retention : 'autogen', Validators.required],
      Precision: [this.influxserverForm ? this.influxserverForm.value.Precision : 's', Validators.required],
      Timeout: [this.influxserverForm ? this.influxserverForm.value.Timeout : 30, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      UserAgent: [this.influxserverForm ? this.influxserverForm.value.UserAgent : ''],
      EnableSSL: [this.influxserverForm ? this.influxserverForm.value.EnableSSL : 'false'],
      SSLCA: [this.influxserverForm ? this.influxserverForm.value.SSLCA : ''],
      SSLCert: [this.influxserverForm ? this.influxserverForm.value.SSLCert : ''],
      SSLKey: [this.influxserverForm ? this.influxserverForm.value.SSLKey : ''],
      InsecureSkipVerify: [this.influxserverForm ? this.influxserverForm.value.InsecureSkipVerify : 'true'],
      BufferSize: [this.influxserverForm ? this.influxserverForm.value.BufferSize : 65535, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      Description: [this.influxserverForm ? this.influxserverForm.value.Description : '']
    });
  }

  reloadData() {
    // now it's a simple subscription to the observable
    this.alertHandler = null;
    this.influxServerService.getInfluxServer(null)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.influxservers = data
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
        this.newInfluxServer()
      case 'view':
        this.viewItem(action.event);
      break;
      case 'edit':
        this.editInfluxServer(action.event);
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
      this.deleteInfluxServer(myArray[i].ID,true);
      obsArray.push(this.deleteInfluxServer(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.influxServerService.checkOnDeleteInfluxServer(id)
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
  newInfluxServer() {
    //No hidden fields, so create fixed Form
    this.createStaticForm();
    this.editmode = "create";
  }

  editInfluxServer(row) {
    let id = row.ID;
    this.influxServerService.getInfluxServerById(id)
      .subscribe(data => {
        this.influxserverForm = {};
        this.influxserverForm.value = data;
        this.oldID = data.ID
        this.createStaticForm();
        this.editmode = "modify";
      },
      err => console.error(err)
      );
 	}

  deleteInfluxServer(id, recursive?) {
    if (!recursive) {
    this.influxServerService.deleteInfluxServer(id)
      .subscribe(data => { },
      err => console.error(err),
      () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
      );
    } else {
      return this.influxServerService.deleteInfluxServer(id, true)
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

  saveInfluxServer() {
    if (this.influxserverForm.valid) {
      this.influxServerService.addInfluxServer(this.influxserverForm.value)
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
      obsArray.push(this.updateInfluxServer(true,component));
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
        obsArray.push(this.updateInfluxServer(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateInfluxServer(recursive?, component?) {
    if(!recursive) {
      if (this.influxserverForm.valid) {
        var r = true;
        if (this.influxserverForm.value.ID != this.oldID) {
          r = confirm("Changing Influx Server ID from " + this.oldID + " to " + this.influxserverForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.influxServerService.editInfluxServer(this.influxserverForm.value, this.oldID, true)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.influxServerService.editInfluxServer(component, component.ID)
      .do(
        (test) =>  { this.counterItems++ },
        (err) => { this.counterErrors.push({'ID': component['ID'], 'error' : err['_body']})}
      )
      .catch((err) => {
        return Observable.of({'ID': component.ID , 'error': err['_body']})
      })
    }
  }


  testInfluxServerConnection() {
    this.influxServerService.testInfluxServer(this.influxserverForm.value, true)
    .subscribe(
    data =>  this.alertHandler = {msg: 'Influx Version: '+data['Message'], result : data['Result'], elapsed: data['Elapsed'], type: 'success', closable: true},
    err => {
        let error = err.json();
        this.alertHandler = {msg: error['Message'], elapsed: error['Elapsed'], result : error['Result'], type: 'danger', closable: true}
      },
    () =>  { console.log("DONE")}
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
