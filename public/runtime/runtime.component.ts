import { ChangeDetectionStrategy, Component } from '@angular/core';
import { FormBuilder,  Validators} from '@angular/forms';
import { RuntimeService } from './runtime.service';

@Component({
  selector: 'runtime',
  providers: [RuntimeService],
  templateUrl: 'public/runtime/runtimeview.html',
  styleUrls:['public/runtime/runtimeeditor.css'],
})

export class RuntimeComponent {
  public oneAtATime:boolean = true;
  editmode: string; //list , create, modify
  runtime_devs: Array<any>;
  filter: string;

	runtime_dev: any;
	subItem: any;

	isObject(val) { return typeof val === 'object'; }
	isArray(val) { return typeof val === 'array'}


	viewItem(snmpdev){
	 this.runtime_dev = snmpdev;
	 console.log(snmpdev);
	}

	loadRuntimeById(id) {
	 this.runtimeService.getRuntimeById(id)
		 .subscribe(
			data => {
				this.runtime_dev = data;
				this.runtime_dev.ID = id;
			},
			err => console.error(err),
			() => console.log('DONE')
		);
	 }

   changeActiveDevice(id,event) {
		 console.log("ID,event",id,event);
		 this.runtimeService.changeDeviceActive(id,event)
			 .subscribe(
				data => {
					this.runtime_dev = data;
					this.runtime_dev.DeviceActive = !data.DeviceActive;
					this.runtime_dev.ID = id;
					this.reloadData();
				},
				err => console.error(err),
				() => console.log('DONE')
			);
   }

	 changeStateDebug(id,event) {
		console.log("ID,event",id,event);
		this.runtimeService.changeStateDebug(id,event)
			.subscribe(
			 data => {
				 this.runtime_dev = data;
				 this.runtime_dev.StateDebug = !data.StateDebug;
				 this.runtime_dev.ID = id;
				 this.reloadData();

			 },
			 err => console.error(err),
			 () => console.log('DONE')
		 );
	 }


	reloadData(){
  // now it's a simple subscription to the observable
    this.runtimeService.getRuntime(this.filter)
      .subscribe(
				data => {
					this.runtime_devs = data
				},
		  	err => console.error(err),
		  	() => console.log('DONE')
		  );
  }

  constructor(public runtimeService: RuntimeService, builder: FormBuilder) {
	  this.editmode='list';
	  this.reloadData();
    }

  onFilter(){
	this.reloadData();
  }

/* updateSnmpDev(oldId){
	 console.log(oldId);
	 console.log(this.snmpdevForm.value.id);
	 console.log("FORM", this.snmpdevForm.value);
	 if(this.snmpdevForm.valid) {
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
}*/

}
