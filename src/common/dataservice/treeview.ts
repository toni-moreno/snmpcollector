import { Component, Input, Output, Pipe, PipeTransform, ViewChild, EventEmitter } from '@angular/core';
import { ModalDirective } from 'ng2-bootstrap';
import { Validators, FormGroup, FormControl, FormArray, FormBuilder } from '@angular/forms';
import { ExportServiceCfg } from './export.service'

@Component({
  selector: 'treeview',
  template: `
      <div>
      <div class="panel panel-default" style="margin-bottom:0px">
        <div class="panel-heading col-md-12">
          <div style="display: inline" class="text-left col-md-8">
          <i role=button *ngIf="visibleToogleEnable" [ngClass]="visible === true ? ['glyphicon glyphicon-minus'] : ['glyphicon glyphicon-plus']" (click)="toggleVisible()" style="padding-right: 10px"></i>
          <span>{{title}}</span>
          <label *ngIf="keyProperty && keyProperty != 'ID'" class="label label-default">{{keyProperty}} : {{object[keyProperty] ? object[keyProperty] : "--"}} </label>
          <label *ngIf="type && showType === true" [ngClass]="['label label-'+colorsObject[type]]"> {{type}} </label>
          </div>
          <div *ngIf="error" style="display: inline; width: 100%" class="text-left col-md-8 ">
          <h4 [ngClass]="error ? ['text-danger'] : ['text-success']">
          <i [ngClass]="error ? ['glyphicon glyphicon-warning-sign'] : ['glyphicon glyphicon-ok-sign']" tooltipAnimation="true" tooltip="{{error}}"></i>
          {{error | json }}
          </h4>
          </div>
          <div style="display: inline" class="text-right col-md-4">
            <i *ngIf="addClickEnable===true" role="button" [ngClass]="alreadySelected === false ? ['text-success glyphicon glyphicon-ok-circle'] : ['glyphicon glyphicon-arrow-right']" (click)="addItem(title, type)" style="float:right"></i>
            <span *ngIf="recursiveToogle === true">
              <input type="checkbox" [checked]="this.recursive" (click)="recursiveClick()"> Recursive
            </span>
          </div>
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
  @Input() keyProperty : any;
  @Input() visible : boolean;
  @Input() object : any;
  @Input() error: string;
  @Input() index : any;
  @Input() showType: boolean = true;
  @Input() addClickEnable: boolean = false;
  @Input() visibleToogleEnable: boolean = false;
  @Input() recursiveToogle : boolean;
  @Input() alreadySelected: boolean;
  @Output() public addClicked: EventEmitter<any> = new EventEmitter();
  @Output() public recursiveClicked: EventEmitter<any> = new EventEmitter();


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
 recursive : boolean;

    show() {
    console.log("SHOWN");
  }
  constructor(){
    this.showType
    this.visible = false;
    this.recursive = false;
  }
  toggleVisible() {
    this.visible = !this.visible;
  }


  public addItem(ObjectID: any, ObjectTypeID: any): void {
    this.addClicked.emit({ ObjectID, ObjectTypeID, "Options" : {'Recursive': false }});
  }

  public recursiveClick(): void {
    this.recursive = !this.recursive
    this.recursiveClicked.emit({"Recursive":this.recursive, "Index":this.index});
  }

}
