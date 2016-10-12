import { Component, Pipe, PipeTransform  } from '@angular/core';
import {  FormBuilder,  Validators} from '@angular/forms';
import { InfluxMeasService } from './influxmeascfg.service';
import { SnmpMetricService } from './snmpmetriccfg.service';


@Component({
  selector: 'influxmeas',
  providers: [InfluxMeasService, SnmpMetricService],
  templateUrl: 'public/home/influxmeaseditor.html',
  styleUrls:['public/home/influxmeaseditor.css'],
})

export class InfluxMeasCfgComponent {
  editmode: string; //list , create, modify
  influxmeas: Array<any>;
  filter: string;
  influxmeasForm: any;
	testinfluxmeas: any;
	snmpmetrics: Array<any>;

  reloadData(){
  // now it's a simple subscription to the observable
    this.influxMeasService.getMeas(this.filter)
      .subscribe(
				data => { this.influxmeas = data },
		  	err => console.error(err),
		  	() => {console.log('DONE'); }
		  );
  }
  constructor(public influxMeasService: InfluxMeasService, public metricMeasService: SnmpMetricService, builder: FormBuilder) {
	  this.editmode='list';

	  this.reloadData();
	  this.influxmeasForm = builder.group({
			id: ['',Validators.compose([Validators.required, Validators.minLength(4)])],
  		Name: ['', Validators.required],
			GetMode: [''],
			IndexOID: [''],
			IndexTag: [''],
			Fields: [''],
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
	var r = confirm("Deleting MEASUREMENT: "+id+". Proceed?");
 	if (r == true) {
		 var result=this.influxMeasService.deleteMeas(id)
		 .subscribe(
			data => { console.log(data) },
			err => console.error(err),
			() =>  {this.editmode = "list"; this.reloadData()}
			);
		 console.log(result);
 	}
 }
 newMeas(){
	 this.editmode = "create";
	 this.getMetricsforMeas();
 }

 editMeas(id){
	 this.getMetricsforMeas();
	 this.influxMeasService.getMeasById(id)
 		 .subscribe(data => { this.testinfluxmeas = data },
 		 err => console.error(err),
 		 () =>  this.editmode = "modify"
 		);
 	}
 cancelEdit(){
	 this.editmode = "list";
 }
 saveInfluxMeas(){
	 if(this.influxmeasForm.dirty && this.influxmeasForm.valid) {
		 this.influxMeasService.addMeas(this.influxmeasForm.value)
		 .subscribe(data => { console.log(data) },
      err => console.error(err),
      () =>  {this.editmode = "list"; this.reloadData()}
			);
		}
 }

 updateInfluxMeas(oldId){
	 console.log(oldId);
	 console.log(this.influxmeasForm.value.id);
	 if(this.influxmeasForm.dirty && this.influxmeasForm.valid) {
		 var r = true;
		 if (this.influxmeasForm.value.id != oldId) {
			 r = confirm("Changing Measurement ID from "+oldId+" to " +this.influxmeasForm.value.id+". Proceed?");
		 }
		if (r == true) {
				this.influxMeasService.editMeas(this.influxmeasForm.value, oldId)
				.subscribe(data => { console.log(data) },
	       err => console.error(err),
	       () =>  {this.editmode = "list"; this.reloadData()}
	 			);
		}
	}
 }

 	getMetricsforMeas(){
		this.metricMeasService.getMetrics(null)
		.subscribe(
			data => { this.snmpmetrics = data },
			err => console.error(err),
			() => console.log('DONE')
		);
	}

}
