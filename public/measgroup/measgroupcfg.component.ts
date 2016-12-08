import { Component, ChangeDetectionStrategy, ViewChild  } from '@angular/core';
import { FormBuilder,  Validators} from '@angular/forms';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';
import { MeasGroupService } from './measgroupcfg.service';
import { InfluxMeasService } from '../influxmeas/influxmeascfg.service';
import { ValidationService } from '../common/validation.service'

import { GenericModal } from '../common/generic-modal';

@Component({
  selector: 'measgroups',
  providers: [MeasGroupService, InfluxMeasService],
  templateUrl: 'public/measgroup/measgroupeditor.html',
  styleUrls:['public/measgroup/measgroupeditor.css'],
})

export class MeasGroupCfgComponent {
	@ViewChild('viewModal') public viewModal: GenericModal;
	@ViewChild('viewModalDelete') public viewModalDelete: GenericModal;

  editmode: string; //list , create, modify
  measgroups: Array<any>;
  filter: string;
  measgroupForm: any;
	testmeasgroups: any;
	influxmeas: Array<any>;
  selectmeas: IMultiSelectOption[];

	//Initialization data, rows, colunms for Table
	private data:Array<any> = [];
	public rows:Array<any> = [];
	public columns:Array<any> = [
	{title: 'ID', name: 'ID'},
	{title: 'Measurements', name: 'Measurements'},
	];

	public page:number = 1;
	public itemsPerPage:number = 10;
	public maxSize:number = 5;
	public numPages:number = 1;
	public length:number = 0;

	//Set config
	public config:any = {
		paging: true,
		sorting: {columns: this.columns},
		filtering: {filterString: ''},
		className: ['table-striped', 'table-bordered']
	};

	constructor(public measGroupService: MeasGroupService, public measMeasGroupService: InfluxMeasService, builder: FormBuilder) {
		this.editmode='list';
		this.reloadData();
		this.measgroupForm = builder.group({
			id: ['',Validators.required],
			Measurements: ['', Validators.compose([Validators.required,Validators.minLength(1)])],
			Description: ['']
		});
	}


  reloadData(){
  // now it's a simple subscription to the observable
    this.measGroupService.getMeasGroup(null)
      .subscribe(
				data => {
					this.measgroups = data;
				 	this.data = data;
					this.onChangeTable(this.config);
				},
		  	err => console.error(err),
		  	() => console.log('DONE')
		  );
  }

	public changePage(page:any, data:Array<any> = this.data):Array<any> {
		let start = (page.page - 1) * page.itemsPerPage;
		let end = page.itemsPerPage > -1 ? (start + page.itemsPerPage) : data.length;
		return data.slice(start, end);
	}

	public changeSort(data:any, config:any):any {
		if (!config.sorting) {
			return data;
		}

		let columns = this.config.sorting.columns || [];
		let columnName:string = void 0;
		let sort:string = void 0;

		for (let i = 0; i < columns.length; i++) {
			if (columns[i].sort !== '' && columns[i].sort !== false) {
				columnName = columns[i].name;
				sort = columns[i].sort;
			}
		}

		if (!columnName) {
			return data;
		}

		// simple sorting
		return data.sort((previous:any, current:any) => {
			if (previous[columnName] > current[columnName]) {
				return sort === 'desc' ? -1 : 1;
			} else if (previous[columnName] < current[columnName]) {
				return sort === 'asc' ? -1 : 1;
			}
			return 0;
		});
	}

	public changeFilter(data:any, config:any):any {
		let filteredData:Array<any> = data;
		this.columns.forEach((column:any) => {
			if (column.filtering) {
				filteredData = filteredData.filter((item:any) => {
					return item[column.name].match(column.filtering.filterString);
				});
			}
		});

		if (!config.filtering) {
			return filteredData;
		}

		if (config.filtering.columnName) {
			return filteredData.filter((item:any) =>
				item[config.filtering.columnName].match(this.config.filtering.filterString));
		}

		let tempArray:Array<any> = [];
		filteredData.forEach((item:any) => {
			let flag = false;
			this.columns.forEach((column:any) => {
				if(!item[column.name]){
					item[column.name] = '--'
				}
					if (item[column.name].toString().match(this.config.filtering.filterString)) {
						flag = true;
					}

			});
			if (flag) {
				tempArray.push(item);
			}
		});
		filteredData = tempArray;

		return filteredData;
	}

	public onChangeTable(config:any, page:any = {page: this.page, itemsPerPage: this.itemsPerPage}):any {
		if (config.filtering) {
			Object.assign(this.config.filtering, config.filtering);
		}
		if (config.sorting) {
			Object.assign(this.config.sorting, config.sorting);
		}
		let filteredData = this.changeFilter(this.data, this.config);
		let sortedData = this.changeSort(filteredData, this.config);
		this.rows = page && config.paging ? this.changePage(page, sortedData) : sortedData;
		this.length = sortedData.length;
	}

	public onCellClick(data: any): any {
		console.log(data);
	}

	onChange(value){
		this.measgroupForm.controls['Measurements'].patchValue(value);
	}

  onFilter(){
	this.reloadData();
  }

 viewItem(id,event){
	console.log('view',id);
	this.viewModal.parseObject(id);
 }
 removeItem(row){
	let id = row.ID;
	console.log('remove',id);
	this.measGroupService.checkOnDeleteMeasGroups(id)
	.subscribe(
	 	 data => {
		 console.log(data);
		 let temp = data;
		 this.viewModalDelete.parseObject(temp)
	 },
	 err => console.error(err),
	 () =>  {}
	);
 }

 newMeasGroup(){
	 this.editmode = "create";
	 this.getMeasforMeasGroups();
 }

 editMeasGroup(row){
	 let id = row.ID;
	 this.getMeasforMeasGroups();
	 this.measGroupService.getMeasGroupById(id)
 		 .subscribe(data => {
			this.testmeasgroups = data;
			 //Update measurements
	 		this.measgroupForm.controls['Measurements'].patchValue(this.testmeasgroups.Measurements);
		  },
 		 err => console.error(err),
 		 () =>  this.editmode = "modify"
 		);
 	}
	deleteMeasGroup(id){
		this.measGroupService.deleteMeasGroup(id)
			.subscribe( data => {},
			err => console.error(err),
			() => {this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData()}
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
		}
 }

 updateMeasGroup(oldId){
	 console.log(oldId);
	 console.log(this.measgroupForm.value.id);
	 if(this.measgroupForm.valid) {
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
