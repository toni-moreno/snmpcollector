import { Component } from '@angular/core';
import {  FormBuilder,  Validators} from '@angular/forms';
import { SnmpDeviceService } from './snmpdevicecfg.service';
import { InfluxServerService } from './influxservercfg.service';
import { MeasGroupService } from './measgroupcfg.service';
import { MeasFilterService } from './measfiltercfg.service';

@Component({
  selector: 'snmpdevs',
  providers: [SnmpDeviceService, InfluxServerService, MeasGroupService, MeasFilterService],
  templateUrl: 'public/home/snmpdeviceeditor.html',
  styleUrls:['public/home/snmpdeviceeditor.css'],
})

export class SnmpDeviceCfgComponent {
  public oneAtATime:boolean = true;
  editmode: string; //list , create, modify
  snmpdevs: Array<any>;
  filter: string;
  snmpdevForm: any;
	testsnmpdev: any;
	influxservers: Array<any>;
  measfilters: Array<any>;
  measgroups: Array<any>;

  reloadData(){
  // now it's a simple subscription to the observable
    this.snmpDeviceService.getDevices(this.filter)
      .subscribe(
				data => { this.snmpdevs = data },
		  	err => console.error(err),
		  	() => console.log('DONE')
		  );

  }
  constructor(public snmpDeviceService: SnmpDeviceService, public influxserverDeviceService: InfluxServerService, public measgroupsDeviceService: MeasGroupService, public measfiltersDeviceService: MeasFilterService, builder: FormBuilder) {
	  this.editmode='list';
	  this.reloadData();
	  this.snmpdevForm = builder.group({
		id: ['',Validators.compose([Validators.required, Validators.minLength(4)])],
      		Host: ['', Validators.required],
		Port: [161,Validators.required],
		Retries: [],
		Timeout: [],
		SnmpVersion:['2c',Validators.required],
		Community: ['public'],
		V3SecLevel:[''],
		V3AuthUser:[''],
		V3AuthPass:[''],
		V3AuthProt:[''],
		V3PrivPass:[''],
		V3PrivProt:[''],
		Freq:[60,Validators.required],
		OutDB: ['',Validators.required],
		Config:[''],
		LogLevel:['info',Validators.required],
		LogFile:[''],
		SnmpDebug:['false',Validators.required],
		DeviceTagMame: ['tagname',Validators.required],
		DeviceTagValue: [''],
		Extratags:[''],
		MetricGroups: [''],
		MeasFilters: ['']
	});
    }
  onFilter(){
	this.reloadData();
  }

 viewItem(id,event){
	console.log('view',id);
 }
 removeItem(id){
	console.log('remove',id);
	var r = confirm("Deleting SNMPDEVICE: "+id+". Proceed?");
 	if (r == true) {
		this.snmpDeviceService.deleteDevice(id)
		 .subscribe(
			data => { console.log(data) },
			err => console.error(err),
			() => {this.editmode = "list"; this.reloadData()}
			);

 	}
 }
 newDevice(){
	 this.editmode = "create";
	 this.getInfluxServersforDevices();
	 this.getMeasGroupsforDevices();
	 this.getMeasFiltersforDevices();

 }
 editDevice(id){
	 this.getInfluxServersforDevices();
	 this.getMeasGroupsforDevices();
	 this.getMeasFiltersforDevices();

	 this.snmpDeviceService.getDevicesById(id)
 		 .subscribe(data => { this.testsnmpdev = data },
 		 err => console.error(err),
 		 () =>  this.editmode = "modify"
 				 );
 }
 cancelEdit(){
	 this.editmode = "list";
 }
 saveSnmpDev(){
	 if(this.snmpdevForm.dirty && this.snmpdevForm.valid) {
		 this.snmpDeviceService.addDevice(this.snmpdevForm.value)
		 .subscribe(data => { console.log(data) },
      err => console.error(err),
      () =>  {this.editmode = "list"; this.reloadData()}
			);
		}
 }

 updateSnmpDev(oldId){
	 console.log(oldId);
	 console.log(this.snmpdevForm.value.id);
	 if(this.snmpdevForm.dirty && this.snmpdevForm.valid) {
		 var r = true;
		 if (this.snmpdevForm.value.id != oldId) {
			r = confirm("Changing Device ID from "+oldId+" to " +this.snmpdevForm.value.id+". Proceed?");
		 }
		if (r == true) {
				this.snmpDeviceService.editDevice(this.snmpdevForm.value, oldId)
				.subscribe(data => { console.log(data) },
	       err => console.error(err),
	       () =>  {this.editmode = "list"; this.reloadData()}
	 			);
		}
	}
 }

 getMeasGroupsforDevices(){
 this.measgroupsDeviceService.getMeasGroup(null)
 .subscribe(
	 data => { this.measgroups = data },
	 err => console.error(err),
	 () => console.log('DONE')
 	);
 }

 getInfluxServersforDevices(){
 this.influxserverDeviceService.getInfluxServer(null)
 .subscribe(
	 data => { this.influxservers = data },
	 err => console.error(err),
	 () => console.log('DONE')
 	);
 }

 getMeasFiltersforDevices(){
 this.measfiltersDeviceService.getMeasFilter(null)
 .subscribe(
	data => { this.measfilters = data },
	err => console.error(err),
	() => console.log('DONE')
	 );
 }
}
