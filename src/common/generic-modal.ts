import { Component, Input, Output, Pipe, PipeTransform, ViewChild, EventEmitter } from '@angular/core';
import { FormGroup,FormControl } from '@angular/forms';
import { ModalDirective } from 'ngx-bootstrap';


@Component({
    selector: 'test-modal',
    template: `
      <div bsModal #childModal="bs-modal" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="myLargeModalLabel" aria-hidden="true">
          <div class="modal-dialog">
            <div class="modal-content">
              <div class="modal-header">
                <button type="button" class="close" (click)="childModal.hide()" aria-label="Close">
                  <span aria-hidden="true">&times;</span>
                </button>
                <h4 class="modal-title" *ngIf="myObject != null">{{titleName}} <b>{{ myObject.ID }}</b></h4>
              </div>
              <div class="modal-body">
                <div *ngIf = "myObject">
                  <div [ngClass]="empty ? customMessageClass[0] : (customMessageClass[1] || customMessageClass[0])" role="alert" *ngIf="customMessage">
                    <span class="glyphicon glyphicon-exclamation-sign"></span>
                    {{empty ? customMessage[0] : (customMessage[1] || customMessage[0])}}
                  </div>
                </div>
                <div  *ngFor="let entry of myObject | objectParser">
                  <dl class="dl-horizontal" *ngIf="entry.value !='' && entry.value != null">
                    <dt>{{entry.key}} <label *ngIf="isArray(entry.value)" class="label label-primary" style="display: inline-table; margin:0px">{{ entry.value.length}}</label></dt>
                    <dd *ngIf="!isArray(entry.value)">
                      {{entry.value}}
                    </dd>
                    <div *ngIf="isArray(entry.value)" style="margin-bottom:10px">
                      <div *ngFor="let val of entry.value; let i = index">
                        <div *ngIf="isObject(val)">
                          <dd *ngIf="val.Report != null" [ngClass]="reportMetricStatus[val.Report]?.class">{{val?.ID}} <i [ngClass]="reportMetricStatus[val.Report]?.icon"></i>
                          <dd *ngIf="val.Report == null">{{val.TagID}} - {{val.Alias}}</dd>
                        </div>
                        <div *ngIf="!isObject(val)">
                          <dd> {{val}}</dd>
                        </div>
                      </div>
                    </div>
                    <dt *ngIf = "entry.value.Description">Description</dt>
                    <dd *ngIf = "entry.value.Description">{{entry.value.Description}}</dd>
                    <hr>
                  </dl>
                </div>
              </div>
              <div class="modal-footer" *ngIf="showValidation === true">
               <button type="button" class="btn btn-default" (click)="childModal.hide()">Close</button>
               <button type="button" class="btn btn-primary" (click)="validationClick(myObject.ID)">{{textValidation ? textValidation : Save}}</button>
             </div>
            </div>
          </div>
        </div>`
})

export class GenericModal {
  @ViewChild('childModal') public childModal: ModalDirective;
  @Input() titleName : any;
  @Input() customMessage: string;
  @Input() showValidation: boolean;
  @Input() textValidation: string;
  @Input() customMessageClass: string;
  @Input() controlSize: boolean;

  @Output() public validationClicked:EventEmitter<any> = new EventEmitter();

  public validationClick(myId: string):void {
    this.validationClicked.emit(myId);
  }

  constructor() { }
  myObject: any = null;
  empty : any = false;

  public reportMetricStatus: Array<Object> = [
    { value: 0, name: 'Never Report', icon: 'glyphicon glyphicon-remove-circle', class: 'text-danger' },
    { value: 1, name: 'Report', icon: 'glyphicon glyphicon-ok-circle', class: 'text-success' },
    { value: 2, name: 'Report if not zero', icon: 'glyphicon glyphicon-ban-circle', class: 'text-warning' }
  ];

  parseObject( myObject : any){
    this.myObject = myObject;
    this.empty = (Object.keys(myObject).length > 1);
    this.childModal.show();
  }

  isArray (myObject) {
    return myObject instanceof Array;
  }

  isObject (myObject) {
  return typeof myObject === 'object'
}


  hide() {
    this.childModal.hide();
  }

}
