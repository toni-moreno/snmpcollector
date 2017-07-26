import { Component, ChangeDetectionStrategy, ViewChild} from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';
import { InfluxMeasService } from './influxmeascfg.service';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';
import { SnmpMetricService } from '../snmpmetric/snmpmetriccfg.service';
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
  selector: 'influxmeas',
  providers: [InfluxMeasService, SnmpMetricService],
  templateUrl: './influxmeaseditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class InfluxMeasCfgComponent {
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
  influxmeas: Array<any>;
  filter: string;
  influxmeasForm: any;
  testinfluxmeas: any;
  snmpmetrics: Array<any>;
  selectmetrics: IMultiSelectOption[] = [];
  metricArray: Array<Object> = [];
  selectedMetrics: any = [];

  public reportMetricStatus: Array<Object> = [
    { value: 0, name: 'Never Report', icon: 'glyphicon glyphicon-remove-circle', class: 'text-danger' },
    { value: 1, name: 'Report', icon: 'glyphicon glyphicon-ok-circle', class: 'text-success' },
    { value: 2, name: 'Report if not zero', icon: 'glyphicon glyphicon-ban-circle', class: 'text-warning' }
  ];

  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public columns: Array<any> = [
    { title: 'ID', name: 'ID' },
    { title: 'Name', name: 'Name' },
    { title: 'GetMode', name: 'GetMode' },
    { title: 'Index OID', name: 'IndexOID' },
    { title: 'Tag OID', name: 'TagOID' },
    { title: 'Index Tag', name: 'IndexTag' },
    { title: 'Index Tag Format', name: 'IndexTagFormat' },
    { title: 'Index as Value', name: 'IndexAsValue' },
    { title: 'Metric Fields', name: 'Fields' }
  ];

  public tableAvailableActions : any;

  public page: number = 1;
  public itemsPerPage: number = 20;
  public maxSize: number = 5;
  public numPages: number = 1;
  public length: number = 0;
  myFilterValue: any;
  private builder;
  private oldID : string;
  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.columns },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public influxMeasService: InfluxMeasService, public metricMeasService: SnmpMetricService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  enableEdit() {
    this.editEnabled = !this.editEnabled;
    console.log(this.editEnabled);
    let obsArray = [];
    this.tableAvailableActions = new AvailableTableActions('measurement').availableOptions;
  }

  createStaticForm() {
    this.influxmeasForm = this.builder.group({
      ID: [this.influxmeasForm ? this.influxmeasForm.value.ID : '', Validators.required],
      Name: [this.influxmeasForm ? this.influxmeasForm.value.Name : '', Validators.required],
      GetMode: [this.influxmeasForm ? this.influxmeasForm.value.GetMode : 'value', Validators.required],
      Fields: this.builder.array(this.influxmeasForm ? ((this.influxmeasForm.value.Fields) !== null ? this.influxmeasForm.value.Fields : []) : []),
      Description: [this.influxmeasForm ? this.influxmeasForm.value.Description : '']
    });
  }

  createDynamicForm(fieldsArray: any) : void {
    //Saves the actual to check later if there are shared values
    let tmpform : any;
    if (this.influxmeasForm)  tmpform = this.influxmeasForm.value;
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
      this.influxmeasForm.addControl(entry.ID, new FormControl(value, entry.Validators));
    }
  }

  setDynamicFields (field : any, override? : boolean) : void  {
    //Saves on the array all values to push into formGroup
    let controlArray : Array<any> = [];
    console.log(field);
    switch (field) {
      case 'indexed_it':
        controlArray.push({'ID': 'TagOID', 'defVal' : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required])});
      case 'indexed':
        controlArray.push({'ID': 'IndexOID', 'defVal' : '', 'Validators' : Validators.compose([ValidationService.OIDValidator, Validators.required])});
        controlArray.push({'ID': 'IndexTag', 'defVal' : '', 'Validators' : Validators.required});
        controlArray.push({'ID': 'IndexTagFormat', 'defVal' : ''});
        controlArray.push({'ID': 'IndexAsValue', 'defVal' : 'false', 'Validators' : Validators.required});
      break;
      default:
        break;
    }
    //Reload the formGroup with new values saved on controlArray
    this.createDynamicForm(controlArray);
  }


  onChangeMetricArray(id) {

    //Create the array with ID:
    let testMetricID = this.metricArray.map( x => {return x['ID']})
    let delEntries = _.differenceWith(testMetricID,id,_.isEqual);
    let newEntries = _.differenceWith(id,testMetricID,_.isEqual);
    //Remove detected delEntries
    _.remove(this.metricArray, function(n) {
      return delEntries.indexOf(n['ID']) != -1;
    });
    //Add new entries
    for (let a of newEntries) {
      this.metricArray.push ({'ID': a, 'Report': 1});
    }
  }

  onCheckMetric(index: number, reportIndex: number) {
    this.metricArray[index]['Report'] = this.reportMetricStatus[reportIndex]['value'];
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

  reloadData() {
    this.selectedArray = [];
    this.isRequesting = true;
    // now it's a simple subscription to the observable
    this.influxMeasService.getMeas(this.filter)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.influxmeas = data
        this.data = data;
        this.onChangeTable(this.config);
      },
      err => console.error(err),
      () => { console.log('DONE'); }
      );
  }

  onResetFilter() : void {
    this.page = 1;
    this.myFilterValue = "";
    this.config.filtering = {filtering: { filterString: '' }};
    this.onChangeTable(this.config);
  }

  applyAction(test : any) : void {
    //test.devices = [];
    //test.devices = this.selectedArray.map((item) => {return item.ID});
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
      this.deleteInfluxMeas(myArray[i].ID,true);
      obsArray.push(this.deleteInfluxMeas(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.influxMeasService.checkOnDeleteInfluxMeas(id)
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

  newMeas() {
    if (this.influxmeasForm) {
      this.setDynamicFields(this.influxmeasForm.value.GetMode);
    } else {
      this.setDynamicFields(null);
    }
    this.editmode = "create";
    this.getMetricsforMeas();
  }

  editMeas(row) {
    let id = row.ID;
    this.metricArray = [];
    this.selectedMetrics = [];
    this.influxMeasService.getMeasById(id)
      .subscribe(data => {
        this.influxmeasForm = {};
        this.influxmeasForm.value = data;
        if (data.Fields) {
          for (var values of data.Fields) {
            this.metricArray.push({ ID: values.ID, Report: values.Report });
            this.selectedMetrics.push(values.ID);
          }
        }
        this.setDynamicFields(row.GetMode, false);
        this.getMetricsforMeas();
        this.oldID = data.ID
        this.editmode = "modify"
      },
      err => console.error(err)
      );
  }

  deleteInfluxMeas(id, recursive?) {
    if(!recursive) {
      this.influxMeasService.deleteMeas(id)
        .subscribe(data => { },
        err => console.error(err),
        () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
        );
    } else {
      return this.influxMeasService.deleteMeas(id)
      .do(
        (test) =>  { this.counterItems++},
        (err) => { this.counterErrors.push({'ID': id, 'error' : err})}
      );
    }
  }

  cancelEdit() {
    this.editmode = "list";
  }

  saveInfluxMeas() {
    this.influxmeasForm.value['Fields'] = this.metricArray;
    console.log(this.influxmeasForm.value);
    if (this.influxmeasForm.valid) {
      this.influxMeasService.addMeas(this.influxmeasForm.value)
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
      obsArray.push(this.updateInfluxMeas(true,component));
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
        obsArray.push(this.updateInfluxMeas(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateInfluxMeas(recursive?, component?) {
    if (!recursive) {
      if (this.influxmeasForm.valid) {
        var r = true;
        if (this.influxmeasForm.value.ID != this.oldID) {
          r = confirm("Changing Measurement ID from " + this.oldID + " to " + this.influxmeasForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.influxmeasForm.value['Fields'] = this.metricArray;
          this.influxMeasService.editMeas(this.influxmeasForm.value, this.oldID)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.influxMeasService.editMeas(component, component.ID)
      .do(
        (test) =>  { this.counterItems++ },
        (err) => { this.counterErrors.push({'ID': component['ID'], 'error' : err['_body']})}
      )
      .catch((err) => {
        return Observable.of({'ID': component.ID , 'error': err['_body']})
      })
    }
  }

  createMultiselectArray(tempArray) : any {
    let myarray = [];
    for (let entry of tempArray) {
      myarray.push({ 'id': entry.ID, 'name': entry.ID });
    }
    return myarray;
  }

  getMetricsforMeas() {
    this.metricMeasService.getMetrics(null)
      .subscribe(
      data => {
        this.snmpmetrics = data;
        this.selectmetrics = [];
        this.influxmeasForm.controls['Fields'].reset();
        this.selectmetrics = this.createMultiselectArray(data);
        /*for (let entry of data) {
          this.selectmetrics.push({ 'id': entry.ID, 'name': entry.ID });
        }*/
      },
      err => console.error(err),
      () => console.log('DONE')
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
