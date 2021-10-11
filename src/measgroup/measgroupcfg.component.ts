import { Component, ChangeDetectionStrategy, ViewChild, ChangeDetectorRef } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';
import { MeasGroupService } from './measgroupcfg.service';
import { MeasurementService } from '../measurement/measurementcfg.service';
import { ValidationService } from '../common/validation.service'
import { FormArray, FormGroup, FormControl} from '@angular/forms';
import { ExportServiceCfg } from '../common/dataservice/export.service'
import { Observable } from 'rxjs/Rx';

import { GenericModal } from '../common/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { ItemsPerPageOptions } from '../common/global-constants';

import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

import { TableListComponent } from '../common/table-list.component';
import { MeasGroupCfgComponentConfig, TableRole, OverrideRoleActions } from './measgroupcfg.data';

declare var _:any;

@Component({
  selector: 'measgroups',
  providers: [MeasGroupService, MeasurementService],
  templateUrl: './measgroupeditor.html',
  styleUrls: ['../css/component-styles.css']
})

export class MeasGroupCfgComponent {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

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
  measurement: Array<any>;
  selectmeas: IMultiSelectOption[] = [];
  public defaultConfig : any = MeasGroupCfgComponentConfig;
  public tableRole : any = TableRole;
  public overrideRoleActions: any = OverrideRoleActions;
  myFilterValue: any;
  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];

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


  constructor(public measGroupService: MeasGroupService, public measMeasGroupService: MeasurementService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
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
        this.newMeasGroup()
      case 'view':
        this.viewItem(action.event);
      break;
      case 'edit':
        this.editMeasGroup(action.event);
      break;
      case 'remove':
        this.removeItem(action.event);
      break;
      case 'tableaction':
        this.applyAction(action.event, action.data);
      break;
    }
  }

  viewItem(id) {
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
      return this.measGroupService.deleteMeasGroup(id, true)
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
      return this.measGroupService.editMeasGroup(component, component.ID, true)
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
        this.measurement = data;
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
