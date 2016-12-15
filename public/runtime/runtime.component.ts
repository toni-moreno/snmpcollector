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
	islogLevelChanged : boolean = false;
	newLogLevel : string = null;

	loglLevelArray : Array<string> = [
		'panic',
		'fatal',
		'error',
		'warning',
		'info',
		'debug'
	];

	isObject(val) { return typeof val === 'object'; }
	isArray(val) { return typeof val === 'array'}


	viewItem(snmpdev){
	 this.runtime_dev = snmpdev;
	 console.log(snmpdev);
	}

	loadRuntimeById(id) {
	 this.islogLevelChanged = false;
	 this.newLogLevel = null;
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

	 onChangeLogLevel(level) {
		 console.log(level);
		 this.islogLevelChanged = true;
		 this.newLogLevel = level;
	 }


	 changeLogLevel(id) {
		console.log("ID,event");
		this.runtimeService.changeLogLevel(id,this.newLogLevel)
			.subscribe(
			 data => {
				 this.runtime_dev = data;
				 this.runtime_dev.CurLogLevel = this.newLogLevel;
				 this.runtime_dev.ID = id;
				 this.islogLevelChanged = !true;
				 this.reloadData();
			 },
			 err => console.error(err),
			 () => console.log('DONE')
		 );
	 }

		/*downloadLogFile(id) {
		console.log("Download Log file from device",id);
		var reader = new FileReader();
		this.runtimeService.downloadLogFile(id)
			.subscribe(
			 data => {
					reader.readAsArrayBuffer(data)
					console.log("download done")
			 },
			 err => {
					console.error(err)
					console.log("Error downloading the file.")
				},
			 () => console.log('Completed file download.')
		 );
		 reader.onloadend = function (e) {
				console.log("PRE",reader.result)
				//var blob = new Blob([reader.result],{type:"text/plain;charset=utf-8"})
				var blob = new Blob([reader.result])
				saveAs(blob,id+".log");
				//var file = new File([reader.result],id+".log",{type:"text/plain;charset=utf-8"});
				//saveAs(file);
		}
	 }*/

   downloadLogFile(id) {
   console.log("Download Log file from device",id);
   this.runtimeService.downloadLogFile(id)
     .subscribe(
      data => {
         saveAs(data,id+".log")
         console.log("download done")
      },
      err => {
         console.error(err)
         console.log("Error downloading the file.")
       },
      () => console.log('Completed file download.')
    );
  }

   forceFltUpdate(id) {
   console.log("ID,event",id,event);
   this.runtimeService.forceFltUpdate(id)
     .subscribe(
      data => {
         console.log("download done")
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
