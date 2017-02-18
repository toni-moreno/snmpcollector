import { Component, ChangeDetectionStrategy, ViewChild  } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { MeasFilterService } from './measfiltercfg.service';
import { InfluxMeasService } from '../influxmeas/influxmeascfg.service';
import { CustomFilterService } from '../customfilter/customfilter.service';
import { OidConditionService } from '../oidcondition/oidconditioncfg.service';
import { FormArray, FormGroup, FormControl} from '@angular/forms';

import { ValidationService } from '../common/validation.service'
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';


import { GenericModal } from '../common/generic-modal';

@Component({
  selector: 'measfilters',
  providers: [MeasFilterService, InfluxMeasService, CustomFilterService,OidConditionService],
  templateUrl: './measfiltereditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class MeasFilterCfgComponent {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;

  editmode: string; //list , create, modify
  measfilters: Array<any>;
  filter: string;
  measfilterForm: any;
  influxmeas: Array<any>;
  selectmeas: IMultiSelectOption[] = [];
  selectCustomFilters:  IMultiSelectOption[] = [];

  oidconditions: Array<any>;
  selectoidcond: IMultiSelectOption[] = [];

  myFilterValue: any;

  private mySettings: IMultiSelectSettings = {
      singleSelect: true,
  };

  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public columns: Array<any> = [
    { title: 'ID', name: 'ID' },
    { title: 'Measurement ID', name: 'IDMeasurementCfg' },
    { title: 'Filter Type', name: 'FType' },
    { title: 'Filter Name', name: 'FilterName' },
    { title: 'EnableAlias', name: 'EnableAlias' }
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

  constructor(public oidCondService: OidConditionService,public customFilterService: CustomFilterService, public measFilterService: MeasFilterService, public measMeasFilterService: InfluxMeasService, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  createStaticForm() {
    this.measfilterForm = this.builder.group({
      ID: [this.measfilterForm ? this.measfilterForm.value.ID : '', Validators.required],
      IDMeasurementCfg: [this.measfilterForm ? this.measfilterForm.value.IDMeasurementCfg : '', Validators.required],
      FType: [this.measfilterForm ? this.measfilterForm.value.FType : 'OIDCondition', Validators.required],
      Description: [this.measfilterForm ? this.measfilterForm.value.Description : '']
    });
  }

  createDynamicForm(fieldsArray: any) : void {
    //Saves the actual to check later if there are shared values
    let tmpform : any;
    if (this.measfilterForm)  tmpform = this.measfilterForm.value;
    this.createStaticForm();

    for (let entry of fieldsArray) {
      let value = entry.defVal;
      //Check if there are common values from the previous selected item
      if (tmpform) {
        if (tmpform[entry.ID] && entry.override !== true) {
          value = tmpform[entry.ID];
        }
      }
      //Set different controls:
      this.measfilterForm.addControl(entry.ID, new FormControl(value, entry.Validators));
    }
  }

  setDynamicFields (field : any, override? : boolean) : void  {
    //Saves on the array all values to push into formGroup
    let controlArray : Array<any> = [];
    switch (field) {
      case 'file':
        controlArray.push({'ID': 'FilterName', 'defVal' : '', 'Validators' : Validators.required, 'override' : override});
        controlArray.push({'ID': 'EnableAlias', 'defVal' : 'true', 'Validators' : Validators.required});

        break;
      case 'CustomFilter':
        this.getCustomFiltersforMeasFilters();
        controlArray.push({'ID': 'FilterName', 'defVal' : '', 'Validators' : Validators.required, 'override' : override});
        controlArray.push({'ID': 'EnableAlias', 'defVal' : 'true', 'Validators' : Validators.required});

      break;
      default: //OID Condition
        this.getOidCond();
        controlArray.push({'ID': 'FilterName', 'defVal' : '', 'Validators' : Validators.required, 'override' : override});
        break;
    }
    //Reload the formGroup with new values saved on controlArray
    this.createDynamicForm(controlArray);
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

  reloadData() {
    // now it's a simple subscription to the observable
    this.measFilterService.getMeasFilter(this.filter)
      .subscribe(
      data => {
        this.measfilters = data;
        this.data = data;
        this.onChangeTable(this.config);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  onFilter() {
    this.reloadData();
  }

  viewItem(id, event) {
    console.log('view', id);
    this.viewModal.parseObject(id);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.measFilterService.checkOnDeleteMeasFilter(id)
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

  newMeasFilter() {
    if (this.measfilterForm) {
      this.setDynamicFields(this.measfilterForm.value.FType);
    } else {
      this.setDynamicFields(null);
    }
    this.getMeasforMeasFilters();
    this.editmode = "create";
  }

  editMeasFilter(row) {
    let id = row.ID;
    this.getMeasforMeasFilters();
    this.measFilterService.getMeasFilterById(id)
      .subscribe(data => {
        this.measfilterForm = {};
        this.measfilterForm.value = data;
        this.setDynamicFields(row.FType, false);
        this.oldID = data.ID
        this.editmode = "modify"
      },
      err => console.error(err),
      );
  }

  deleteMeasFilter(id) {
    this.measFilterService.deleteMeasFilter(id)
      .subscribe(data => { },
      err => console.error(err),
      () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
      );
  }

  cancelEdit() {
    this.editmode = "list";
  }
  saveMeasFilter() {
    if (this.measfilterForm.valid) {
      this.measFilterService.addMeasFilter(this.measfilterForm.value)
        .subscribe(data => { console.log(data) },
        err => console.error(err),
        () => { this.editmode = "list"; this.reloadData() }
        );
    }
  }

  updateMeasFilter() {
    if (this.measfilterForm.valid) {
      var r = true;
      if (this.measfilterForm.value.ID != this.oldID) {
        r = confirm("Changing Measurement Filter ID from " + this.oldID + " to " + this.measfilterForm.value.ID + ". Proceed?");
      }
      if (r == true) {
        this.measFilterService.editMeasFilter(this.measfilterForm.value, this.oldID)
          .subscribe(data => { console.log(data) },
          err => console.error(err),
          () => { this.editmode = "list"; this.reloadData() }
          );
      }
    }
  }

  getMeasforMeasFilters() {
    this.measMeasFilterService.getMeas(null)
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
      () => console.log('DONE')
      );
  }

  getOidCond() {
    this.oidCondService.getConditions(null)
      .subscribe(
      data => {
        this.oidconditions = data;
        this.selectoidcond = [];
        for (let entry of data) {
          console.log(entry)
          this.selectoidcond.push({ 'id': entry.ID, 'name': entry.ID });
        }
      },
      err => console.error(err),
      () => { console.log('DONE') }
      );
  }


  getCustomFiltersforMeasFilters() {
    this.customFilterService.getCustomFilter(null)
      .subscribe(
      data => {
        this.selectCustomFilters = [];
        for (let entry of data) {
          console.log(entry)
          this.selectCustomFilters.push({ 'id': entry.ID, 'name': entry.ID });
        }
       },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

}
