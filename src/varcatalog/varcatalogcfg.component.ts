import { Component, ChangeDetectionStrategy, ViewChild } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { FormArray, FormGroup, FormControl} from '@angular/forms';

import { VarCatalogService } from './varcatalogcfg.service';
import { ValidationService } from '../common/validation.service'
import { ExportServiceCfg } from '../common/dataservice/export.service'

import { GenericModal } from '../common/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { Observable } from 'rxjs/Rx';

import { ItemsPerPageOptions } from '../common/global-constants';
import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

declare var _:any;

@Component({
  selector: 'varcatalog',
  providers: [VarCatalogService, ValidationService],
  templateUrl: './varcatalogeditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class VarCatalogCfgComponent {

  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

  itemsPerPageOptions : any = ItemsPerPageOptions;
  editmode: string; //list , create, modify
  varcatalogs: Array<any>;
  filter: string;
  varcatalogForm: any;
  myFilterValue: any;
  alertHandler : any = null;


  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public columns: Array<any> = [
    { title: 'ID', name: 'ID' },
    { title: 'type', name: 'Type' },
    { title: 'Default Value', name: 'Value' }
  ];

  public tableAvailableActions : any;

  editEnabled : boolean = false;
  selectedArray : any = [];

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
    sorting: { columns: this.columns },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public varCatalogService: VarCatalogService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  enableEdit() {
    this.editEnabled = !this.editEnabled;
    let obsArray = [];
    this.tableAvailableActions = new AvailableTableActions('varcatalog').availableOptions;
  }

  createStaticForm() {
    this.varcatalogForm = this.builder.group({
      ID: [this.varcatalogForm ? this.varcatalogForm.value.ID : '', Validators.required],
      Type: [this.varcatalogForm ? this.varcatalogForm.value.Type : '', Validators.required],
      Value: [this.varcatalogForm ? this.varcatalogForm.value.Value : ''],
      Description: [this.varcatalogForm ? this.varcatalogForm.value.Description : '']
    });
  }

  reloadData() {
    // now it's a simple subscription to the observable
    this.alertHandler = null;
    this.varCatalogService.getVarCatalog(null)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.varcatalogs = data
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

  exportItem(item : any) : void {
    this.exportFileModal.initExportModal(item);
  }

  removeAllSelectedItems(myArray) {
    let obsArray = [];
    this.counterItems = 0;
    this.isRequesting = true;
    for (let i in myArray) {
      console.log("Removing ",myArray[i].ID)
      this.deleteVarCatalog(myArray[i].ID,true);
      obsArray.push(this.deleteVarCatalog(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.varCatalogService.checkOnDeleteVarCatalog(id)
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
  newVarCatalog() {
    //No hidden fields, so create fixed Form
    this.createStaticForm();
    this.editmode = "create";
  }

  editVarCatalog(row) {
    let id = row.ID;
    this.varCatalogService.getVarCatalogById(id)
      .subscribe(data => {
        this.varcatalogForm = {};
        this.varcatalogForm.value = data;
        this.oldID = data.ID
        this.createStaticForm();
        this.editmode = "modify";
      },
      err => console.error(err)
      );
 	}

  deleteVarCatalog(id, recursive?) {
    if (!recursive) {
    this.varCatalogService.deleteVarCatalog(id)
      .subscribe(data => { },
      err => console.error(err),
      () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
      );
    } else {
      return this.varCatalogService.deleteVarCatalog(id, true)
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

  saveVarCatalog() {
    if (this.varcatalogForm.valid) {
      this.varCatalogService.addVarCatalog(this.varcatalogForm.value)
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
      obsArray.push(this.updateVarCatalog(true,component));
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
        obsArray.push(this.updateVarCatalog(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateVarCatalog(recursive?, component?) {
    if(!recursive) {
      if (this.varcatalogForm.valid) {
        var r = true;
        if (this.varcatalogForm.value.ID != this.oldID) {
          r = confirm("Changing Influx Server ID from " + this.oldID + " to " + this.varcatalogForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.varCatalogService.editVarCatalog(this.varcatalogForm.value, this.oldID, true)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.varCatalogService.editVarCatalog(component, component.ID)
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
