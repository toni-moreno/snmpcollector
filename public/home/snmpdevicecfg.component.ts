import { Component } from 'angular2/core';
import { CORE_DIRECTIVES } from 'angular2/common';
import { SnmpDeviceService } from './snmpdevicecfg.service';
import { SnmpDeviceCfg } from './snmpdevicecfg';
import { ACCORDION_DIRECTIVES } from 'ng2-bootstrap';

@Component({
  selector: 'snmpdevs',
  providers: [SnmpDeviceService],
  template: `
<div class="row-fluid">
      <div class=".col-md-3">
	<h2>SNMP devices</h2>
      </div>
      <div  class=".col-md-3">
<input type="text"  class="form-control" required   [(ngModel)]="filter" >
 <button type="submit" (click)="onFilter()" class="btn btn-primary">Filter</button>
 <button type="button" (click)="newDevice()" class="btn btn-primary">New</button>



     </div>
</div>
<accordion [closeOthers]="oneAtATime">
  <accordion-group #group  *ngFor="#snmpdev of snmpdevs" [heading]="snmpdev.id">
   <div accordion-heading>
    <i style="top: -15px" class="pull-right glyphicon"
         [ngClass]="{'glyphicon-chevron-down': group?.isOpen, 'glyphicon-chevron-right': !group?.isOpen}"></i>
   </div>

    <i class="pull-right glyphicon glyphicon glyphicon-eye-open" (click)="viewItem(group,$event)"></i>
    <i class="pull-right glyphicon glyphicon glyphicon glyphicon-remove"  (click)="removeItem(group,$event)"></i>

     {{ snmpdev.Host }}
  </accordion-group>
</accordion>
  `,
  bindings: [SnmpDeviceService],
  directives: [ACCORDION_DIRECTIVES,CORE_DIRECTIVES]
})

export class SnmpDeviceCfgComponent {
  snmpdevs: Array<any>;
  filter: string;

  reloadData(){
  // now it's a simple subscription to the observable
    this.snmpDeviceService.getDevices(this.filter)
      .subscribe(data => { this.snmpdevs = data },
		  err => console.error(err),
		  () => console.log('SNMPDEV DONE')
		      );

  }
  constructor(public snmpDeviceService: SnmpDeviceService) {
	  this.reloadData();
    }
  onFilter(){
	this.reloadData();
  }
 
 viewItem(id,event){
	console.log('view',id);
 }
 removeItem(id,event){
	console.log('remove',id);
 }

}
