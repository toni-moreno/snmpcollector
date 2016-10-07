import { Component, Pipe, PipeTransform  } from 'angular2/core';
import { CORE_DIRECTIVES } from 'angular2/common';
import {FORM_DIRECTIVES, FORM_BINDINGS,FormBuilder, NgFormModel, ControlGroup, Control, Validators} from 'angular2/common';
import { ACCORDION_DIRECTIVES } from 'ng2-bootstrap';
import { InfluxServerService } from './influxservercfg.service';
import {ControlMessages} from './control-messages.component';

@Component({
  selector: 'influxservers',
  providers: [InfluxServerService],
  templateUrl: '/public/home/influxservereditor.html',
  styleUrls:['public/home/influxservereditor.css'],
  bindings: [InfluxServerService],
  viewBindings: [FORM_BINDINGS],
  directives: [ACCORDION_DIRECTIVES,CORE_DIRECTIVES,FORM_DIRECTIVES,ControlMessages]
})

export class InfluxServerCfgComponent {
  editmode: string; //list , create, modify
  influxservers: Array<any>;
  filter: string;
  influxserverForm: ControlGroup;
	testinfluxservers: any;

  reloadData(){
  // now it's a simple subscription to the observable
    this.influxServerService.getInfluxServer(this.filter)
      .subscribe(
				data => { this.influxservers = data },
		  	err => console.error(err),
		  	() => console.log('DONE')
		  );
  }
  constructor(public influxServerService: InfluxServerService, builder: FormBuilder) {
	  this.editmode='list';
	  this.reloadData();
	  this.influxserverForm = builder.group({
			id: ['',Validators.compose([Validators.required, Validators.minLength(4)])],
			Host: [''],
			Port: [''],
			DB: [''],
			User: [''],
			Password: [''],
			Retention: [''],
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
	var r = confirm("Deleting INFLUXSERVER: "+id+". Proceed?");
 	if (r == true) {
		 var result=this.influxServerService.deleteInfluxServer(id)
		 .subscribe(
			data => { console.log(data) },
			err => console.error(err),
			() =>  {this.editmode = "list"; this.reloadData()}
			);
		 console.log(result);
 	}
 }
 newInfluxServer(){
	 this.editmode = "create";
 }

 editInfluxServer(id){
	 this.influxServerService.getInfluxServerById(id)
 		 .subscribe(data => { this.testinfluxservers = data },
 		 err => console.error(err),
 		 () =>  this.editmode = "modify"
 		);
 	}
 cancelEdit(){
	 this.editmode = "list";
 }
 saveInfluxServer(){
	 if(this.influxserverForm.dirty && this.influxserverForm.valid) {
		 this.influxServerService.addInfluxServer(this.influxserverForm.value)
		 .subscribe(data => { console.log(data) },
      err => console.error(err),
      () =>  {this.editmode = "list"; this.reloadData()}
			);
		}
 }

 updateInfluxServer(oldId){
	 console.log(oldId);
	 console.log(this.influxserverForm.value.id);
	 if(this.influxserverForm.dirty && this.influxserverForm.valid) {
		 var r = true;
		 if (this.influxserverForm.value.id != oldId) {
			 r = confirm("Changing Influx Server ID from "+oldId+" to " +this.influxserverForm.value.id+". Proceed?");
		 }
		if (r == true) {
				this.influxServerService.editInfluxServer(this.influxserverForm.value, oldId)
				.subscribe(data => { console.log(data) },
	       err => console.error(err),
	       () =>  {this.editmode = "list"; this.reloadData()}
	 			);
		}
	}
 }

}
