import { Component, Pipe, PipeTransform  } from 'angular2/core';
import { CORE_DIRECTIVES } from 'angular2/common';
import {FORM_DIRECTIVES, FORM_BINDINGS,FormBuilder, NgFormModel, ControlGroup, Control, Validators} from 'angular2/common';
import { ACCORDION_DIRECTIVES } from 'ng2-bootstrap';
import { MeasFilterService } from './measfiltercfg.service';
import { InfluxMeasService } from './influxmeascfg.service';
import {ControlMessages} from './control-messages.component';

@Component({
  selector: 'measfilters',
  providers: [MeasFilterService, InfluxMeasService],
  templateUrl: '/public/home/measfiltereditor.html',
  styleUrls:['public/home/measfiltereditor.css'],
  bindings: [MeasFilterService, InfluxMeasService],
  viewBindings: [FORM_BINDINGS],
  directives: [ACCORDION_DIRECTIVES,CORE_DIRECTIVES,FORM_DIRECTIVES,ControlMessages]
})

export class MeasFilterCfgComponent {
  editmode: string; //list , create, modify
  measfilters: Array<any>;
  filter: string;
  measfilterForm: ControlGroup;
	testmeasfilters: any;
	influxmeas: Array<any>;

  reloadData(){
  // now it's a simple subscription to the observable
    this.measFilterService.getMeasFilter(this.filter)
      .subscribe(
				data => { this.measfilters = data },
		  	err => console.error(err),
		  	() => console.log('DONE')
		  );
  }
  constructor(public measFilterService: MeasFilterService, public measMeasFilterService: InfluxMeasService, builder: FormBuilder) {
	  this.editmode='list';
	  this.reloadData();
	  this.measfilterForm = builder.group({
			id: ['',Validators.compose([Validators.required, Validators.minLength(4)])],
			IDMeasurementCfg: [''],
			FType: [''],
			FileName: [''],
			EnableAlias: [''],
			OIDCond: [''],
			CondType: [''],
			CondValue: ['']
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
	var r = confirm("Deleting FILTER: "+id+". Proceed?");
 	if (r == true) {
		 var result=this.measFilterService.deleteMeasFilter(id)
		 .subscribe(
			data => { console.log(data) },
			err => console.error(err),
			() =>  {this.editmode = "list"; this.reloadData()}
			);
		 console.log(result);
 	}
 }
 newMeasFilter(){
	 this.editmode = "create";
	 this.getMeasforMeasFilters();
 }

 editMeasFilter(id){
	 this.getMeasforMeasFilters();
	 this.measFilterService.getMeasFilterById(id)
 		 .subscribe(data => { this.testmeasfilters = data },
 		 err => console.error(err),
 		 () =>  this.editmode = "modify"
 		);
 	}
 cancelEdit(){
	 this.editmode = "list";
 }
 saveMeasFilter(){
	 if(this.measfilterForm.dirty && this.measfilterForm.valid) {
		 this.measFilterService.addMeasFilter(this.measfilterForm.value)
		 .subscribe(data => { console.log(data) },
      err => console.error(err),
      () =>  {this.editmode = "list"; this.reloadData()}
			);
		}
 }

 updateMeasFilter(oldId){
	 console.log(oldId);
	 console.log(this.measfilterForm.value.id);
	 if(this.measfilterForm.dirty && this.measfilterForm.valid) {
		 var r = true;
		 if (this.measfilterForm.value.id != oldId) {
			 r = confirm("Changing Measurement Filter ID from "+oldId+" to " +this.measfilterForm.value.id+". Proceed?");
		 }
		if (r == true) {
				this.measFilterService.editMeasFilter(this.measfilterForm.value, oldId)
				.subscribe(data => { console.log(data) },
	       err => console.error(err),
	       () =>  {this.editmode = "list"; this.reloadData()}
	 			);
		}
	}
 }

 	getMeasforMeasFilters(){
		this.measMeasFilterService.getMeas(null)
		.subscribe(
			data => { this.influxmeas = data },
			err => console.error(err),
			() => console.log('DONE')
		);
	}

}
