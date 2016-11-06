import { Component, Input, Output, Pipe, PipeTransform, ViewChild, EventEmitter } from '@angular/core';
import { FormGroup,FormControl } from '@angular/forms';
import { ModalDirective } from 'ng2-bootstrap/components/modal/modal.component';


@Component({
    selector: 'test-modal',
    template: `
      <div bsModal #childModal="bs-modal" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="myLargeModalLabel" aria-hidden="true">
          <div class="modal-dialog modal-lg">
            <div class="modal-content">
              <div class="modal-header">
                <button type="button" class="close" (click)="childModal.hide()" aria-label="Close">
                  <span aria-hidden="true">&times;</span>
                </button>
                <h4 class="modal-title" *ngIf="myObject != null">{{titleName}} <b>{{ myObject.ID }}</b></h4>
                <h5 class="modal-title" *ngIf="myObject != null" >
                <span *ngFor="let field of countFields">
                  <span>{{field}}:{{myObject[field] != null ? myObject[field].length : '0' }}
                  </span>
                   </span>
                </h5>
              </div>
              <div class="modal-body">
                <div *ngFor="let entry of myObject | objectParser">
                  <dl class="dl-horizontal" *ngIf="entry.value !='' && entry.value != null ">
                    <dt>{{ entry.key }}</dt>
                    <dd>{{ entry.value }}</dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>
        </div>`
})

export class GenericModal {
  @ViewChild('childModal') public childModal: ModalDirective;
  @Input() countFields : any;
  @Input() titleName : any;


  constructor() { }
  myObject: any = null;
  parseObject( myObject : any){
    this.myObject = myObject;
    this.childModal.show();
  }
}
