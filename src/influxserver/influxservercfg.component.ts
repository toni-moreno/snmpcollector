import { Component, ChangeDetectionStrategy, ViewChild } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { FormArray, FormGroup, FormControl} from '@angular/forms';

import { InfluxServerService } from './influxservercfg.service';
import { ValidationService } from '../common/validation.service'
import { ExportServiceCfg } from '../common/dataservice/export.service'

import { GenericModal } from '../common/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';

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


  editmode: string; //list , create, modify
  influxservers: Array<any>;
  filter: string;
  influxserverForm: any;
  myFilterValue: any;
  alertHandler : any = null;


  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public columns: Array<any> = [
    { title: 'ID', name: 'ID' },
    { title: 'Host', name: 'Host' },
    { title: 'Port', name: 'Port' },
    { title: 'DB', name: 'DB' },
    { title: 'User', name: 'User' },
    { title: 'Retention', name: 'Retention' },
    { title: 'Precision', name: 'Precision' },
    { title: 'Timeout', name: 'Timeout' },
    { title: 'User Agent', name: 'UserAgent' }
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
      Description: [this.influxserverForm ? this.influxserverForm.value.Description : '']
    });
  }

  reloadData() {
    // now it's a simple subscription to the observable
    this.alertHandler = null;
    this.influxServerService.getInfluxServer(null)
      .subscribe(
      data => {
        this.influxservers = data
        this.data = data;
        this.onChangeTable(this.config)
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

  viewItem(id, event) {
    console.log('view', id);
    this.viewModal.parseObject(id);
  }

  exportItem(item : any) : void {
    this.exportFileModal.initExportModal(item);
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

  deleteInfluxServer(id) {
    this.influxServerService.deleteInfluxServer(id)
      .subscribe(data => { },
      err => console.error(err),
      () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
      );
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

  updateInfluxServer() {
    if (this.influxserverForm.valid) {
      var r = true;
      if (this.influxserverForm.value.ID != this.oldID) {
        r = confirm("Changing Influx Server ID from " + this.oldID + " to " + this.influxserverForm.value.ID + ". Proceed?");
      }
      if (r == true) {
        this.influxServerService.editInfluxServer(this.influxserverForm.value, this.oldID)
          .subscribe(data => { console.log(data) },
          err => console.error(err),
          () => { this.editmode = "list"; this.reloadData() }
          );
      }
    }
  }


  testInfluxServerConnection() {
    this.influxServerService.testInfluxServer(this.influxserverForm.value)
    .subscribe(
    data =>  this.alertHandler = {msg: 'Influx Version: '+data['Message'], result : data['Result'], elapsed: data['Elapsed'], type: 'success', closable: true},
    err => {
        let error = err.json();
        this.alertHandler = {msg: error['Message'], elapsed: error['Elapsed'], result : error['Result'], type: 'danger', closable: true}
      },
    () =>  { console.log("DONE")}
  );

  }

}
