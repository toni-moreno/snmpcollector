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

import { TableListComponent } from '../common/table-list.component';
import { CustomFilterCfgComponentConfig, TableRole, OverrideRoleActions } from './customfiltercfg.data';

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
  public defaultConfig : any = CustomFilterCfgComponentConfig;
  public tableRole : any = TableRole;
  public overrideRoleActions: any = OverrideRoleActions;

  public page: number = 1;
  public itemsPerPage: number = 20;
  public maxSize: number = 5;
  public numPages: number = 1;
  public length: number = 0;

  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.defaultConfig['table-columns'] },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public customFilterService: CustomFilterService, public snmpDeviceService: SnmpDeviceService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
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
    }
  }

  customActions(action : any) {
    switch (action.option) {
      case 'export' : 
        this.exportItem(action.event);
      break;
      case 'new' :
        this.newCustomFilter()
      case 'view':
        this.viewItem(action.event);
      break;
      case 'edit':
        this.editCustomFilter(action.event);
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
      return this.customFilterService.deleteCustomFilter(id, true)
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
