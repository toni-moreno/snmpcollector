import { Component, Input, Output, Pipe, PipeTransform, ViewChild, EventEmitter } from '@angular/core';
import { ModalDirective } from 'ng2-bootstrap';
import { Validators, FormGroup, FormControl, FormArray, FormBuilder } from '@angular/forms';
import { ExportServiceCfg } from './export.service'

@Component({
  selector: 'treeview',
  template: `
      <div>
      <div class="panel panel-default">
        <div class="panel-heading">
        <i role=button [ngClass]="visible === true ? ['glyphicon glyphicon-minus'] : ['glyphicon glyphicon-plus']" (click)="toggleVisible()" style="padding-right: 10px"></i>
        <span>{{title}}</span>
        <label [ngClass]="['label label-'+colorsObject[type]]"> {{type}} </label>
        </div>
        <div class="panel-body" *ngIf="visible === true">
        <pre > {{object | json }} </pre>
        </div>
      </div>
      </div>
      `,
    providers: [ExportServiceCfg]
})

export class TreeView {
  @Input() title : any;
  @Input() type: any;
  @Input() visible : boolean;
  @Input() object : any;

  public colorsObject : Object = {
   "snmpdevicecfg" : 'danger',
   "influxcfg" : 'info',
   "measfiltercfg": 'warning',
   "oidconditioncfg" : 'success',
   "customfiltercfg" : 'default',
   "measurementcfg" : 'primary',
   "snmpmetriccfg" : 'warning',
   "measgroupcfg" : 'success'
 };

    show() {
    console.log("SHOWN");
  }
  constructor(){
    this.visible = false;
  }
  toggleVisible() {
    this.visible = !this.visible;
  }



}
