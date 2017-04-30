import { Component, Input, Output, Pipe, PipeTransform, ViewChild, EventEmitter } from '@angular/core';
import { FormGroup,FormControl } from '@angular/forms';
import { ModalDirective } from 'ng2-bootstrap';


@Component({
    selector: 'about-modal',
    template: `
      <div bsModal #childModal="bs-modal" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="myLargeModalLabel" aria-hidden="true">
          <div class="modal-dialog">
            <div class="modal-content">
              <div class="modal-header">
                <button type="button" class="close" (click)="childModal.hide()" aria-label="Close">
                  <span aria-hidden="true">&times;</span>
                </button>
                <h4 class="modal-title">{{titleName}}</h4>
              </div>
              <div class="modal-body" *ngIf="info">
              <h4> SNMPCollector </h4>
                <dl class="dl-horizontal">
                  <ng-container *ngFor="let item of info | objectParser">
                    <dt>{{item.key}}</dt>
                    <dd>
                      {{item.value}}
                    </dd>
                  </ng-container>
                </dl>
                <hr>
              <h4> Authors: </h4>
              <dl class="dl-horizontal">
                  <dt>Toni Moreno</dt>
                  <dd>
                    <a href="http://github.com/toni-moreno">GitHub</a>
                  </dd>
                  <dt>Sergio Bengoechea</dt>
                  <dd>
                    <a href="http://github.com/sbengo">GitHub</a>
                  </dd>
              </dl>
              <hr>
              <h4> License: </h4>
              <dl class="dl-horizontal">
                <dt> MIT License </dt>
                  <dd>A short and simple permissive license with conditions only requiring preservation of copyright and license notices. Licensed works, modifications, and larger works may be distributed under different terms and without source code.
                  </dd>
                </dl>
              </div>
              <div class="modal-footer" *ngIf="showValidation === true">
               <button type="button" class="btn btn-primary" (click)="childModal.hide()">Close</button>
             </div>
            </div>
          </div>
        </div>`
})

export class AboutModal {
  @ViewChild('childModal') public childModal: ModalDirective;
  @Input() titleName : any;
  @Input() customMessage: string;
  @Input() showValidation: boolean;
  @Input() textValidation: string;

  @Output() public validationClicked:EventEmitter<any> = new EventEmitter();

  public info : any;

  public validationClick(myId: string):void {
    this.validationClicked.emit(myId);
  }

  constructor() { }

  showModal(info : any){
    this.info = info;
    this.childModal.show();
  }

  hide() {
    this.childModal.hide();
  }

}
