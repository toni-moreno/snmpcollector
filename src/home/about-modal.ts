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
                <h4 class="modal-title"><i class="glyphicon glyphicon-info-sign"></i> {{titleName}}</h4>
              </div>
              <div class="modal-body" *ngIf="info">
              <h4 class="text-primary"> <b>SNMPCollector</b> </h4>
              <span> SNMPCollector is a full featured generic SNMP data collector with web administration interface. It is an open Source tool which has as main goal simplify the configuration for getting data from any device with SNMP protocol support and send resulting data to an InfluxDB backend</span>
              <div class="text-right">
                <a href="https://github.com/toni-moreno/snmpcollector" class="text-link"> More info <i class="glyphicon glyphicon-plus-sign"></i></a>
              </div>
              <hr/>
              <h4> Release information </h4>
                <dl class="dl-horizontal">
                  <dt>Instance ID:</dt><dd>{{ info.InstanceID}}</dd>
                  <dt>Version:</dt><dd>{{info.Version}}</dd>
                  <dt>Commit:</dt><dd>{{info.Commit}}</dd>
                  <dt>Build Date:</dt><dd>{{ info.BuildStamp*1000 | date:'yyyy/M/d HH:mm:ss' }}</dd>
                  <dt>License:</dt><dd>MIT License</dd>
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

  public info : RInfo;

  public validationClick(myId: string):void {
    this.validationClicked.emit(myId);
  }

  constructor() { }

  showModal(info : RInfo){
    this.info = info;
    this.childModal.show();
  }

  hide() {
    this.childModal.hide();
  }

}
