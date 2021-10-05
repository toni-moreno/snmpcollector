import { Component, ChangeDetectionStrategy, ViewChild,ViewContainerRef } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';
import { SnmpDeviceService } from '../snmpdevice/snmpdevicecfg.service';
import { InfluxServerService } from '../influxserver/influxservercfg.service';
import { MeasGroupService } from '../measgroup/measgroupcfg.service';
import { MeasFilterService } from '../measfilter/measfiltercfg.service';
import { VarCatalogService } from '../varcatalog/varcatalogcfg.service';
import { ValidationService } from '../common/validation.service';
import { PollerLocationService } from '../pollerlocations/pollerlocations.service';
import { Observable } from 'rxjs/Rx';
import { FormArray, FormGroup, FormControl} from '@angular/forms';
import { BlockUIService } from '../common/blockui/blockui-service';
import { BlockUIComponent } from '../common/blockui/blockui-component';

import { GenericModal } from '../common/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { TestConnectionModal } from '../common/test-connection-modal';
import { TestFilterModal } from '../customfilter/test-filter-modal'
import { ExportServiceCfg } from '../common/dataservice/export.service'
import { ItemsPerPageOptions } from '../common/global-constants';
import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

import { SpinnerComponent } from '../common/spinner';

import { TableListComponent } from '../common/table-list.component';
import { SnmpDeviceCfgComponentConfig, TableRole, OverrideRoleActions, ExtraActions } from './snmpdevicecfg.data';

declare var _:any;

@Component({
  selector: 'snmpdevs',
  providers: [SnmpDeviceService, InfluxServerService, MeasGroupService, MeasFilterService, PollerLocationService, VarCatalogService, BlockUIService],
  templateUrl: './snmpdeviceeditor.html',
  styleUrls: ['../css/component-styles.css']
})
export class SnmpDeviceCfgComponent {
  //TEST:
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('viewTestConnectionModal') public viewTestConnectionModal: TestConnectionModal;
  @ViewChild('viewTestFilterModal') public viewTestFilterModal: TestFilterModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;
  @ViewChild('blocker', { read: ViewContainerRef }) container: ViewContainerRef;

  editEnabled : boolean = false;
  selectedArray : any = [];

  public isRequesting : boolean;
  public counterItems : number = null;
  public counterErrors: any = [];

  itemsPerPageOptions : any = ItemsPerPageOptions;
  //ADDED
  editmode: string; //list , create, modify
  snmpdevs: Array<any>;
  filter: string;
  snmpdevForm: any;
  //influxservers: Array<any>;
  measfilters: Array<any>;
  measgroups: Array<any>;
  varcatalogs: Array<any>;
  pollerlocations: Array<any>;
  filteroptions: any;
  selectgroups: IMultiSelectOption[] = [];
  selectfilters: IMultiSelectOption[] = [];
  selectinfluxservers: IMultiSelectOption[] = [];
  selectvarcatalogs: IMultiSelectOption[] = [];
  selectpollerlocations: IMultiSelectOption[] = [];
  private mySettingsInflux: IMultiSelectSettings = {
      singleSelect: true,
  };
  alertHandler: any = [];

  myFilterValue: any;

  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public defaultConfig : any = SnmpDeviceCfgComponentConfig;
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

  varsArray: Array<Object> = [];
  selectedVars: Array<any> = [];
  public extraActions: any = ExtraActions;

  constructor(
    public snmpDeviceService: SnmpDeviceService,
    public varCatalogService: VarCatalogService,
    public influxserverDeviceService: InfluxServerService,
    public measgroupsDeviceService: MeasGroupService,
    public measfiltersDeviceService: MeasFilterService,
    public pollerlocationsService: PollerLocationService,
    public exportServiceCfg: ExportServiceCfg,
    builder: FormBuilder, private _blocker: BlockUIService
  ) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  enableEdit() {
    let obsArray = [];
    //obsArray.push(this.getMeasFiltersforDevices);
    Observable.forkJoin([this.measgroupsDeviceService.getMeasGroup(null),this.measfiltersDeviceService.getMeasFilter(null)])
              .subscribe(
                data => {
                  this.tableAvailableActions = new AvailableTableActions('snmpdevicecfg',[this.createMultiselectArray(data[0]),this.createMultiselectArray(data[1])]).availableOptions;
                },
                err => console.log(err),
                () => console.log()
              );
    }

  createStaticForm() {
    this.snmpdevForm = this.builder.group({
      ID: [this.snmpdevForm ? this.snmpdevForm.value.ID : '', Validators.required],
      Host: [this.snmpdevForm ? this.snmpdevForm.value.Host : '', Validators.compose([Validators.required, ValidationService.hostNameValidator])],
      Port: [this.snmpdevForm ? this.snmpdevForm.value.Port : 161, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      Retries: [this.snmpdevForm ? this.snmpdevForm.value.Retries : 5, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      Timeout: [this.snmpdevForm ? this.snmpdevForm.value.Timeout : 20, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      Active: [this.snmpdevForm ? this.snmpdevForm.value.Active : 'true', Validators.required],
      SnmpVersion: [this.snmpdevForm ? this.snmpdevForm.value.SnmpVersion : '2c', Validators.required],
      DisableBulk: [this.snmpdevForm ? this.snmpdevForm.value.DisableBulk : 'false'],
      MaxOids: [this.snmpdevForm ? this.snmpdevForm.value.MaxOids : 60, Validators.compose([Validators.required,ValidationService.uintegerNotZeroValidator])],
      MaxRepetitions: [this.snmpdevForm ? this.snmpdevForm.value.MaxRepetitions : 50, Validators.compose([Validators.required,ValidationService.uinteger8NotZeroValidator])],
      Freq: [this.snmpdevForm ? this.snmpdevForm.value.Freq : 60, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      UpdateFltFreq: [this.snmpdevForm ? this.snmpdevForm.value.UpdateFltFreq : 60, Validators.compose([Validators.required, ValidationService.uintegerAndLessOneValidator])],
      ConcurrentGather: [this.snmpdevForm ? this.snmpdevForm.value.ConcurrentGather : 'true', Validators.required],
      Location: [this.snmpdevForm ? this.snmpdevForm.value.Location : '', Validators.required],
      OutDB: [this.snmpdevForm ? this.snmpdevForm.value.OutDB : '', Validators.required],
      LogLevel: [this.snmpdevForm ? this.snmpdevForm.value.LogLevel : 'info', Validators.required],
      SnmpDebug: [this.snmpdevForm ? this.snmpdevForm.value.SnmpDebug : 'false', Validators.required],
      DeviceTagName: [this.snmpdevForm ? this.snmpdevForm.value.DeviceTagName : '', Validators.required],
      DeviceTagValue: [this.snmpdevForm ? this.snmpdevForm.value.DeviceTagValue : 'id'],
      ExtraTags: [this.snmpdevForm ? (this.snmpdevForm.value.ExtraTags ? this.snmpdevForm.value.ExtraTags : "" ) : "" , Validators.compose([ ValidationService.extraTags])],
      SystemOIDs: [this.snmpdevForm ? (this.snmpdevForm.value.SystemOIDs ? this.snmpdevForm.value.SystemOIDs : "" ) : "" , Validators.compose([ValidationService.noWhiteSpaces, ValidationService.extraTags])],
      DeviceVars: [this.snmpdevForm ? this.snmpdevForm.value.DeviceVars : null],
      MeasurementGroups: [this.snmpdevForm ? this.snmpdevForm.value.MeasurementGroups : null],
      MeasFilters: [this.snmpdevForm ? this.snmpdevForm.value.MeasFilters : null],
      Description: [this.snmpdevForm ? this.snmpdevForm.value.Description : ''],
    });
  }

  createDynamicForm(fieldsArray: any) : void {

    //Generates the static form:
    //Saves the actual to check later if there are shared values
    let tmpform : any;
    if (this.snmpdevForm)  tmpform = this.snmpdevForm.value;
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
      this.snmpdevForm.addControl(entry.ID, new FormControl(value, entry.Validators));
    }
}

  setDynamicFields (field : any, override? : boolean) : void  {
    //Saves on the array all values to push into formGroup
    let controlArray : Array<any> = [];

    switch (field) {
      case 'AuthPriv':
      controlArray.push({'ID': 'V3ContextEngineID', 'defVal' : '', 'Validators' : Validators.nullValidator });
      controlArray.push({'ID': 'V3ContextName', 'defVal' : '', 'Validators' : Validators.nullValidator });
      controlArray.push({'ID': 'V3PrivPass', 'defVal' : '', 'Validators' : Validators.required });
      controlArray.push({'ID': 'V3PrivProt', 'defVal' : '', 'Validators' : Validators.required });
      case 'AuthNoPriv':
      controlArray.push({'ID': 'V3ContextEngineID', 'defVal' : '', 'Validators' : Validators.nullValidator });
      controlArray.push({'ID': 'V3ContextName', 'defVal' : '', 'Validators' : Validators.nullValidator });
      controlArray.push({'ID': 'V3AuthPass', 'defVal' : '', 'Validators' : Validators.required });
      controlArray.push({'ID': 'V3AuthProt', 'defVal' : '', 'Validators' : Validators.required });
      case 'NoAuthNoPriv':
      controlArray.push({'ID': 'V3ContextEngineID', 'defVal' : '', 'Validators' : Validators.nullValidator });
      controlArray.push({'ID': 'V3ContextName', 'defVal' : '', 'Validators' : Validators.nullValidator });
      controlArray.push({'ID': 'V3SecLevel', 'defVal' : field, 'Validators' : Validators.required });
      controlArray.push({'ID': 'V3AuthUser', 'defVal' : '', 'Validators' : Validators.required });
      break;
      case '3':
      controlArray.push({'ID': 'V3ContextEngineID', 'defVal' : '', 'Validators' : Validators.nullValidator });
      controlArray.push({'ID': 'V3ContextName', 'defVal' : '', 'Validators' : Validators.nullValidator });
      controlArray.push({'ID': 'V3SecLevel', 'defVal' : 'NoAuthNoPriv', 'Validators' : Validators.required });
      controlArray.push({'ID': 'V3AuthUser', 'defVal' : '', 'Validators' : Validators.required });
      break;
      case '1':
      controlArray.push({'ID': 'Community', 'defVal' : 'public', 'Validators' : Validators.required });
      break;
      case '2c':
      controlArray.push({'ID': 'Community', 'defVal' : 'public', 'Validators' : Validators.required });
      break;
      case 'Location':
        controlArray.push({ 'ID': 'Location', 'defVal': '', 'Validators': Validators.required });
      default: //Gauge32
      controlArray.push({'ID': 'SnmpVersion', 'defVal' : '2c', 'Validators' : Validators.required });
      controlArray.push({'ID': 'Community', 'defVal' : 'public', 'Validators' : Validators.required });
      break;
    }
    //Reload the formGroup with new values saved on controlArray
    this.createDynamicForm(controlArray);
  }

  reloadData() {
    //In order to avoid issues with the array we clean it every time we refresh data
    this.selectedArray = [];
    this.isRequesting = true;
    // now it's a simple subscription to the observable
    this.snmpDeviceService.getDevices(null)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.snmpdevs = data;
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
        this.newDevice()
      case 'view':
        this.viewItem(action.event);
      break;
      case 'edit':
        this.editDevice(action.event);
      break;
      case 'remove':
        this.removeItem(action.event);
      break;
      case 'editenabled':
        this.enableEdit();
      break;
      case 'test-connection':
        this.showTestConnectionModal(action.event);
      break;
      case 'tableaction':
        this.applyAction(action.event, action.data);
      break;
    }
  }

  onExtraActionClicked(data: any) {
    console.log(data);
    switch (data.action) {
      case 'changeruntime':
        this.listChangeRTSnmpDev(data.row,data.property)
        break;
      case 'updateruntime':
        this.listChangeRTSnmpDev(data.row)
        break;
      case 'deletefull':
        this.deleteSnmpDevice(data.row.ID,'full',false);
      break;
      default:
      break;
    }
  }

  listChangeRTSnmpDev(data,property?) {
    console.log(property)
    if (property) {
        if (data[property] === false) {
          this._blocker.start(this.container, "Adding new device on runtime. Please wait...");
          this.snmpDeviceService.addDevice(data,'runtime')
          .subscribe(data => { this.editmode = "list"; this._blocker.stop(); this.reloadData() },
          err => {console.error(err),this._blocker.stop();},
          () => { console.log("DONE") }
          );
      
        } else if (data[property] === true) {
          this._blocker.start(this.container, "Removing new device on runtime. Please wait...");
          this.snmpDeviceService.deleteDevice(data.ID,'runtime')
          .subscribe(data => { this.editmode = "list"; this._blocker.stop(); this.reloadData() },
          err => {console.error(err),this._blocker.stop();},
          () => { console.log("DONE") }
          );  
        } 
      } else {
        this._blocker.start(this.container, "Updating new device on runtime. Please wait...");
        this.snmpDeviceService.editDevice(data,data.ID,'runtime')
        .subscribe(data => { this.editmode = "list"; this._blocker.stop(); this.reloadData() },
          err => {console.error(err),this._blocker.stop();},
        () => { console.log("DONE") }
        );
    }
  }


  viewItem(id) {
    console.log("id: ", id);
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
      this.deleteSnmpDevice(myArray[i].ID,false,true);
      obsArray.push(this.deleteSnmpDevice(myArray[i].ID,false,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.snmpDeviceService.checkOnDeleteSNMPDevice(id)
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

  onChangevarsArray(id) {
    //Create the array with ID:
    let varArrayID = this.varsArray.map( x => {return x['ID']})
    let delEntries = _.differenceWith(varArrayID,id,_.isEqual);
    let newEntries = _.differenceWith(id,varArrayID,_.isEqual);
    //Remove detected delEntries
    _.remove(this.varsArray, function(n) {
      return delEntries.indexOf(n['ID']) != -1;
    });
    //Add new entries
    for (let a of newEntries) {
      this.varsArray.push ({'ID': a, 'value': ''});
    }
  }

  newDevice() {
    //Check for subhidden fields
    if (this.snmpdevForm) {
      this.setDynamicFields(this.snmpdevForm.value.SnmpVersion === '3' ? this.snmpdevForm.value.V3SecLevel : this.snmpdevForm.value.SnmpVersion);
    } else {
      this.setDynamicFields(null);
    }
    this.getInfluxServersforDevices();
    this.getMeasGroupsforDevices();
    this.getMeasFiltersforDevices();
    this.getVarCatalogsforDevices();
    this.getLocationsforDevices();
    this.editmode = "create";
  }

  editDevice(row) {
    let id = row.ID;
    //Get select options
    this.getInfluxServersforDevices();
    this.getMeasGroupsforDevices();
    this.getMeasFiltersforDevices();
    this.getVarCatalogsforDevices();
    this.getLocationsforDevices();

    this.snmpDeviceService.getDevicesById(id)
      .subscribe(data => {
        this.varsArray = [];
        this.selectedVars = [];
        this.snmpdevForm = {};
        this.snmpdevForm.value = data;
        if (data.DeviceVars) {
          for (var values of data.DeviceVars) {
            let id = values.split('=');
            this.varsArray.push({ 'ID': id[0], 'value': id[1] });
            this.selectedVars.push(id[0]);
          }
        }
        this.oldID = data.ID
        this.setDynamicFields(row.SnmpVersion === '3' ? row.V3SecLevel : row.SnmpVersion);
        this.editmode = "modify"
      },
      err => console.error(err),
    );
  }

  deleteSnmpDevice(id, mode, recursive?) {
    if (!recursive) {
      if (mode) {
        let r = confirm("Remove from runtime and configuration device "+id+". Proceed?");
         if (r === true) {
          this._blocker.start(this.container, "Removing device from runtime and configuration. Please wait...");
          this.snmpDeviceService.deleteDevice(id,mode)
          .subscribe(data => { this._blocker.stop() },
            err => {console.error(err); this._blocker.stop()},
            () => {this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData()}
            );
         }
         return
        } 
      this.snmpDeviceService.deleteDevice(id,mode)
      .subscribe(data => { },
        err => console.error(err),
        () => {this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData()}
        );
    } else {
      return this.snmpDeviceService.deleteDevice(id, mode,true)
      .do(
        (test) =>  { this.counterItems++},
        (err) => { this.counterErrors.push({'ID': id, 'error' : err})}
      );
    }
  }

  cancelEdit() {
    this.reloadData();
    this.viewTestConnectionModal.hide();
    this.editmode = "list";
  }

  saveSnmpDev(data,online?) {
    let varCatalogsID : Array<any> = [];
      for (let i of this.varsArray) {
        varCatalogsID.push(i['ID']+(i['value'] ? '='+i['value'] : ''));
      }
      data['DeviceVars']=varCatalogsID;
      if (online) {
        this._blocker.start(this.container, "Adding new device on runtime. Please wait...");
      }
      this.snmpDeviceService.addDevice(data,online)
        .subscribe(data => { this._blocker.stop(); console.log(data) },
        err =>  { console.error(err); this._blocker.stop(); },
        () => { this.editmode = "list"; this.reloadData() }
        );
   }

updateAllSelectedItems(mySelectedArray,field,value, append?) {
  let obsArray = [];
  this.counterItems = 0;
  this.isRequesting = true;
  if (!append)
  for (let component of mySelectedArray) {
    component[field] = value;
    obsArray.push(this.updateSnmpDev(true,component));
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
      obsArray.push(this.updateSnmpDev(true,component));
    }
  }
  this.genericForkJoin(obsArray);
  //Make sync calls and wait the result
  this.counterErrors = [];
}

  updateSnmpDev(recursive,component,mode?) {
    if(!recursive) {
      var r = true;
      if (component.ID != this.oldID) {
        r = confirm("Changing Device ID from " + this.oldID + " to " + component.ID + ". Proceed?");
      }
      if (r == true) {
        let varCatalogsID : Array<any> = [];
        for (let i of this.varsArray) {
          varCatalogsID.push(i['ID']+(i['value'] ? '='+i['value'] : ''));
        }
        if(mode) {
          this._blocker.start(this.container, "Updating new device on runtime. Please wait...");
        }
        component['DeviceVars']=varCatalogsID;
        this.snmpDeviceService.editDevice(component, this.oldID,mode)
          .subscribe(data => { this._blocker.stop(); this.editmode = "list"; this.reloadData() },
          err =>  { console.error(err); this._blocker.stop(); },
          () => { }
          );
      }      
    }
    else {
      return this.snmpDeviceService.editDevice(component, component.ID,mode, true)
      .do(
        (test) =>  { this.counterItems++ },
        (err) => { this.counterErrors.push({'ID': component['ID'], 'error' : err})}
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

  showTestConnectionModal(data) {
      this.viewTestConnectionModal.show(data);
  }

  showFilterModal(){
    if(this.snmpdevForm.valid) {
      this.viewTestFilterModal.newCustomFilter(this.snmpdevForm.value);
    }
  }

  getMeasGroupsforDevices() {
    this.measgroupsDeviceService.getMeasGroup(null)
      .subscribe(
      data => {
        this.measgroups = data
        this.selectgroups = [];
        this.selectgroups = this.createMultiselectArray(data)
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  getInfluxServersforDevices() {
    this.influxserverDeviceService.getInfluxServer(null)
      .subscribe(
      data => {
      //  this.influxservers = data;
        this.selectinfluxservers = [];
        for (let entry of data) {
          console.log(entry)
          this.selectinfluxservers.push({ 'id': entry.ID, 'name': entry.ID });
        }
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  createMultiselectArray(tempArray) : any {
    let myarray = [];
    for (let entry of tempArray) {
      myarray.push({ 'id': entry.ID, 'name': entry.ID });
    }
    return myarray;
  }

  getMeasFiltersforDevices() {
    return this.measfiltersDeviceService.getMeasFilter(null)
      .subscribe(
      data => {
        this.measfilters = data
        this.selectfilters = [];
        this.selectfilters =  this.createMultiselectArray(data);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  getVarCatalogsforDevices() {
    return this.varCatalogService.getVarCatalog(null)
      .subscribe(
      data => {
        this.varcatalogs = data
        this.selectvarcatalogs = [];
        this.selectvarcatalogs =  this.createMultiselectArray(data);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }


  getLocationsforDevices() {
    this.pollerlocationsService.getPollerLocation(null)
      .subscribe(
        data => {
          this.pollerlocations = data
          this.selectpollerlocations = [];
          for (let entry of data) {
            this.selectpollerlocations.push({ 'id': entry.ID, 'name': entry.Instance_ID + ' - ' + entry.Location + ' (' + entry.IP + ')'});
          }
        },
        err => console.error(err),
        () => console.log('DONE')
      );
  }
}
