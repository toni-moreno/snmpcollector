import { Component, Input, Output, Pipe, PipeTransform, ViewChild, EventEmitter, OnInit  } from '@angular/core';
import { Validators, FormGroup, FormArray, FormBuilder } from '@angular/forms';
import { ModalDirective } from 'ng2-bootstrap/components/modal/modal.component';
import { SnmpDeviceService } from '../snmpdevice/snmpdevicecfg.service';
import { SnmpMetricService } from '../snmpmetric/snmpmetriccfg.service';
import {SpinnerComponent} from '../common/spinner';


@Component({
    selector: 'test-connection-modal',
    template: `
      <div bsModal #childModal="bs-modal" class="modal fade" tabindex="-1" role="dialog" arisa-labelledby="myLargeModalLabel" aria-hidden="true">
          <div class="modal-dialog modal-lg" style="width: 80%">
            <div class="modal-content">
              <div class="modal-header">
                <button type="button" class="close" (click)="childModal.hide()" aria-label="Close">
                  <span aria-hidden="true">&times;</span>
                </button>
                <h4 class="modal-title" *ngIf="formValues != null">{{titleName}} <b>{{ formValues.id }}</b></h4>
             </div>
              <div class="modal-body">
              <!--System Info Panel-->
              <div class="panel panel-default">
                <div class="panel-heading">System Info</div>
                <my-spinner [isRunning]="isRequesting"></my-spinner>
                <div [ngClass]="['panel-body', 'bg-'+alertHandler.type]">
                  {{alertHandler.msg}}
                </div>
              </div>
              <div class="row" *ngIf="isConnected">
              <!--System Info Panel-->
                <div class="col-md-6">
                <div class="panel panel-default">
                  <div class="panel-heading">Source from OID</div>
                  <div class="panel-body">
                    <label class="checkbox-inline col-md-5">
                      <input type="radio" class=""  (click)="selectOption('OID')" [checked]="selectedOption === 'OID'">Direct OID
                    </label>
                    <label class="checkbox-inline col-md-5">
                      <input type="radio" class="btn btn-default" (click)="selectedOption = 'Metric'"   [checked]="selectedOption === 'Metric'">Metric OID
                    </label>
                    <div class="col-md-offset-5">
                    <select class="form-control" *ngIf="selectedOption === 'Metric'" [(ngModel)]="metric" (ngModelChange)="selectMetric($event)" [ngModelOptions]="{standalone: true}">
                      <option *ngFor="let metric of snmpmetrics" [value]="metric.OID" >{{metric.ID}}</option>
                    </select>
                    </div>
                  </div>
                </div>
                </div>
                <form [formGroup]="testForm" class="form-horizontal">
                  <div class="col-md-6">
                    <div class="panel panel-default">
                      <div class="panel-heading">Connection data</div>
                        <div class="panel-body">
                        <!--MODE-->
                        <div class="form-group">
                          <label for="Mode" class="col-sm-4 control-label">Mode</label>
                          <div class="col-sm-8">
                          <select class="form-control" formControlName="Mode" id="Mode">
                            <option *ngFor="let mode of modeGo" [value]="mode" >{{mode | uppercase}}</option>
                          </select>
                          </div>
                        </div>
                        <!--OID-->
                        <div class="form-group">
                          <label for="OID" class="col-sm-4 control-label">OID</label>
                          <div class="col-sm-8">
                          <input type="text" class="form-control" placeholder="Text input" [ngModel]="selectedMetric" formControlName="OID" id="OID">
                          </div>
                        </div>
                            <button type="button" class="btn btn-default pull-right" style="margin-top:10px" [disabled]="!testForm.valid" (click)="sendQuery()">Send query</button>
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
                    Query Result
                    <label class="label label-primary" style="padding-top: 0.5em; margin:0px">
                      {{queryResult.QueryResult.length}}
                    </label>
                    <span class="pull-right">  elapsed: {{queryResult.TimeTaken}} s </span>
                  </h4>
                </div>
              <div class="panel-body" style="max-height : 300px; overflow-y:scroll">
                <table class="table table-hover table-striped table-condensed">
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
               <button type="button" class="btn btn-primary" (click)="childModal.hide()">Close</button>
             </div>
            </div>
          </div>
        </div>`,
        providers: [SnmpDeviceService, SnmpMetricService, SpinnerComponent],
})

export class TestConnectionModal implements OnInit  {
  @ViewChild('childModal') public childModal: ModalDirective;
  @Input() formValues : any;
  @Input() titleName : any;
  @Input() systemInfo: any;
  @Output() public validationClicked:EventEmitter<any> = new EventEmitter();

  //ConnectionForm
  testForm : any;

  public validationClick(myId: string):void {
    this.validationClicked.emit(myId);
  }

  constructor(private builder: FormBuilder, public metricMeasService: SnmpMetricService, public snmpDeviceService: SnmpDeviceService) {
  }

  ngOnInit () {
    this.testForm = this.builder.group({
    Mode: ['get', Validators.required],
    OID: ['', Validators.required]
    });
  this.getMetricsforMeas();
  }

  //Sysinfo
   alertHandler : any = {};
   isRequesting : boolean ;
   isConnected: boolean;


  //Panel OID source
  selectedOption : any = 'OID';
  snmpmetrics : any = [];
  selectedMetric : any;

  //Panel connection
  modeGo : Array<string> = [
    'get',
    'walk'
  ];

  //Result params
  queryResult : any;
  editResults: boolean = false;

  selectOption(id : string) {
    this.selectedOption = id;
    this.selectedMetric = null;
  }

  selectMetric(metric) {
    this.selectedMetric = metric;
    // ... do other stuff here ...
}

  show() {
    //reset var values
    this.alertHandler = {};
    this.queryResult = null;
    this.isConnected = false;
    this.isRequesting = true;
    this.pingDevice();
    this.childModal.show();
  }

  hide() {
    this.childModal.hide();
  }

  getMetricsforMeas(){
    this.metricMeasService.getMetrics(null)
    .subscribe(
      data => {
        for (let entry of data) {
          this.snmpmetrics.push({'ID' : entry.ID , 'OID' : entry.BaseOID});
        }
      },
      err => console.error(err),
      () => console.log('DONE')
    );
  }

  sendQuery() {
    this.snmpDeviceService.sendQuery(this.formValues,this.testForm.value.Mode, this.testForm.value.OID)
    .subscribe(data => {
      this.queryResult = data;
     },
      err => {
      console.error(err);
      },
      () =>  {console.log("DONE")}
     );
  }


  //WAIT
   pingDevice(){
    this.snmpDeviceService.pingDevice(this.formValues)
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
}
