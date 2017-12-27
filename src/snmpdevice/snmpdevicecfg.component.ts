import { Component, ChangeDetectionStrategy, ViewChild } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';
import { SnmpDeviceService } from '../snmpdevice/snmpdevicecfg.service';
import { InfluxServerService } from '../influxserver/influxservercfg.service';
import { MeasGroupService } from '../measgroup/measgroupcfg.service';
import { MeasFilterService } from '../measfilter/measfiltercfg.service';
import { VarCatalogService } from '../varcatalog/varcatalogcfg.service';
import { ValidationService } from '../common/validation.service';
import { Observable } from 'rxjs/Rx';
import { FormArray, FormGroup, FormControl} from '@angular/forms';


import { GenericModal } from '../common/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { TestConnectionModal } from '../common/test-connection-modal';
import { TestFilterModal } from '../customfilter/test-filter-modal'
import { ExportServiceCfg } from '../common/dataservice/export.service'
import { ItemsPerPageOptions } from '../common/global-constants';
import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

import { SpinnerComponent } from '../common/spinner';

declare var _:any;

@Component({
  selector: 'snmpdevs',
  providers: [SnmpDeviceService, InfluxServerService, MeasGroupService, MeasFilterService, VarCatalogService],
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
  filteroptions: any;
  selectgroups: IMultiSelectOption[] = [];
  selectfilters: IMultiSelectOption[] = [];
  selectinfluxservers: IMultiSelectOption[] = [];
  selectvarcatalogs: IMultiSelectOption[] = [];
  private mySettingsInflux: IMultiSelectSettings = {
      singleSelect: true,
  };
  alertHandler: any = [];

  myFilterValue: any;

  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public columns: Array<any> = [
    { title: 'ID', name: 'ID' },
    { title: 'Host', name: 'Host' },
    { title: 'Port', name: 'Port' },
    { title: 'Active', name: 'Active' },
    { title: 'Alternate System OIDs', name: 'SystemOIDs' },
    { title: 'Snmp Version', name: 'SnmpVersion' },
    { title: 'Snmp Debug', name: 'SnmpDebug' },
    { title: 'Polling Period (sec)', name: 'Freq' },
    { title: 'Update Filter (Cycles)', name: 'UpdateFltFreq' },
    { title: 'Concurrent Gather', name: 'ConcurrentGather' },
    { title: 'Influx DB', name: 'OutDB' },
    { title: 'Log Level', name: 'LogLevel' },
    { title: 'Disable Snmp Bulk Queries', name: 'DisableBulk' },
    { title: 'Timeout', name: 'Timeout' },
    { title: 'Retries', name: 'Retries' },
    { title: 'SNMP Max Repetitions', name: 'MaxRepetitions' },
    { title: 'Tag Name', name: 'DeviceTagName' },
    { title: 'Tag Value', name: 'DeviceTagValue' },
    { title: 'Extra Tags', name: 'ExtraTags' },
    { title: 'Device Variables', name: 'DeviceVars' },
    { title: 'Measurement Groups', name: 'MeasurementGroups' },
    { title: 'Measurement Filters', name: 'MeasFilters' }
  ];
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
    sorting: { columns: this.columns },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  varsArray: Array<Object> = [];
  selectedVars: Array<any> = [];

  constructor(public snmpDeviceService: SnmpDeviceService, public varCatalogService: VarCatalogService, public influxserverDeviceService: InfluxServerService, public measgroupsDeviceService: MeasGroupService, public measfiltersDeviceService: MeasFilterService,  public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  enableEdit() {
    this.editEnabled = !this.editEnabled;
    let obsArray = [];
    //obsArray.push(this.getMeasFiltersforDevices);
    Observable.forkJoin([this.measgroupsDeviceService.getMeasGroup(null),this.measfiltersDeviceService.getMeasFilter(null)])
              .subscribe(
                data => {
                  this.tableAvailableActions = new AvailableTableActions('device',[this.createMultiselectArray(data[0]),this.createMultiselectArray(data[1])]).availableOptions;
                },
                err => console.log(err),
                () => console.log()
              );
    }

  createStaticForm() {
    this.snmpdevForm = this.builder.group({
      ID: [this.snmpdevForm ? this.snmpdevForm.value.ID : '', Validators.required],
      Host: [this.snmpdevForm ? this.snmpdevForm.value.Host : '', Validators.required],
      Port: [this.snmpdevForm ? this.snmpdevForm.value.Port : 161, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      Retries: [this.snmpdevForm ? this.snmpdevForm.value.Retries : 5, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      Timeout: [this.snmpdevForm ? this.snmpdevForm.value.Timeout : 20, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      Active: [this.snmpdevForm ? this.snmpdevForm.value.Active : 'true', Validators.required],
      SnmpVersion: [this.snmpdevForm ? this.snmpdevForm.value.SnmpVersion : '2c', Validators.required],
      DisableBulk: [this.snmpdevForm ? this.snmpdevForm.value.DisableBulk : 'false'],
      MaxRepetitions: [this.snmpdevForm ? this.snmpdevForm.value.MaxRepetitions : 50, Validators.compose([Validators.required,ValidationService.uinteger8NotZeroValidator])],
      Freq: [this.snmpdevForm ? this.snmpdevForm.value.Freq : 60, Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator])],
      UpdateFltFreq: [this.snmpdevForm ? this.snmpdevForm.value.UpdateFltFreq : 60, Validators.compose([Validators.required, ValidationService.uintegerAndLessOneValidator])],
      ConcurrentGather: [this.snmpdevForm ? this.snmpdevForm.value.ConcurrentGather : 'true', Validators.required],
      OutDB: [this.snmpdevForm ? this.snmpdevForm.value.OutDB :  '', Validators.required],
      LogLevel: [this.snmpdevForm ? this.snmpdevForm.value.LogLevel : 'info', Validators.required],
      SnmpDebug: [this.snmpdevForm ? this.snmpdevForm.value.SnmpDebug : 'false', Validators.required],
      DeviceTagName: [this.snmpdevForm ? this.snmpdevForm.value.DeviceTagName : '', Validators.required],
      DeviceTagValue: [this.snmpdevForm ? this.snmpdevForm.value.DeviceTagValue : 'id'],
      ExtraTags: [this.snmpdevForm ? (this.snmpdevForm.value.ExtraTags ? this.snmpdevForm.value.ExtraTags : "" ) : "" , Validators.compose([ValidationService.noWhiteSpaces, ValidationService.extraTags])],
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
        this.onChangeTable(this.config);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  onResetFilter() : void {
    this.page = 1;
    this.myFilterValue = "";
    this.config.filtering = {filtering: { filterString: '' }};
    this.onChangeTable(this.config);
  }

  applyAction(test : any) : void {
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

  public changePage(page: any, data: Array<any> = this.data): Array<any> {
    //Check if we have to change the actual page

    let maxPage =  Math.ceil(data.length/this.itemsPerPage);
    if (page.page > maxPage && page.page != 1) this.page = page.page = maxPage;
    let start = (page.page - 1) * page.itemsPerPage;
    let end = page.itemsPerPage > -1 ? (start + page.itemsPerPage) : data.length;
    return data.slice(start, end);
  }

  public changeSort(data: any, config: any): any {
    if (!config.sorting) {
      return data;
    }

    let columns = this.config.sorting.columns || [];
    let columnName: string = void 0;
    let sort: string = void 0;

    for (let i = 0; i < columns.length; i++) {
      if (columns[i].sort !== '' && columns[i].sort !== false) {
        columnName = columns[i].name;
        sort = columns[i].sort;
      }
    }

    if (!columnName) {
      return data;
    }

    // simple sorting
    return data.sort((previous: any, current: any) => {
      if (previous[columnName] > current[columnName]) {
        return sort === 'desc' ? -1 : 1;
      } else if (previous[columnName] < current[columnName]) {
        return sort === 'asc' ? -1 : 1;
      }
      return 0;
    });
  }

  public changeFilter(data: any, config: any): any {
    let filteredData: Array<any> = data;
    this.columns.forEach((column: any) => {
      if (column.filtering) {
        filteredData = filteredData.filter((item: any) => {
          return item[column.name].match(column.filtering.filterString);
        });
      }
    });

    if (!config.filtering) {
      return filteredData;
    }

    if (config.filtering.columnName) {
      return filteredData.filter((item: any) =>
        item[config.filtering.columnName].match(this.config.filtering.filterString));
    }

    let tempArray: Array<any> = [];
    filteredData.forEach((item: any) => {
      let flag = false;
      this.columns.forEach((column: any) => {
        if (item[column.name] === null) {
        item[column.name] = ''
        }
        if (item[column.name].toString().match(this.config.filtering.filterString)) {
          flag = true;
        }

      });
      if (flag) {
        tempArray.push(item);
      }
    });
    filteredData = tempArray;

    return filteredData;
  }

  changeItemsPerPage (items) {
    this.itemsPerPage = parseInt(items);
    let maxPage =  Math.ceil(this.length/this.itemsPerPage);
    if (this.page > maxPage) this.page = maxPage;
    this.onChangeTable(this.config);
  }

  public onChangeTable(config: any, page: any = { page: this.page, itemsPerPage: this.itemsPerPage }): any {
    if (config.filtering) {
      Object.assign(this.config.filtering, config.filtering);
    }
    if (config.sorting) {
      Object.assign(this.config.sorting, config.sorting);
    }
    let filteredData = this.changeFilter(this.data, this.config);
    let sortedData = this.changeSort(filteredData, this.config);
    this.rows = page && config.paging ? this.changePage(page, sortedData) : sortedData;
    this.length = sortedData.length;
  }

  onFilter() {
    this.reloadData();
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
      this.deleteSnmpDevice(myArray[i].ID,true);
      obsArray.push(this.deleteSnmpDevice(myArray[i].ID,true));
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
    this.editmode = "create";
  }

  editDevice(row) {
    let id = row.ID;
    //Get select options
    this.getInfluxServersforDevices();
    this.getMeasGroupsforDevices();
    this.getMeasFiltersforDevices();
    this.getVarCatalogsforDevices();

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

  deleteSnmpDevice(id, recursive?) {
    if (!recursive) {
      this.snmpDeviceService.deleteDevice(id)
      .subscribe(data => { },
        err => console.error(err),
        () => {this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData()}
        );
    } else {
      return this.snmpDeviceService.deleteDevice(id, true)
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

  saveSnmpDev() {
    if (this.snmpdevForm.valid) {
      let varCatalogsID : Array<any> = [];
      for (let i of this.varsArray) {
        varCatalogsID.push(i['ID']+(i['value'] ? '='+i['value'] : ''));
      }
      this.snmpdevForm.value['DeviceVars']=varCatalogsID;
      this.snmpDeviceService.addDevice(this.snmpdevForm.value)
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

  updateSnmpDev(recursive?,component?) {
    if(!recursive) {
      if (this.snmpdevForm.valid) {
        var r = true;
        if (this.snmpdevForm.value.ID != this.oldID) {
          r = confirm("Changing Device ID from " + this.oldID + " to " + this.snmpdevForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          let varCatalogsID : Array<any> = [];
          for (let i of this.varsArray) {
            varCatalogsID.push(i['ID']+(i['value'] ? '='+i['value'] : ''));
          }
          this.snmpdevForm.value['DeviceVars']=varCatalogsID;
          this.snmpDeviceService.editDevice(this.snmpdevForm.value, this.oldID)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    }
    else {
      return this.snmpDeviceService.editDevice(component, component.ID, true)
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

  showTestConnectionModal() {
    if (this.snmpdevForm.valid) {
      this.viewTestConnectionModal.show(this.snmpdevForm.value);
    }
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
}
