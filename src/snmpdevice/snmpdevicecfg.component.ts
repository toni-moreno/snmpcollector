import { Component, ChangeDetectionStrategy, ViewChild } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';
import { SnmpDeviceService } from '../snmpdevice/snmpdevicecfg.service';
import { InfluxServerService } from '../influxserver/influxservercfg.service';
import { MeasGroupService } from '../measgroup/measgroupcfg.service';
import { MeasFilterService } from '../measfilter/measfiltercfg.service';
import { AlertModule } from 'ng2-bootstrap/ng2-bootstrap';
import { ValidationService } from '../common/validation.service'
import { Observable } from 'rxjs/Observable';
import { FormArray, FormGroup, FormControl} from '@angular/forms';


import { GenericModal } from '../common/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { TestConnectionModal } from '../common/test-connection-modal';
import { TestFilterModal } from '../customfilter/test-filter-modal'
import { ExportServiceCfg } from '../common/dataservice/export.service'


@Component({
  selector: 'snmpdevs',
  providers: [SnmpDeviceService, InfluxServerService, MeasGroupService, MeasFilterService],
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


  //ADDED
  editmode: string; //list , create, modify
  snmpdevs: Array<any>;
  filter: string;
  snmpdevForm: any;
  //influxservers: Array<any>;
  measfilters: Array<any>;
  measgroups: Array<any>;
  filteroptions: any;
  selectgroups: IMultiSelectOption[] = [];
  selectfilters: IMultiSelectOption[] = [];
  selectinfluxservers: IMultiSelectOption[] = [];
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
    { title: 'Snmp Version', name: 'SnmpVersion' },
    { title: 'Snmp Debug', name: 'SnmpDebug' },
    { title: 'Polling Period (sec)', name: 'Freq' },
    { title: 'Update Filter (Cicles)', name: 'UpdateFltFreq' },
    { title: 'Concurrent Gather', name: 'ConcurrentGather' },
    { title: 'Influx DB', name: 'OutDB' },
    { title: 'Log Level', name: 'LogLevel' },
    { title: 'Disable Snmp Bulk Queries', name: 'DisableBulk' },
    { title: 'Tag Name', name: 'DeviceTagName' },
    { title: 'Tag Value', name: 'DeviceTagValue' },
    { title: 'Extra Tags', name: 'ExtraTags' },
    { title: 'Measurement Groups', name: 'MeasurementGroups' },
    { title: 'Measurement Filters', name: 'MeasFilters' }
  ];

  public page: number = 1;
  public itemsPerPage: number = 10;
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



  constructor(public snmpDeviceService: SnmpDeviceService, public influxserverDeviceService: InfluxServerService, public measgroupsDeviceService: MeasGroupService, public measfiltersDeviceService: MeasFilterService,  public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  createStaticForm() {
    this.snmpdevForm = this.builder.group({
      ID: [this.snmpdevForm ? this.snmpdevForm.value.ID : '', Validators.required],
      Host: [this.snmpdevForm ? this.snmpdevForm.value.Host : '', Validators.required],
      Port: [this.snmpdevForm ? this.snmpdevForm.value.Port : 161, Validators.compose([Validators.required, ValidationService.integerValidator])],
      Retries: [this.snmpdevForm ? this.snmpdevForm.value.Retries : 5, Validators.compose([Validators.required, ValidationService.integerValidator])],
      Timeout: [this.snmpdevForm ? this.snmpdevForm.value.Timeout : 20, Validators.compose([Validators.required, ValidationService.integerValidator])],
      Active: [this.snmpdevForm ? this.snmpdevForm.value.Active : 'true', Validators.required],
      SnmpVersion: [this.snmpdevForm ? this.snmpdevForm.value.SnmpVersion : '2c', Validators.required],
      DisableBulk: [this.snmpdevForm ? this.snmpdevForm.value.DisableBulk : 'false'],
      Freq: [this.snmpdevForm ? this.snmpdevForm.value.Freq : 60, Validators.compose([Validators.required, ValidationService.integerValidator])],
      UpdateFltFreq: [this.snmpdevForm ? this.snmpdevForm.value.UpdateFltFreq : 60, Validators.compose([Validators.required, ValidationService.integerValidator])],
      ConcurrentGather: [this.snmpdevForm ? this.snmpdevForm.value.ConcurrentGather : 'true', Validators.required],
      OutDB: [this.snmpdevForm ? this.snmpdevForm.value.OutDB :  '', Validators.required],
      LogLevel: [this.snmpdevForm ? this.snmpdevForm.value.LogLevel : 'info', Validators.required],
      SnmpDebug: [this.snmpdevForm ? this.snmpdevForm.value.SnmpDebug : 'false', Validators.required],
      DeviceTagName: [this.snmpdevForm ? this.snmpdevForm.value.DeviceTagName : '', Validators.required],
      DeviceTagValue: [this.snmpdevForm ? this.snmpdevForm.value.DeviceTagValue : 'id'],
      ExtraTags: [this.snmpdevForm ? this.snmpdevForm.value.ExtraTags : null, Validators.compose([ValidationService.noWhiteSpaces, ValidationService.extraTags])],
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
      controlArray.push({'ID': 'V3PrivPass', 'defVal' : '', 'Validators' : Validators.required });
      controlArray.push({'ID': 'V3PrivProt', 'defVal' : '', 'Validators' : Validators.required });
      case 'AuthNoPriv':
      controlArray.push({'ID': 'V3AuthPass', 'defVal' : '', 'Validators' : Validators.required });
      controlArray.push({'ID': 'V3AuthProt', 'defVal' : '', 'Validators' : Validators.required });
      case 'NoAuthNoPriv':
      controlArray.push({'ID': 'V3SecLevel', 'defVal' : field, 'Validators' : Validators.required });
      controlArray.push({'ID': 'V3AuthUser', 'defVal' : '', 'Validators' : Validators.required });
      break;
      case '3':
      controlArray.push({'ID': 'V3SecLevel', 'defVal' : 'NoAuthNoPriv', 'Validators' : Validators.required });
      controlArray.push({'ID': 'V3AuthUser', 'defVal' : '', 'Validators' : Validators.required });
      break;
      case '1':
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
    // now it's a simple subscription to the observable
    this.snmpDeviceService.getDevices(null)
      .subscribe(
      data => {
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
          item[column.name] = '--'
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

  public onCellClick(data: any): any {
    console.log(data);
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

  newDevice() {
    //Check for subhidden fields
    if (this.snmpdevForm) {
      this.setDynamicFields(this.snmpdevForm.value.SnmpVersion === '3' ? this.snmpdevForm.value.V3SecLevel : this.snmpdevForm.value.SnmpVersion);
    } else {
      this.setDynamicFields(null);
    }

    this.editmode = "create";
    this.getInfluxServersforDevices();
    this.getMeasGroupsforDevices();
    this.getMeasFiltersforDevices();
  }

  editDevice(row) {
    let id = row.ID;
    //Get select options
    this.getInfluxServersforDevices();
    this.getMeasGroupsforDevices();
    this.getMeasFiltersforDevices();

    this.snmpDeviceService.getDevicesById(id)
      .subscribe(data => {
        this.snmpdevForm = {};
        this.snmpdevForm.value = data;
        this.snmpdevForm.value.ExtraTags = this.snmpdevForm.value.ExtraTags.join(',');
        this.oldID = data.ID
        this.setDynamicFields(row.SnmpVersion === '3' ? row.V3SecLevel : row.SnmpVersion);
        this.editmode = "modify"
      },
      err => console.error(err),
    );
  }
  deleteSnmpDevice(id) {
    this.snmpDeviceService.deleteDevice(id)
      .subscribe(data => { },
      err => console.error(err),
      () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
      );
  }
  cancelEdit() {
    this.reloadData();
    this.viewTestConnectionModal.hide();
    this.editmode = "list";
  }
  saveSnmpDev() {
    if (this.snmpdevForm.valid) {
      this.snmpDeviceService.addDevice(this.snmpdevForm.value)
        .subscribe(data => { console.log(data) },
        err => console.error(err),
        () => { this.editmode = "list"; this.reloadData() }
        );
    }
  }

  updateSnmpDev() {
    if (this.snmpdevForm.valid) {
      var r = true;
      if (this.snmpdevForm.value.ID != this.oldID) {
        r = confirm("Changing Device ID from " + this.oldID + " to " + this.snmpdevForm.value.ID + ". Proceed?");
      }
      if (r == true) {
        this.snmpDeviceService.editDevice(this.snmpdevForm.value, this.oldID)
          .subscribe(data => { console.log(data) },
          err => console.error(err),
          () => { this.editmode = "list"; this.reloadData() }
          );
      }
    }
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
        for (let entry of data) {
          console.log(entry)
          this.selectgroups.push({ 'id': entry.ID, 'name': entry.ID });
        }
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

  getMeasFiltersforDevices() {
    this.measfiltersDeviceService.getMeasFilter(null)
      .subscribe(
      data => {
        this.measfilters = data
        this.selectfilters = [];
        for (let entry of data) {
          console.log(entry)
          this.selectfilters.push({ 'id': entry.ID, 'name': entry.ID });
        }
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }
}
