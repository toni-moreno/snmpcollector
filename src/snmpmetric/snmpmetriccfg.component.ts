import { Component, ChangeDetectionStrategy, ViewChild  } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';
import { SnmpMetricService } from './snmpmetriccfg.service';
import { OidConditionService } from '../oidcondition/oidconditioncfg.service';
import { ControlMessagesComponent } from '../common/control-messages.component'
import { ValidationService } from '../common/validation.service'
import { FormArray, FormGroup, FormControl} from '@angular/forms';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { Observable } from 'rxjs/Rx';

import { GenericModal } from '../common/generic-modal';
import { ExportServiceCfg } from '../common/dataservice/export.service'
import { ItemsPerPageOptions } from '../common/global-constants';

import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

declare var _:any;

@Component({
  selector: 'snmpmetrics',
  providers: [SnmpMetricService,OidConditionService],
  templateUrl: './snmpmetriceditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class SnmpMetricCfgComponent {
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
  snmpmetrics: Array<any>;
  filter: string;
  snmpmetForm: any;
  myFilterValue: any;
  //OID selector
  oidconditions: Array<any>;
  selectoidcond: IMultiSelectOption[] = [];
  private mySettings: IMultiSelectSettings = {
       singleSelect: true,
 };

  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public columns: Array<any> = [
    { title: 'ID', name: 'ID' },
    { title: 'FieldName', name: 'FieldName' },
    { title: 'BaseOID', name: 'BaseOID' },
    { title: 'DataSrcType', name: 'DataSrcType' },
    { title: 'ExtraData', name: 'ExtraData'},
    { title: 'GetRate', name: 'GetRate' },
    { title: 'Scale', name: 'Scale' },
    { title: 'Shift', name: 'Shift' },
    { title: 'IsTag', name: 'IsTag' }
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

  constructor(public snmpMetricService: SnmpMetricService, public oidCondService: OidConditionService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  enableEdit() {
    this.editEnabled = !this.editEnabled;
    console.log(this.editEnabled);
    let obsArray = [];
    this.tableAvailableActions = new AvailableTableActions('metric').availableOptions;
  }

  createStaticForm() {
    this.snmpmetForm = this.builder.group({
      ID: [this.snmpmetForm ? this.snmpmetForm.value.ID : '', Validators.required],
      FieldName: [this.snmpmetForm ? this.snmpmetForm.value.FieldName : '', Validators.required],
      DataSrcType: [this.snmpmetForm ? this.snmpmetForm.value.DataSrcType : 'Gauge32', Validators.required],
      IsTag: [this.snmpmetForm ? this.snmpmetForm.value.IsTag : 'false', Validators.required],
      Description: [this.snmpmetForm ? this.snmpmetForm.value.Description : '']
    });
  }


  createDynamicForm(fieldsArray: any) : void {
    //Saves the actual to check later if there are shared values
    let tmpform : any;
    if (this.snmpmetForm)  tmpform = this.snmpmetForm.value;
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
      this.snmpmetForm.addControl(entry.ID, new FormControl(value, entry.Validators));
    }
  }

  setDynamicFields (field : any, override? : boolean) : void  {
    //Saves on the array all values to push into formGroup
    let controlArray : Array<any> = [];

    switch (field) {
      case 'BITS':
      case 'BITSCHK':
        controlArray.push({'ID': 'ExtraData', 'defVal' : '', 'Validators' : Validators.required, 'override' : override });
      case 'OCTETSTRING':
      case 'HWADDR':
      case 'IpAddress':
        controlArray.push({'ID': 'BaseOID', 'defVal' : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required]) })
      break;
      case 'CONDITIONEVAL':
        this.getOidCond();
        controlArray.push({'ID': 'ExtraData', 'defVal' : '', 'Validators' : Validators.required, 'override' : override });
        break;
      case 'STRINGPARSER':
        controlArray.push({'ID': 'BaseOID', 'defVal' : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required]) })
      case 'STRINGEVAL':
        controlArray.push({'ID': 'ExtraData', 'defVal' : '', 'Validators' : Validators.required, 'override' : override });
        controlArray.push({'ID': 'Scale', 'defVal' : '0', 'Validators' : Validators.compose([Validators.required, ValidationService.floatValidator]) })
        controlArray.push({'ID': 'Shift', 'defVal' : '0', 'Validators' : Validators.compose([Validators.required, ValidationService.floatValidator]) })
        break;
      case 'COUNTER32':
      case 'COUNTER64':
      case 'COUNTERXX':
        controlArray.push({'ID': 'GetRate', 'defVal' : 'false', 'Validators' : Validators.required});
      default: //Gauge32
        controlArray.push({'ID': 'BaseOID', 'defVal' : '', 'Validators' :Validators.compose([ValidationService.OIDValidator, Validators.required]) })
        controlArray.push({'ID': 'Scale', 'defVal' : '0', 'Validators' : Validators.compose([Validators.required, ValidationService.floatValidator]) })
        controlArray.push({'ID': 'Shift', 'defVal' : '0', 'Validators' : Validators.compose([Validators.required, ValidationService.floatValidator]) })
        break;
    }
    //Reload the formGroup with new values saved on controlArray
    this.createDynamicForm(controlArray);
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

  reloadData() {
    //In order to avoid issues with the array we clean it every time we refresh data
    this.selectedArray = [];
    this.isRequesting = true;
    // now it's a simple subscription to the observable
    this.snmpMetricService.getMetrics(null)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.snmpmetrics = data;
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
      this.deleteSNMPMetric(myArray[i].ID,true);
      obsArray.push(this.deleteSNMPMetric(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.snmpMetricService.checkOnDeleteMetric(id)
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
  newMetric() {
    if (this.snmpmetForm) {
      this.setDynamicFields(this.snmpmetForm.value.DataSrcType);
    } else {
      this.setDynamicFields(null);
    }
    this.editmode = "create";
  }

  editMetric(row) {
    let id = row.ID;
    this.snmpMetricService.getMetricsById(id)
      .subscribe(data => {
        this.snmpmetForm = {};
        this.snmpmetForm.value = data;
        this.setDynamicFields(row.DataSrcType, false);
        this.oldID = data.ID
        this.editmode = "modify"
       },
      err => console.error(err)
      );
  }
  deleteSNMPMetric(id, recursive?) {
    if (!recursive) {
      this.snmpMetricService.deleteMetric(id)
        .subscribe(data => { },
        err => console.error(err),
        () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
        );
    } else {
      return this.snmpMetricService.deleteMetric(id, true)
      .do(
        (test) =>  { this.counterItems++},
        (err) => { this.counterErrors.push({'ID': id, 'error' : err})}
      );
    }
  }

  cancelEdit() {
    this.editmode = "list";
  }

  saveSnmpMet() {
    console.log(this.snmpmetForm);
    if (this.snmpmetForm.valid) {
      this.snmpMetricService.addMetric(this.snmpmetForm.value)
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
      obsArray.push(this.updateSnmpMet(true,component));
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
        obsArray.push(this.updateSnmpMet(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateSnmpMet(recursive?, component?) {
    if(!recursive) {
      if (this.snmpmetForm.valid) {
        var r = true;
        if (this.snmpmetForm.value.ID != this.oldID) {
          r = confirm("Changing Metric ID from " + this.oldID + " to " + this.snmpmetForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.snmpMetricService.editMetric(this.snmpmetForm.value, this.oldID)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.snmpMetricService.editMetric(component, component.ID, true)
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
