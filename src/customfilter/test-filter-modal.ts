import { Component, Input, Output, Pipe, PipeTransform, ViewChild, EventEmitter, OnInit, ChangeDetectionStrategy  } from '@angular/core';
import { Validators, FormGroup, FormArray, FormBuilder } from '@angular/forms';
import { ModalDirective, ModalOptions } from 'ngx-bootstrap';
import { SnmpDeviceService } from '../snmpdevice/snmpdevicecfg.service';
import { MeasurementService } from '../measurement/measurementcfg.service';
import { MeasFilterService } from '../measfilter/measfiltercfg.service';
import { MeasGroupService } from '../measgroup/measgroupcfg.service';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';


import { SpinnerComponent } from '../common/spinner';
import { Subscription } from "rxjs";

import { CustomFilterService } from './customfilter.service';


@Component({
  selector: 'test-filter-modal',
  templateUrl: './testingfilter.html',
  styleUrls: ['./filter-modal-styles.css'],
  providers: [SnmpDeviceService, MeasGroupService, MeasurementService, MeasFilterService, SpinnerComponent, CustomFilterService],
})

export class TestFilterModal implements OnInit {
  @ViewChild('childModal') public childModal: ModalDirective;
  @Input() formValues: any;
  @Input() titleName: any;
  @Input() systemInfo: any;
  @Output() public validationClicked: EventEmitter<any> = new EventEmitter();
  @Input() showValidation: boolean;
  @Input() textValidation: string;

  //EMITERS
  public validationClick(myId?: string): void {
    this.childModal.hide();
    this.validationClicked.emit();
  }

  public mode : string;
  public builder : any;

  //CONSTRUCTOR

  constructor(builder: FormBuilder, public measurementService: MeasurementService, public customFilterService: CustomFilterService, public measGroupService: MeasGroupService, public measFilterService: MeasFilterService, public snmpDeviceService: SnmpDeviceService) {
    this.builder = builder;
  }


  //INIT
  ngOnInit() {
  }

  createStaticForm() {
    this.filterForm = this.builder.group({
      ID: [this.filterForm ? this.filterForm.value.ID : '', Validators.required],
      RelatedDev: [this.filterForm ? this.filterForm.value.RelatedDev : ''],
      RelatedMeas: [this.filterForm ? this.filterForm.value.RelatedMeas : ''],
      Items: this.builder.array([]),
      Description: [this.filterForm ? this.filterForm.value.Description : '']
    });
    console.log(this.filterForm);
  }

  clearConfig() {
  //  this.loadFromDevice = true;
    this.formValues = null;
    this.measOnMeasGroup = [];
    this.measGroups = [];
    this.selectedOID = "";
    this.selectedMeas = null;
  }

  clearItems() {
    this.checkedResults = [];
    this.isConnected = false;
  }

  clearQuery() {
    this.alertHandler = {};
    this.queryResult = null;
  }

  //SHOW MODAL FUNCTION: SPLITTED INTO NEW AND EDIT

  newCustomFilter(formValues? : any) {
    //Reset forms!
    this.showNewFilterForm = false;
    this.mode = "new";
    this.clearQuery();
    this.clearItems();
    this.clearConfig();
    this.unsubscribeRequest();
    if (formValues) {
      //Get Related Device INFO:
      this.formValues = formValues;
      this.loadDeviceValues(this.formValues);
      this.pingDevice(this.formValues);
      this.createStaticForm();
    } else{
      this.initGetDevices();
    }
    this.childModal.show();
  }

  editCustomFilter(_formValues : any, editForm? : any, alertMessage? : any) {
    this.mode = "edit";
    //reset var values
    this.clearQuery();
    this.clearItems();
    this.clearConfig();
    //secure forms
    this.showNewFilterForm = false;
    this.formValues = _formValues || null;
    this.unsubscribeRequest();

    //Get Related Device INFO:
    this.loadDeviceValues(this.formValues);
    this.pingDevice(this.formValues);
    this.filterForm = {};
    this.filterForm.value = editForm;
    this.createStaticForm();
    //Fill with items:
    this.fillFilterValues(editForm.Items);
    this.oldID = editForm.ID;
    this.selectedMeas = editForm.RelatedMeas;
    this.getMeasByIdforModal(this.selectedMeas);
    this.childModal.show();
  }

  fillFilterValues(items) {
    for (let item of items) {
     this.addFilterValue(item.TagID, item.Alias);
   }
  }


  clearSelected() {
    this.filterForm.controls.Items = this.builder.array([]);
  }

  //Forms:

  filterForm: any
  newFilterForm: any;
  itemForm: any;

  //Bool controllers
  all: boolean = true;
  showNewFilterForm: boolean = false;
  showItemForm: boolean = false;


  //Selector Data:

  private mySettings: IMultiSelectSettings = {
    singleSelect: true,
  };

  snmpdevs: IMultiSelectOption[] = [];
  measGroups: IMultiSelectOption[] = [];

  //SELECTED DATA:
  selectedMeas: any;
//  selectedMeasGroup: any;
  selectedOID: any;
  customFilterValues: any;

  //Data
  checkedResults: Array<any> = [];
  queryResult: any;
  dataArray : any = [];
  filter : any = null;
  //snmpdevs: any;
  oidValue: any;
  filterText: string = "";
  //Measurements on selected MeasGroup
  measOnMeasGroup: Array<string> = [];
  test: any;
  //Sysinfo
  alertHandler: any = {};
  isRequesting: boolean;
  isConnected: boolean;
  myObservable: Subscription;
  oldID : string;

  public editForm : any;

  //FILTERFORM
  //ADD
  addFilterValue(selectedTag: string, customAlias: string): void {
    // add address to the list
    this.checkedResults.push(selectedTag);
    const control = <FormArray>this.filterForm.controls['Items'];
    control.push(this.builder.group({
      TagID: [selectedTag.toString(), Validators.required],
      Alias: [customAlias || '']
    })
    );
  }

  //REMOVE
  removeFilterValue(i: number) {
    // remove address from the list
    const control = <FormArray>this.filterForm.controls['Items'];
    control.removeAt(i);
  }

  //SELECTORS
  //MeasGroup Selector
  selectMeasGroup(measGroup: string): void {
    this.queryResult = null;
    this.selectedOID = null;
    this.getMeasGroupByIdForModal(measGroup);
    this.selectedMeas = "";
  }

  loadDeviceValues (formValues) : void {
    this.measGroups = [];
    if (formValues.MeasurementGroups) {
      for (let entry of formValues.MeasurementGroups) {
        this.measGroups.push({ 'id': entry, 'name': entry });
      }
      this.measOnMeasGroup = [];
      this.selectedOID = "";
    }
  }

  //Device Selector
  selectDevice(id: string): void {
    this.unsubscribeRequest();
  //  this.resetVars();
    this.isConnected = false;
    this.snmpDeviceService.getDevicesById(id)
      .subscribe(
      data => {
        this.clearConfig();
        if (data.MeasurementGroups) {
          for (let entry of data.MeasurementGroups) {
            this.measGroups.push({ 'id': entry, 'name': entry });
          }
        //clearItems
        this.clearQuery()
        }
        this.formValues = data;
        this.pingDevice(this.formValues);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  //Select Measurement -> Load OID
  selectMeasOID(id: string) {
      if (this.selectedMeas !== id) {
        this.clearQuery();
        this.selectedMeas = id;
        this.getMeasByIdforModal(id);
      }
  }

  //Select ALL Options
  selectAllOptions(checkAll: boolean): void {
    for (let entry of this.queryResult.QueryResult) {
      if (checkAll) {
        if (this.checkedResults.indexOf(entry.Value) == -1)
          this.selectOption(entry.Value);
      } else {
        let test = this.checkedResults.indexOf(entry.Value);
        if (test > -1) {
          this.removeFilterValue(test);
          this.checkedResults.splice(test, 1);
        }
        entry.checked = false;
      }
    }
    if (this.queryResult.QueryResult.length === this.checkedResults.length) {
      this.all = false;
    } else {
      this.all = !this.all
    }
  }

  //SHOW ITEMS FORM

  showItemFormPanel(): void {
    this.showItemForm = true;
    this.itemForm = this.builder.group({
      TagID: ['', Validators.required],
      Alias: ['']
    });
  }

  showNewFilterFormPanel(): void {
    this.showNewFilterForm = true;
    this.newFilterForm = this.builder.group({
      id: ['', Validators.required],
      IDMeasurementCfg: [''],
      FType: ['CustomFilter'],
      FilterName: [''],
      EnableAlias: ['false'],
      Description: ['']
    });
  }

  addCustomItem(tagID: string): void {
    //Check if it already  exist
    if ((this.checkedResults && this.checkedResults.indexOf(tagID) > -1) == false) {
      if (this.selectOption(this.itemForm.controls['TagID'].value) == false) {
        this.addFilterValue(this.itemForm.controls['TagID'].value, this.itemForm.controls['Alias'].value);
        this.showItemForm = false;
      }
    } else {
      alert("Tag ID: " + tagID + " already exists")
    }
  }

  onChange(event){
    let tmpArray = this.dataArray.filter((item: any) => {
      return item['Value'].match(event);
    });
    this.queryResult.QueryResult = tmpArray;
  }

  //QUERYRESULT PANEL
  sendQuery() {
    this.filter = null;
    this.isRequesting = true;
    this.snmpDeviceService.sendQuery(this.formValues, 'walk', this.selectedOID, true)
      .subscribe(data => {
        this.queryResult = data;
        for (let res of this.queryResult.QueryResult) {
          res.Value = res.Value.toString();
          res.checked = false;
          let a = this.checkedResults.indexOf(res.Value)
          if (a > -1){
            res.checked = true;
          }
        }
        this.queryResult.OID = this.selectedOID;
        this.isRequesting = false;
        this.dataArray = this.queryResult.QueryResult;

      },
      err => {
        console.error(err);
      },
      () => { console.log("DONE") }
      );
  }

  removeOption(id: any): void {
    if (this.queryResult) {
      this.selectOption(id);
    } else {
      if (this.checkedResults.indexOf(id) != -1) {
        let index = this.checkedResults.indexOf(id);
        this.checkedResults.splice(index, 1);
        this.removeFilterValue(index);
      }
    }
  }

  selectOption(id: any): boolean {
    //Check if there is some OID
    let ifexist: boolean = false;
    let tmpValue = id;
    //Look for every queryResult value:

    if (this.queryResult) {
      for (let entry of this.queryResult.QueryResult) {
        //if matches, changes the checked status
        if (entry.Value === tmpValue) {
          //Change the check status
          entry.checked = !entry.checked;
          //When selected:
          if (entry.checked === true) {
            //Push the first ocrurante to avoid duplicating entries with the same value
            if (this.checkedResults.indexOf(tmpValue) == -1) {
              ifexist = true;
              this.addFilterValue(tmpValue, null);
            }
            //When non selected:
          } else {
            //Delete at first occurance
            if (this.checkedResults.indexOf(tmpValue) != -1) {
              let index = this.checkedResults.indexOf(tmpValue);
              this.checkedResults.splice(index, 1);
              this.removeFilterValue(index);
            }
          }
        }
        //Must check if the value is custom value => It is not on the queryResult array and its added on checkedResults array
        if (ifexist == false && (this.checkedResults.indexOf(tmpValue) != -1)) {
          let index = this.checkedResults.indexOf(tmpValue);
          this.checkedResults.splice(index, 1);
          this.removeFilterValue(index);
        }
      }
      //Change titles if all are un/selected
      if (this.queryResult.QueryResult.length === this.checkedResults.length) {
        this.all = false;
      } else if (this.checkedResults.length == 0) {
        this.all = !this.all;
      } else {
        this.all != this.all;
      }
    }

    return ifexist;
  }


  //PROVIDERS FOR MODAL
  initGetDevices() {
    this.isRequesting = false;
    this.alertHandler = {};
    this.unsubscribeRequest();
    this.clearItems();
    this.snmpDeviceService.getDevices(null)
      .subscribe(
      data => {
        this.snmpdevs = [];
        for (let entry of data) {
          this.snmpdevs.push({ 'id': entry.ID, 'name': entry.ID });
        }
        this.createStaticForm();
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  getMeasGroupByIdForModal(id: string) {
    this.myObservable = this.measGroupService.getMeasGroupById(id)
      .subscribe(
      data => {
        this.measOnMeasGroup = data.Measurements;
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }


  getMeasByIdforModal(id: string) {
    this.myObservable = this.measurementService.getMeasById(id)
      .subscribe(
      data => {
        this.selectedOID = data.IndexOID || null;
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  addCustomFilter() {
    console.log(this.formValues);
    this.filterForm.controls['RelatedDev'].patchValue(this.formValues.ID);
    this.filterForm.controls['RelatedMeas'].patchValue(this.selectedMeas);
    if (this.mode === "new") {
    this.customFilterService.addCustomFilter(this.filterForm.value)
      .subscribe(data => {
        this.showNewFilterForm = true;
        this.showNewFilterFormPanel();
        this.customFilterValues = data;
        this.alertHandler = {
          msg: 'Your custom filter' + data['ID'] + ' was added succesfully. You can assign it to a new Measurement Filter', type: 'success', closable: true
        };
      },
      err => {
        console.error(err),
          this.alertHandler = { msg: 'Something went wrong... ' + err['_body'], type: 'danger', closable: true };
      },
      () => { console.log("DONE"); }
      );
    } else {
      if (this.filterForm.valid) {
        var r = true;
        if (this.filterForm.value.ID != this.oldID) {
          r = confirm("Changing CustomFilter ID from " + this.oldID + " to " + this.filterForm.value.ID + ". Proceed?");
        }
        if (r == true) {
            this.customFilterService.editCustomFilter(this.filterForm.value, this.oldID)
            .subscribe(data => {
              this.validationClick();
            },
            err => console.error(err),
            () => { }
            );
        }
      }
    }
  }

  saveMeasFilter() {
    if (this.newFilterForm.valid) {
      this.measFilterService.addMeasFilter(this.newFilterForm.value)
      .subscribe(data => {
        this.validationClick();
      },
      err => console.error(err),
      () => { console.log("DONE")}
      );
    }
  }

  pingDevice(formValues) {
    this.alertHandler = {};
    this.isRequesting = true;
    this.myObservable = this.snmpDeviceService.pingDevice(formValues,true)
      .subscribe(data => {
        this.alertHandler = { msg: 'Test succesfull ' + data['SysDescr'], type: 'success', closable: true };
        this.isConnected = true;
        this.isRequesting = false
      },
      err => {
        this.alertHandler = { msg: 'Test failed! ' + err['_body'], type: 'danger', closable: true };
        this.isConnected = false;
        this.isRequesting = false
      },
      () => {
        console.log("OK");
      }
      );
  }

  unsubscribeRequest() {
    if (this.myObservable) this.myObservable.unsubscribe();
  }


  ngOnDestroy() {
    console.log("UNSUBSCRIBING");
    if (this.myObservable) this.myObservable.unsubscribe();
  }
}
