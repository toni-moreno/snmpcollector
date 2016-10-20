import { Component, ChangeDetectionStrategy, Pipe, PipeTransform  } from '@angular/core';
import {  FormBuilder,  Validators} from '@angular/forms';
import { SnmpMetricService } from './snmpmetriccfg.service';


@Component({
  selector: 'snmpmetrics',
  providers: [SnmpMetricService],
  templateUrl: 'public/home/snmpmetriceditor.html',
  styleUrls:['public/home/snmpmetriceditor.css'],

})

export class SnmpMetricCfgComponent {
  editmode: string; //list , create, modify
  snmpmetrics: Array<any>;
  filter: string;
  snmpmetForm: any;
	testsnmpmetric: any;

  reloadData(){
  // now it's a simple subscription to the observable
    this.snmpMetricService.getMetrics(this.filter)
    .subscribe(
			data => { this.snmpmetrics = data },
	  	err => console.error(err),
	  	() => console.log('DONE')
	  );
  }
  constructor(public snmpMetricService: SnmpMetricService,builder: FormBuilder) {
	  this.editmode='list';
	  this.reloadData();
	  this.snmpmetForm = builder.group({
			id: ['',Validators.compose([Validators.required, Validators.minLength(4)])],
  		FieldName: ['', Validators.required],
			Description: [''],
			BaseOID: [''],
			DataSrcType: [''],
			//Depending on datasrctype
			GetRate:[''],
			Scale: [''],
			Shift: ['']
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
	var r = confirm("Deleting METRIC: "+id+". Proceed?");
 	if (r == true) {
	 var result=this.snmpMetricService.deleteMetric(id)
	 .subscribe(
		data => { console.log(data) },
		err => console.error(err),
		() =>  {this.editmode = "list"; this.reloadData()}
		);
	 console.log(result);
 	}
 }
 newMetric(){
	 this.editmode = "create";
 }
 editMetric(id){
	 this.snmpMetricService.getMetricsById(id)
	 .subscribe(data => { this.testsnmpmetric = data },
	 err => console.error(err),
	 () =>  this.editmode = "modify"
   );
 }

 cancelEdit(){
	 this.editmode = "list";
 }

 saveSnmpMet(){
	 if(this.snmpmetForm.dirty && this.snmpmetForm.valid) {
	   this.snmpMetricService.addMetric(this.snmpmetForm.value)
     .subscribe(data => { console.log(data) },
      err => console.error(err),
      () =>  {this.editmode = "list"; this.reloadData()}
      );
		}
 }

 updateSnmpMet(oldId){
	 console.log(oldId);
	 console.log(this.snmpmetForm.value.id);
	 if(this.snmpmetForm.dirty && this.snmpmetForm.valid) {
		var r = true;
		if (this.snmpmetForm.value.id != oldId) {
      r = confirm("Changing Metric ID from "+oldId+" to " +this.snmpmetForm.value.id+". Proceed?");
		 }
		if (r == true) {
			this.snmpMetricService.editMetric(this.snmpmetForm.value, oldId)
			.subscribe(data => { console.log(data) },
      err => console.error(err),
      () =>  {this.editmode = "list"; this.reloadData()}
 			);
		}
	}
 }

}
