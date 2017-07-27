import { Component, ChangeDetectionStrategy, ViewChild, ChangeDetectorRef } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';
import { MeasGroupService } from './measgroupcfg.service';
import { InfluxMeasService } from '../influxmeas/influxmeascfg.service';
import { ValidationService } from '../common/validation.service'
import { FormArray, FormGroup, FormControl} from '@angular/forms';
import { ExportServiceCfg } from '../common/dataservice/export.service'
import { Observable } from 'rxjs/Rx';

import { GenericModal } from '../common/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { ItemsPerPageOptions } from '../common/global-constants';

import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

declare var _:any;

@Component({
  selector: 'measgroups',
  providers: [MeasGroupService, InfluxMeasService],
  templateUrl: './measgroupeditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class MeasGroupCfgComponent {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

  editEnabled : boolean = false;
  selectedArray : any = [];
  public isRequesting : boolean;
  public counterItems : number = null;
  public counterErrors: any = [];

  itemsPerPageOptions : any = ItemsPerPageOptions;
  editmode: string; //list , create, modify
  measgroups: Array<any>;
  filter: string;
  measgroupForm: any;
  testmeasgroups: any;
  influxmeas: Array<any>;
  selectmeas: IMultiSelectOption[] = [];

  myFilterValue: any;
  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public columns: Array<any> = [
    { title: 'ID', name: 'ID' },
    { title: 'Measurements', name: 'Measurements' },
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


  constructor(public measGroupService: MeasGroupService, public measMeasGroupService: InfluxMeasService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  enableEdit() {
    this.editEnabled = !this.editEnabled;
    console.log(this.editEnabled);
    let obsArray = [];
    this.tableAvailableActions = new AvailableTableActions('measgroup').availableOptions;
    console.log(this.tableAvailableActions);
  }

  createStaticForm() {
    this.measgroupForm = this.builder.group({
      ID: [this.measgroupForm ? this.measgroupForm.value.ID : '', Validators.required],
      Measurements: [this.measgroupForm ? this.measgroupForm.value.Measurements : null, Validators.compose([Validators.required, ValidationService.emptySelector])],
      Description: [this.measgroupForm ? this.measgroupForm.value.Description : '']
    });
  }

  reloadData() {
    this.selectedArray = [];
    this.isRequesting = true;
    // now it's a simple subscription to the observable
    this.measGroupService.getMeasGroup(null)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.measgroups = data;
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
      this.deleteMeasGroup(myArray[i].ID,true);
      obsArray.push(this.deleteMeasGroup(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.measGroupService.checkOnDeleteMeasGroups(id)
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

  newMeasGroup() {
    this.createStaticForm();
    this.getMeasforMeasGroups();
    this.editmode = "create";
  }

  editMeasGroup(row) {
    let id = row.ID;
    this.getMeasforMeasGroups();
    this.measGroupService.getMeasGroupById(id)
      .subscribe(
      data => {
        this.measgroupForm = {};
        this.measgroupForm.value = data;
        this.oldID = data.ID
        this.createStaticForm();
        this.editmode = "modify"
      },
      err => console.error(err),
      () => console.log("DONE")
      );
  }

  deleteMeasGroup(id, recursive?) {
    if (!recursive) {
    this.measGroupService.deleteMeasGroup(id)
      .subscribe(data => { },
      err => console.error(err),
      () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
      );
    } else {
      return this.measGroupService.deleteMeasGroup(id)
      .do(
        (test) =>  { this.counterItems++},
        (err) => { this.counterErrors.push({'ID': id, 'error' : err})}
      );
    }
  }
  cancelEdit() {
    this.reloadData();
    this.editmode = "list";
  }
  saveMeasGroup() {
    if (this.measgroupForm.valid) {
      this.measGroupService.addMeasGroup(this.measgroupForm.value)
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
      obsArray.push(this.updateMeasGroup(true,component));
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
        obsArray.push(this.updateMeasGroup(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateMeasGroup(recursive?, component?) {
    if (!recursive) {
      if (this.measgroupForm.valid) {
        var r = true;
        if (this.measgroupForm.value.ID != this.oldID) {
          r = confirm("Changing Measurement Group ID from " + this.oldID + " to " + this.measgroupForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.measGroupService.editMeasGroup(this.measgroupForm.value, this.oldID)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.measGroupService.editMeasGroup(component, component.ID)
      .do(
        (test) =>  { this.counterItems++ },
        (err) => { this.counterErrors.push({'ID': component['ID'], 'error' : err['_body']})}
      )
      .catch((err) => {
        return Observable.of({'ID': component.ID , 'error': err['_body']})
      })
    }
  }

  getMeasforMeasGroups() {
    this.measMeasGroupService.getMeas(null)
      .subscribe(
      data => {
        this.influxmeas = data;
        this.selectmeas = [];
        for (let entry of data) {
          console.log(entry)
          this.selectmeas.push({ 'id': entry.ID, 'name': entry.ID });
        }
      },
      err => console.error(err),
      () => { console.log('DONE') }
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
