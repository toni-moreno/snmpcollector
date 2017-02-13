import { Component, Input, Output, Pipe, PipeTransform, ViewChild, EventEmitter, OnInit, ChangeDetectionStrategy  } from '@angular/core';
import { Validators, FormGroup, FormArray, FormBuilder } from '@angular/forms';
import { ModalDirective } from 'ng2-bootstrap';
import { SnmpDeviceService } from '../snmpdevice/snmpdevicecfg.service';
import { SnmpMetricService } from '../snmpmetric/snmpmetriccfg.service';
import { InfluxMeasService } from '../influxmeas/influxmeascfg.service';
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
  providers: [SnmpDeviceService, SnmpMetricService, MeasGroupService, InfluxMeasService, MeasFilterService, SpinnerComponent, CustomFilterService],
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
  public validationClick(myId: string): void {
    this.childModal.hide();
    this.validationClicked.emit(myId);
  }

  //CONSTRUCTOR
  constructor(private builder: FormBuilder, public metricMeasService: SnmpMetricService, public influxMeasService: InfluxMeasService, public customFilterService: CustomFilterService, public measGroupService: MeasGroupService, public measFilterService: MeasFilterService, public snmpDeviceService: SnmpDeviceService) {
  }

  //INITI
  ngOnInit() {
    this.filterForm = this.builder.group({
      ID: ['', Validators.required],
      Description: ['', Validators.required],
      RelatedDev: [''],
      Items: this.builder.array([])
    });
  }

  //SHOW MODAL FUNCTION

  show(_formValues) {
    //reset var values
    if (!_formValues) this.formValues = null;
    this.selectedMeas = null;
    this.selectedMeasGroup = null;
    this.measGroups = null;

    this.unsubscribeRequest();
    this.alertHandler = {};
    this.queryResult = null;
    this.isConnected = false;
    this.isRequesting = true;

    if (this.formValues) {
      this.loadFromDevice = true;
      this.formValues = _formValues;
      this.loadDeviceValues(this.formValues);
      this.pingDevice(this.formValues);
      this.childModal.show();
    }
    else {
      this.snmpDeviceService.getDevices(null)
        .subscribe(
        data => {
          this.snmpdevs = [];
          for (let entry of data) {
            this.snmpdevs.push({ 'id': entry.ID, 'name': entry.ID });
          }

          this.childModal.show();
        },
        err => console.error(err),
        () => console.log('DONE')
        );
    }
  }

  //HIDE MODAL FUNCCTION

  hide() {
    this.childModal.hide();
  }

  clearSelected() {
    this.filterForm.controls.Items = this.builder.array([]);
  }

  resetVars() {
    this.checkedResults = [];
    this.queryResult = null;
    this.all = true;
    this.clearSelected();
  }

  //Forms:

  filterForm: any
  newFilterForm: any;
  itemForm: any;

  //Bool controllers
  all: boolean = true;
  showNewFilterForm: boolean = false;
  showItemForm: boolean = false;
  loadFromDevice : boolean = false;


  //Selector Data:

  private mySettings: IMultiSelectSettings = {
    singleSelect: true,
  };
  snmpdevs: IMultiSelectOption[] = [];
  measGroups: IMultiSelectOption[] = [];

  //SELECTED DATA:
  selectedMeas: any;
  selectedMeasGroup: any;
  selectedOID: any;

  customFilterValues: any;

  //Data
  checkedResults: Array<any> = [];
  queryResult: any;


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


  //FILTERFORM
  //ADD
  addFilterValue(selectedTag: string, customAlias: string): void {
    // add address to the list
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
  selectMeasGroup(measGroup: string, forceMode: string): void {
    this.queryResult = null;
    if (this.selectedMeasGroup !== measGroup) {
      this.selectedMeasGroup = measGroup;
      this.selectedOID = null;
      this.getMeasGroupByIdForModal(measGroup);
      this.clearSelected();
      this.selectedMeas = "";
    }
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
    this.resetVars();
    this.snmpDeviceService.getDevicesById(id)
      .subscribe(
      data => {
        this.measGroups = [];
        if (data.MeasurementGroups) {
          for (let entry of data.MeasurementGroups) {
            this.measGroups.push({ 'id': entry, 'name': entry });
          }
          this.selectedMeasGroup = null;
          this.measOnMeasGroup = [];
          this.selectedOID = "";
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
      this.selectedMeas = id;
      this.getMeasByIdforModal(id);
      this.resetVars();
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
        this.checkedResults.push(tagID);
        this.showItemForm = false;
      }
    } else {
      alert("Tag ID: " + tagID + " already exists")
    }
  }

  //QUERYRESULT PANEL
  sendQuery() {
    this.snmpDeviceService.sendQuery(this.formValues, 'walk', this.selectedOID)
      .map(data => { this.queryResult = data })
      .subscribe(data => {
        for (let res of this.queryResult.QueryResult) {
          res.Value = res.Value.toString();
          res.checked = false;
          if (this.checkedResults.indexOf(res.Value) > -1) res.checked = true;
        }

        this.queryResult.OID = this.selectedOID;
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
              this.checkedResults.push(tmpValue)
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


  //BACKGROUND PROVIDERS FOR MODAL

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
    this.myObservable = this.influxMeasService.getMeasById(id)
      .subscribe(
      data => {
        this.selectedOID = data.IndexOID;
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  addCustomFilter() {
    console.log(this.formValues);
    this.filterForm.controls['RelatedDev'].patchValue(this.formValues.id || this.formValues.ID);
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
  }

  saveMeasFilter() {
    if (this.newFilterForm.valid) {
      this.measFilterService.addMeasFilter(this.newFilterForm.value)
        .subscribe(data => {
          this.hide();
        },
        err => console.error(err),
        () => { console.log("DONE") }
        );
    }
  }


  //WAIT
  pingDevice(formValues) {
    this.alertHandler = null;
    this.isRequesting = true;
    this.myObservable = this.snmpDeviceService.pingDevice(formValues)
      .subscribe(data => {
        this.alertHandler = { msg: 'Test succesfull ' + data['SysDescr'], type: 'success', closable: true };
        this.isConnected = true;
        this.isRequesting = false
      },
      err => {
        console.error(err);
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
