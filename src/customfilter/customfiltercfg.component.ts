import { Component, ChangeDetectionStrategy, ViewChild } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { CustomFilterService } from './customfilter.service';
import { SnmpDeviceService } from '../snmpdevice/snmpdevicecfg.service';
import { ExportServiceCfg } from '../common/dataservice/export.service'

import { ValidationService } from '../common/validation.service'
import { Observable } from 'rxjs/Rx';

import { GenericModal } from '../common/generic-modal';
import { TestFilterModal } from './test-filter-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { ItemsPerPageOptions } from '../common/global-constants';
import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

declare var _:any;

@Component({
  selector: 'customfilters',
  providers: [CustomFilterService, ValidationService, SnmpDeviceService],
  templateUrl: './customfiltereditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class CustomFilterCfgComponent {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('viewTestFilterModal') public viewTestFilterModal: TestFilterModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

  public tableAvailableActions : any;

  editEnabled : boolean = false;
  selectedArray : any = [];
  public isRequesting : boolean;
  public counterItems : number = null;
  public counterErrors: any = [];

  itemsPerPageOptions : any = ItemsPerPageOptions;

  editmode: string; //list , create, modify
  customfilters: Array<any>;
  filter: string;
  customfilterForm: any;
  testinfluxservers: any;
  myFilterValue: any;

  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public columns: Array<any> = [
    { title: 'ID', name: 'ID' },
    { title: 'RelatedDev', name: 'RelatedDev' },
    { title: 'RelatedMeas', name: 'RelatedMeas' },
    { title: 'Items', name: 'Items' },
  ];

  public page: number = 1;
  public itemsPerPage: number = 20;
  public maxSize: number = 5;
  public numPages: number = 1;
  public length: number = 0;

  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.columns },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public customFilterService: CustomFilterService, public snmpDeviceService: SnmpDeviceService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
  }

  enableEdit() {
    this.editEnabled = !this.editEnabled;
    console.log(this.editEnabled);
    let obsArray = [];
    this.tableAvailableActions = new AvailableTableActions('customfilter').availableOptions;
  }


  reloadData() {
    this.selectedArray = [];
    this.isRequesting = true;
    // now it's a simple subscription to the observable
    this.customFilterService.getCustomFilter(null)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.customfilters = data
        this.data = data;
        this.onChangeTable(this.config)
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  applyAction(test : any) : void {
    switch(test.action) {
       case "RemoveAllSelected": {
          this.removeAllSelectedItems(this.selectedArray);
          break;
       }
       default: {
          break;
       }
    }
  }

  onResetFilter(): void {
    this.page = 1;
    this.myFilterValue = "";
    this.config.filtering = { filtering: { filterString: '' } };
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

  onFilter() {
    this.reloadData();
  }

  viewItem(id, event) {
    console.log('view', id);
    this.viewModal.parseObject(id);
  }

  removeAllSelectedItems(myArray) {
    let obsArray = [];
    this.counterItems = 0;
    this.isRequesting = true;
    for (let i in myArray) {
      console.log("Removing ",myArray[i].ID)
      this.deleteCustomFilter(myArray[i].ID,true);
      obsArray.push(this.deleteCustomFilter(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  exportItem(item : any) : void {
    this.exportFileModal.initExportModal(item);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.customFilterService.checkOnDeleteCustomFilter(id)
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

  newCustomFilter() {
    this.viewTestFilterModal.newCustomFilter();
  }

  editCustomFilter(row) {
    this.snmpDeviceService.getDevicesById(row.RelatedDev || null).subscribe(
      data => {
        this.viewTestFilterModal.editCustomFilter(data, row)
      },
      err => {
        this.viewTestFilterModal.newCustomFilter();
        console.log(err);
      },
      () => console.log("DONE")
    )
 	}

  deleteCustomFilter(id,recursive?) {
    if(!recursive){
      this.customFilterService.deleteCustomFilter(id)
        .subscribe(data => { },
        err => console.error(err),
        () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
        );
    } else {
      return this.customFilterService.deleteCustomFilter(id)
      .do(
        (test) =>  { this.counterItems++},
        (err) => { this.counterErrors.push({'ID': id, 'error' : err})}
      );
    }
  }

  cancelEdit() {
    this.editmode = "list";
  }
  saveCustomFilter() {
    if (this.customfilterForm.dirty && this.customfilterForm.valid) {
      this.customFilterService.addCustomFilter(this.customfilterForm.value)
        .subscribe(data => { console.log(data) },
        err => {
          console.log(err);
        },
        () => { this.editmode = "list"; this.reloadData() }
        );
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
