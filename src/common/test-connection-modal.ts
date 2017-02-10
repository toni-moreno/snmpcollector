import { Component, Input, Output, Pipe, PipeTransform, ViewChild, EventEmitter, OnInit  } from '@angular/core';
import { Validators, FormGroup, FormArray, FormBuilder } from '@angular/forms';
import { ModalDirective,ModalOptions} from 'ng2-bootstrap';
import { SnmpDeviceService } from '../snmpdevice/snmpdevicecfg.service';
import { SnmpMetricService } from '../snmpmetric/snmpmetriccfg.service';
import { InfluxMeasService } from '../influxmeas/influxmeascfg.service';
import { OidConditionService } from '../oidcondition/oidconditioncfg.service';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from './multiselect-dropdown';


import {SpinnerComponent} from '../common/spinner';
import { Subscription } from "rxjs";


@Component({
    selector: 'test-connection-modal',
    template: `
      <div bsModal #childModal="bs-modal" [config]="modalOptions" keyboard="false" class="modal fade" tabindex="-1" role="dialog" arisa-labelledby="myLargeModalLabel" aria-hidden="true">
          <div class="modal-dialog modal-lg" style="width: 80%">
            <div class="modal-content">
              <div class="modal-header">
                <button type="button" class="close" (click)="hide()" aria-label="Close">
                  <span aria-hidden="true">&times;</span>
                </button>
                <h4 class="modal-title" *ngIf="formValues != null">{{titleName}} <b>{{ formValues.id }}</b></h4>
             </div>
              <div class="modal-body">
              <!--System Info Panel-->
              <div class="panel panel-primary" *ngIf = "maximized === false">
                <div class="panel-heading">System Info</div>
                <my-spinner [isRunning]="isRequesting"></my-spinner>
                <div [ngClass]="['panel-body', 'bg-'+alertHandler.type]">
                  {{alertHandler.msg}}
                </div>
              </div>
              <div class="row" *ngIf="isConnected && maximized === false" >
              <!--System Info Panel-->
                <div class="col-md-6">
                <div class="panel panel-primary">
                  <div class="panel-heading">Source from OID</div>
                  <div class="panel-body">
                  <div class="col-md-2" *ngFor="let selector of selectors; let i=index">
                  <label class="checkbox-inline">
                    <input type="radio" class=""  (click)="selectOption(selector.option,i)" [checked]="selectedOption === selector.option">{{selector.title}}
                  </label>
                  <ss-multiselect-dropdown *ngIf="selectedOption === selector.option && selector.option !== 'Direct'" [options]="selector.Array" [texts]="myTexts" [settings]="mySettings" ngModel (ngModelChange)="selectItem($event,selector.Array,i)"></ss-multiselect-dropdown>
                  </div>
                  </div>
                </div>
                </div>
                <form [formGroup]="testForm" class="form-horizontal">
                  <div class="col-md-6">
                    <div class="panel panel-primary">
                      <div class="panel-heading">Connection data</div>
                        <div class="panel-body">
                        <div class="col-md-6">
                        <div class="panel panel-default">
                          <!-- Default panel contents -->
                          <div class="panel-heading" style="padding: 0px">History</div>
                            <ul class="list-group">
                            <li class="list-group-item"  style="padding: 3px" *ngIf="histArray.length === 0"> Empty history</li>
                            <li class="list-group-item" style="padding: 3px" *ngFor="let hist of histArray">
                            <div>
                            <span style="padding: 0px; margin-right: 10px" class="glyphicon glyphicon-plus" (click)="selectedOID = hist"></span>
                            <span> {{hist}} </span>
                            </div>
                            </li>
                          </ul>
                        </div>
                        </div>
                        <div class="col-md-5">
                        <!--MODE-->
                        <div class="form-group">
                          <label for="Mode" class="col-sm-4 control-label">Mode</label>
                          <div class="col-sm-8">
                          <select class="form-control" formControlName="Mode" id="Mode" [(ngModel)]="setMode">
                            <option *ngFor="let mode of modeGo" >{{mode}}</option>
                          </select>
                          </div>
                        </div>
                        <!--OID-->
                        <div class="form-group">
                          <label for="OID" class="col-sm-4 control-label">OID</label>
                          <div class="col-sm-8">
                          <input type="text" class="form-control" placeholder="Text input" [ngModel]="selectedOID" formControlName="OID" id="OID">
                          </div>
                        </div>
                            <button type="button" class="btn btn-primary pull-right" style="margin-top:10px" [disabled]="!testForm.valid" (click)="sendQuery()">Send query</button>
                        </div>
                        </div>
                      </div>
                    </div>
                  </form>
              </div>
              <div class="row">
              <div class="col-md-12">
              <div *ngIf="queryResult" class="panel panel-default">
                <div class="panel-heading">
                  <h4>
                    Query OID: {{queryResult.OID}}
                    <label [ngClass]="(queryResult.QueryResult[0].Type != 'ERROR' && queryResult.QueryResult[0].Type != 'NoSuchObject' && queryResult.QueryResult[0].Type != 'NoSuchInstance') ? ['label label-primary'] : ['label label-danger']" style="padding-top: 0.5em; margin:0px">
                      {{queryResult.QueryResult[0].Type != 'ERROR' && queryResult.QueryResult[0].Type != 'NoSuchObject' && queryResult.QueryResult[0].Type != 'NoSuchInstance' ? queryResult.QueryResult.length +' results': '0 results - '+queryResult.QueryResult[0].Type}}
                    </label>
                    <i [ngClass]="maximized ? ['pull-right glyphicon glyphicon-resize-small']: ['pull-right glyphicon glyphicon-resize-full']" style="margin-left: 10px;" (click)="maximizeQueryResults()"></i>
                    <span class="pull-right">  elapsed: {{queryResult.TimeTaken}} s </span>

                  </h4>
                </div>
              <div class="panel-body" [ngStyle]="maximized ? {'max-height' : '100%' } : {'max-height.px' : 300 , 'overflow-y' : 'scroll'}">
                <table class="table table-hover table-striped table-condensed" style="width:100%">
                <thead>
                <tr>
                    <th>OID</th>
                    <th>Type</th>
                    <th>Value</th>
                    <th *ngIf="editResults"> Edit </th>
                </tr>
                </thead>
                <tr *ngFor="let entry of queryResult.QueryResult; let i = index">
                  <td>{{entry.Name}} </td>
                  <td> {{entry.Type}}</td>
                  <td>{{entry.Value}}</td>
                  <td *ngIf="editResults">
                    <label class="checkbox-inline">
                      <input type="checkbox" #i>
                    </label>
                  </td>
                </tr>
              </table>
              </div>
              </div>
              </div>
              <div class="modal-footer">
               <button type="button" class="btn btn-primary" (click)="hide()">Close</button>
             </div>
            </div>
          </div>
        </div>`,
        providers: [SnmpDeviceService, SnmpMetricService, InfluxMeasService, OidConditionService, SpinnerComponent],
})

export class TestConnectionModal implements OnInit  {
  @ViewChild('childModal') public childModal: ModalDirective;
  //@Input() formValues : any;
  @Input() titleName : any;
  @Input() systemInfo: any;
  @Output() public validationClicked:EventEmitter<any> = new EventEmitter();

  public validationClick(myId: string):void {
    this.validationClicked.emit(myId);
  }

  constructor(private builder: FormBuilder, public metricMeasService: SnmpMetricService, public influxMeasService: InfluxMeasService, public oidConditionService : OidConditionService,public snmpDeviceService: SnmpDeviceService) {
  }

  ngOnInit () {
    this.testForm = this.builder.group({
    Mode: ['get', Validators.required],
    OID: ['', Validators.required]
    });

  }

  //Modal Options
  modalOptions : ModalOptions  = {
    'keyboard' : false
  }

  //ConnectionForm
  testForm : any;
  setMode : string = 'get';

  //History OIDs
  histArray : Array<string> = [];
  formValues: any;

  //Sysinfo
   alertHandler : any = {};
   isRequesting : boolean ;
   isConnected: boolean;
   myObservable : Subscription;

  //Panel OID source
  selectedOption : any = 'OID';
  selectedOID : any;

  //Selector object:
  public selectors : Object =  [
    { option : 'Direct',  title : 'Direct OID', forceMode : 'get'},
    { option : 'IndexMeas', title : 'Direct Index. Measurements', forceMode : 'walk', Array: []},
    { option : 'IIndexMeas', title : 'Indirect Index. Measurements', forceMode : 'walk', Array: []},
    { option : 'Metric', title : 'Metric Base OID', forceMode : 'get', Array: []},
    { option : 'OIDCond', title : 'OID Conditions', forceMode : 'walk', Array: []}
  ];

  //Panel connection
  modeGo : Array<string> = [
    'get',
    'walk'
  ];

  //Result params
  queryResult : any;
  editResults: boolean = false;
  maximized : boolean = false;


  private mySettings: IMultiSelectSettings = {
      singleSelect: true,
  };

  selectOption(id : string, index : number) {
    this.selectedOption = id;
    this.setMode = this.selectors[index].forceMode;
    switch (id) {
      case 'Metric':
      this.getMetricsforModal(index);
      break;
      case 'IndexMeas':
      this.getMeasforModal('indexed', index);
      break;
      case 'IIndexMeas':
      this.getMeasforModal('indexed_it', index)
      break;
      case 'OIDCond':
      this.getOidConditionforModal(index);
      default: ;
    }
    ;
  }

  selectItem(selectedItem : string, forceMode : boolean, index: number) : void {
      for (let item of this.selectors[index].Array) {
       if (item.id === selectedItem) {
           this.selectedOID = item.OID;
         break;
       }
     }
   }

  maximizeQueryResults () {
    this.maximized = !this.maximized;
  }

  show(_formValues) {
    //reset var values
    this.formValues = _formValues;
    this.alertHandler = {};
    this.queryResult = null;
    this.maximized = false;
    this.isConnected = false;
    this.isRequesting = true;
    this.pingDevice(this.formValues);
    this.childModal.show();
  }

  hide() {
    if (this.myObservable) this.myObservable.unsubscribe();
    this.childModal.hide();
  }

  getMetricsforModal(index : number){
    this.myObservable = this.metricMeasService.getMetrics(null)
    .subscribe(
      data => {
        this.selectors[index].Array = [];
        for (let entry of data) {
          this.selectors[index].Array.push({'id' : entry.ID , 'name': entry.ID, 'OID' : entry.BaseOID});
        }
      },
      err => console.error(err),
      () => console.log('DONE')
    );
  }

  getMeasforModal(type : string, index : number){
    this.myObservable = this.influxMeasService.getMeasByType(type)
    .subscribe(
      data => {
          this.selectors[index].Array = [];
          for (let entry of data) {
            this.selectors[index].Array.push({ 'id': entry.ID, 'name': entry.ID, 'OID': entry.IndexOID});
          }
      },
      err => console.error(err),
      () => console.log('DONE')
    );
  }

  getOidConditionforModal(index : number){
    this.myObservable = this.oidConditionService.getConditions(null)
    .subscribe(
      data => {
        this.selectors[index].Array = [];
        for (let entry of data) {
          this.selectors[index].Array.push({'id' : entry.ID , 'name': entry.ID, 'OID' : entry.OIDCond});
        }
      },
      err => console.error(err),
      () => console.log('DONE')
    );
  }

  sendQuery() {
    this.histArray.push(this.testForm.value.OID);
    if (this.histArray.length > 5 ) this.histArray.shift();
    this.snmpDeviceService.sendQuery(this.formValues,this.testForm.value.Mode, this.testForm.value.OID)
    .subscribe(data => {
      this.queryResult = data;
      this.queryResult.OID = this.testForm.value.OID;
     },
      err => {
      console.error(err);
      },
      () =>  {console.log("DONE")}
     );
  }


  //WAIT
   pingDevice(formValues){
    this.myObservable = this.snmpDeviceService.pingDevice(formValues)
    .subscribe(data => {
      this.alertHandler = {msg: 'Test succesfull '+data['SysDescr'], type: 'success', closable: true};
      this.isConnected = true;
      this.isRequesting = false
     },
      err => {
      console.error(err);
      this.alertHandler = {msg: 'Test failed! '+err['_body'], type: 'danger', closable: true};
      this.isConnected = false;
      this.isRequesting = false
      },
      () =>  {console.log("OK") ;
            }
     );
   }

   ngOnDestroy() {
    if (this.myObservable) this.myObservable.unsubscribe();
   }
}
