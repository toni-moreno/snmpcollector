import { Component, Pipe, ChangeDetectionStrategy, PipeTransform  } from '@angular/core';
import {  FormBuilder,  Validators} from '@angular/forms';
import { MeasGroupService } from './measgroupcfg.service';
import { InfluxMeasService } from './influxmeascfg.service';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';


@Component({
  selector: 'measgroups',
  providers: [MeasGroupService, InfluxMeasService],
  templateUrl: 'public/home/measgroupeditor.html',
  styleUrls:['public/home/measgroupeditor.css'],
})

export class MeasGroupCfgComponent {
  editmode: string; //list , create, modify
  measgroups: Array<any>;
  filter: string;
  measgroupForm: any;
	testmeasgroups: any;
	influxmeas: Array<any>;
  selectmeas: IMultiSelectOption[];

	private mySettings: IMultiSelectSettings = {
		pullRight: false,
		enableSearch: true,
		checkedStyle: 'glyphicon',
		buttonClasses: 'btn btn-default',
		selectionLimit: 0,
		closeOnSelect: false,
		showCheckAll: false,
		showUncheckAll: false,
		dynamicTitleMaxItems: 100,
		maxHeight: '300px',
	};

	private myTexts: IMultiSelectTexts = {
		checkAll: 'Check all',
		uncheckAll: 'Uncheck all',
		checked: 'checked',
		checkedPlural: 'checked',
		searchPlaceholder: 'Search...',
		defaultTitle: 'Select',
	};

	onChange(value){
		this.measgroupForm.controls['Measurements'].patchValue(value);
	}
  reloadData(){
  // now it's a simple subscription to the observable
    this.measGroupService.getMeasGroup(this.filter)
      .subscribe(
				data => { this.measgroups = data },
		  	err => console.error(err),
		  	() => console.log('DONE')
		  );
  }

  constructor(public measGroupService: MeasGroupService, public measMeasGroupService: InfluxMeasService, builder: FormBuilder) {
	  this.editmode='list';
	  this.reloadData();
	  this.measgroupForm = builder.group({
			id: ['',Validators.compose([Validators.required, Validators.minLength(4)])],
			Measurements: ['']
		});
  }

  onFilter(){
	this.reloadData();
  }

 viewItem(id,event){
	console.log('view',id);
 }
 removeItem(id){
	var r = confirm("Deleting GROUP: "+id+". Proceed?");
 	if (r == true) {
		 var result=this.measGroupService.deleteMeasGroup(id)
		 .subscribe(
			data => { console.log(data) },
			err => console.error(err),
			() =>  {this.editmode = "list"; this.reloadData()}
			);
		 console.log(result);
 	}
 }
 newMeasGroup(){
	 this.editmode = "create";
	 this.getMeasforMeasGroups();
 }

 editMeasGroup(id){
	 this.getMeasforMeasGroups();
	 this.measGroupService.getMeasGroupById(id)
 		 .subscribe(data => { this.testmeasgroups = data },
 		 err => console.error(err),
 		 () =>  this.editmode = "modify"
 		);
 	}
 cancelEdit(){
	 this.editmode = "list";
 }
 saveMeasGroup(){
	 if(this.measgroupForm.dirty && this.measgroupForm.valid) {
		 this.measGroupService.addMeasGroup(this.measgroupForm.value)
		 .subscribe(data => { console.log(data) },
      err => console.error(err),
      () =>  {this.editmode = "list"; this.reloadData()}
			);
			this.measgroupForm.reset();
		}
 }

 updateMeasGroup(oldId){
	 console.log(oldId);
	 console.log(this.measgroupForm.value.id);
	 if(this.measgroupForm.dirty && this.measgroupForm.valid) {
		 var r = true;
		 if (this.measgroupForm.value.id != oldId) {
			 r = confirm("Changing Measurement Group ID from "+oldId+" to " +this.measgroupForm.value.id+". Proceed?");
		 }
		if (r == true) {
				this.measGroupService.editMeasGroup(this.measgroupForm.value, oldId)
				.subscribe(data => { console.log(data) },
	       err => console.error(err),
	       () =>  {this.editmode = "list"; this.reloadData()}
	 			);
		}
	}
 }

 	getMeasforMeasGroups(){
		this.measMeasGroupService.getMeas(null)
		.subscribe(
			data => {
				this.influxmeas = data
				this.selectmeas = [];
        this.measgroupForm.controls['Measurements'].reset();
				for (let entry of data) {
					console.log(entry)
					this.selectmeas.push({'id' : entry.ID , 'name' : entry.ID});
				}
			},
			err => console.error(err),
			() => console.log('DONE')
		);
	}

}
