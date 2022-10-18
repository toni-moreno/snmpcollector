import { Component, ChangeDetectionStrategy, ViewChild } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { FormArray, FormGroup, FormControl} from '@angular/forms';

import { KafkaServerService } from './kafkaservercfg.service';
import { ValidationService } from '../common/validation.service'
import { ExportServiceCfg } from '../common/dataservice/export.service'

import { GenericModal } from '../common/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { Observable } from 'rxjs/Rx';

import { ItemsPerPageOptions } from '../common/global-constants';
import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

import { TableListComponent } from '../common/table-list.component';
import { KafkaServerCfgComponentConfig, TableRole, OverrideRoleActions } from './kafkaservercfg.data';

declare var _:any;

@Component({
  selector: 'kafkaservers',
  providers: [KafkaServerService, ValidationService],
  templateUrl: './kafkaservereditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class KafkaServerCfgComponent {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

  itemsPerPageOptions : any = ItemsPerPageOptions;
  editmode: string; //list , create, modify
  kafkaservers: Array<any>;
  filter: string;
  kafkaserverForm: any;
  myFilterValue: any;
  alertHandler : any = null;


  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public tableAvailableActions : any;

  selectedArray : any = [];
  public defaultConfig : any = KafkaServerCfgComponentConfig;
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

  constructor(public kafkaServerService: KafkaServerService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  createStaticForm() {
    this.kafkaserverForm = this.builder.group({
      ID: [this.kafkaserverForm ? this.kafkaserverForm.value.ID : '', Validators.required],
      Brokers: [this.kafkaserverForm ? this.kafkaserverForm.value.Brokers : '', Validators.required],
      Topic: [this.kafkaserverForm ? this.kafkaserverForm.value.Topic : ''],
      TopicTag: [this.kafkaserverForm ? this.kafkaserverForm.value.TopicTag : ''],
      ExcludeTopicTag: [this.kafkaserverForm ? this.kafkaserverForm.value.ExcludeTopicTag : false],
      Version: [this.kafkaserverForm ? this.kafkaserverForm.value.Version : ''],
      ClientID: [this.kafkaserverForm ? this.kafkaserverForm.value.ClientID : 'Telegraf'],
      CompressionCodec: [this.kafkaserverForm ? this.kafkaserverForm.value.CompressionCodec : 0],
      Method: [this.kafkaserverForm ? this.kafkaserverForm.value.Method : 'measurement', Validators.required],
      Keys: [this.kafkaserverForm ? this.kafkaserverForm.value.Keys : ''],
      Separator: [this.kafkaserverForm ? this.kafkaserverForm.value.Separator : '_'],
      RoutingTag: [this.kafkaserverForm ? this.kafkaserverForm.value.RoutingTag : ''],
      RoutingKey: [this.kafkaserverForm ? this.kafkaserverForm.value.RoutingKey : ''],
      RequiredAcks: [this.kafkaserverForm ? this.kafkaserverForm.value.RequiredAcks : -1, Validators.compose([Validators.required, ValidationService.uintegerAndLessOneValidator])],
      MaxRetry: [this.kafkaserverForm ? this.kafkaserverForm.value.MaxRetry : 3, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      MaxMessageBytes: [this.kafkaserverForm ? this.kafkaserverForm.value.MaxMessageBytes : 0, ValidationService.integerValidator],
      IdempotentWrites: [this.kafkaserverForm ? this.kafkaserverForm.value.IdempotentWrites : false],
      EnableTLS: [this.kafkaserverForm ? this.kafkaserverForm.value.EnableTLS : false],
      Socks5ProxyEnabled: [this.kafkaserverForm ? this.kafkaserverForm.value.Socks5ProxyEnabled : false],
      MetadataFull: [this.kafkaserverForm ? this.kafkaserverForm.value.MetadataFull : false],
      Description: [this.kafkaserverForm ? this.kafkaserverForm.value.Description : '']
    });
  }


  createDynamicForm(fieldsArray: any) : void {

    //Generates the static form:
    //Saves the actual to check later if there are shared values
    let tmpform : any;
    if (this.kafkaserverForm)  tmpform = this.kafkaserverForm.value;
    this.createStaticForm();
    //Set new values and check if we have to mantain the value!
    for (let entry of fieldsArray) {
      let value = entry.defVal;
      //Check if there are common values from the previous selected item
      if (tmpform) {
        if (tmpform[entry.ID] && entry.override !== true) {
          value = tmpform[entry.ID];
        }
      }
      //Set different controls:
      this.kafkaserverForm.addControl(entry.ID, new FormControl(value, entry.Validators));
    }
}

checkFields(tls, socks5) {
  let dyn = ''
  if (tls == true || tls == 'true') {
    dyn = 'tls'
  }
  if (socks5 == true || socks5 == 'true') {
    dyn += 'socks5'
  }
  if (dyn == '') {
    dyn = null
  }
  return dyn
  }

setDynamicFields (field : any, override? : boolean) : void  {
  //Saves on the array all values to push into formGroup
  let controlArray : Array<any> = [];
  switch (field) {
    case 'tls':
    controlArray.push({'ID': 'ServerName', 'defVal' : '', 'Validators' : Validators.required });
    controlArray.push({'ID': 'TLSCA', 'defVal' : '', 'Validators' : Validators.nullValidator });
    controlArray.push({'ID': 'TLSCert', 'defVal' : '', 'Validators' : Validators.nullValidator });
    controlArray.push({'ID': 'TLSKey', 'defVal' : '', 'Validators' : Validators.required });
    controlArray.push({'ID': 'TLSMinVersion', 'defVal' : '', 'Validators' : Validators.required });
    controlArray.push({'ID': 'InsecureSkipVerify', 'defVal' : false, 'Validators' : Validators.required });
    break;
    case 'socks5':
    controlArray.push({'ID': 'Socks5ProxyAddress', 'defVal' : '', 'Validators' : Validators.required });
    controlArray.push({'ID': 'Socks5ProxyUsername', 'defVal' : '', 'Validators' : Validators.nullValidator });
    controlArray.push({'ID': 'Socks5ProxyPassword', 'defVal' : '', 'Validators' : Validators.nullValidator });
    break;
    case 'tlssocks5':
      controlArray.push({'ID': 'ServerName', 'defVal' : '', 'Validators' : Validators.required });
      controlArray.push({'ID': 'TLSCA', 'defVal' : '', 'Validators' : Validators.nullValidator });
      controlArray.push({'ID': 'TLSCert', 'defVal' : '', 'Validators' : Validators.nullValidator });
      controlArray.push({'ID': 'TLSKey', 'defVal' : '', 'Validators' : Validators.required });
      controlArray.push({'ID': 'TLSMinVersion', 'defVal' : '', 'Validators' : Validators.required });
      controlArray.push({'ID': 'InsecureSkipVerify', 'defVal' : false, 'Validators' : Validators.required });
      controlArray.push({'ID': 'Socks5ProxyAddress', 'defVal' : '', 'Validators' : Validators.required });
      controlArray.push({'ID': 'Socks5ProxyUsername', 'defVal' : '', 'Validators' : Validators.nullValidator });
      controlArray.push({'ID': 'Socks5ProxyPassword', 'defVal' : '', 'Validators' : Validators.nullValidator });  
  }

  //Reload the formGroup with new values saved on controlArray
  console.log("GOT CONTROL ARRAY:", controlArray)
  this.createDynamicForm(controlArray);
}


  reloadData() {
    //In order to avoid issues with the array we clean it every time we refresh data
    this.selectedArray = [];
    this.isRequesting = true;
    // now it's a simple subscription to the observable
    this.kafkaServerService.getKafkaServer(null)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.kafkaservers = data
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
        this.newKafkaServer()
      case 'view':
        this.viewItem(action.event);
      break;
      case 'edit':
        this.editKafkaServer(action.event);
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
      this.deleteKafkaServer(myArray[i].ID,true);
      obsArray.push(this.deleteKafkaServer(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.kafkaServerService.checkOnDeleteKafkaServer(id)
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
  newKafkaServer() {
    //No hidden fields, so create fixed Form
    if (this.kafkaserverForm) {
      this.setDynamicFields(this.checkFields(this.kafkaserverForm.value.EnableTLS,this.kafkaserverForm.value.Socks5ProxyEnabled));
    } else {
      this.setDynamicFields(null);
    }
    this.editmode = "create";
  }

  editKafkaServer(row) {
    let id = row.ID;
    this.kafkaServerService.getKafkaServerById(id)
      .subscribe(data => {
        this.kafkaserverForm = {};
        this.kafkaserverForm.value = data;
        this.oldID = data.ID
        this.setDynamicFields(this.checkFields(row.EnableTLS, row.Socks5ProxyEnabled))
        this.editmode = "modify";
      },
      err => console.error(err)
      );
 	}

  deleteKafkaServer(id, recursive?) {
    if (!recursive) {
    this.kafkaServerService.deleteKafkaServer(id)
      .subscribe(data => { },
      err => console.error(err),
      () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
      );
    } else {
      return this.kafkaServerService.deleteKafkaServer(id, true)
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

  saveKafkaServer() {
    if (this.kafkaserverForm.valid) {
      this.kafkaServerService.addKafkaServer(this.kafkaserverForm.value)
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
      obsArray.push(this.updateKafkaServer(true,component));
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
        obsArray.push(this.updateKafkaServer(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateKafkaServer(recursive?, component?) {
    if(!recursive) {
      if (this.kafkaserverForm.valid) {
        var r = true;
        if (this.kafkaserverForm.value.ID != this.oldID) {
          r = confirm("Changing Kafka Server ID from " + this.oldID + " to " + this.kafkaserverForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.kafkaServerService.editKafkaServer(this.kafkaserverForm.value, this.oldID, true)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.kafkaServerService.editKafkaServer(component, component.ID)
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
