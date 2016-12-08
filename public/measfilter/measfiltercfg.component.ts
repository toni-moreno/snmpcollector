import { Component, ChangeDetectionStrategy, ViewChild  } from '@angular/core';
import { FormBuilder,  Validators} from '@angular/forms';
import { MeasFilterService } from './measfiltercfg.service';
import { InfluxMeasService } from '../influxmeas/influxmeascfg.service';
import { ValidationService } from '../common/validation.service'

import { GenericModal } from '../common/generic-modal';

@Component({
  selector: 'measfilters',
  providers: [MeasFilterService, InfluxMeasService],
  templateUrl: 'public/measfilter/measfiltereditor.html',
	styleUrls:['public/css/component-styles.css']
})

export class MeasFilterCfgComponent {
	@ViewChild('viewModal') public viewModal: GenericModal;
	@ViewChild('viewModalDelete') public viewModalDelete: GenericModal;

  editmode: string; //list , create, modify
  measfilters: Array<any>;
  filter: string;
  measfilterForm: any;
	testmeasfilters: any;
	influxmeas: Array<any>;
	deleteobject: Object;

	//Initialization data, rows, colunms for Table
	private data:Array<any> = [];
	public rows:Array<any> = [];
	public columns:Array<any> = [
	{title: 'ID', name: 'ID'},
	{title: 'Measurement ID', name: 'IDMeasurementCfg'},
	{title: 'Filter Type', name: 'FType'},
	{title: 'FileName', name: 'FileName'},
	{title: 'EnableAlias', name: 'EnableAlias'},
	{title: 'OID Condition', name: 'OIDCond'},
	{title: 'Condition Type', name: 'CondType'},
	{title: 'Condition Value', name: 'CondValue'}
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

  constructor(public measFilterService: MeasFilterService, public measMeasFilterService: InfluxMeasService, builder: FormBuilder) {
	  this.editmode='list';
	  this.reloadData();
	  this.measfilterForm = builder.group({
			id: ['',Validators.required],
			IDMeasurementCfg: ['', Validators.required],
			FType: ['', Validators.required],
			FileName: [''],
			EnableAlias: [''],
			OIDCond: ['', ValidationService.OIDValidator],
			CondType: ['', Validators.required],
			CondValue: ['', Validators.required],
			Description: ['']
		});
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

	reloadData(){
	// now it's a simple subscription to the observable
		this.measFilterService.getMeasFilter(this.filter)
			.subscribe(
				data => {
					this.measfilters = data;
					this.data = data;
					this.onChangeTable(this.config);
				},
				err => console.error(err),
				() => console.log('DONE')
			);
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
	this.measFilterService.checkOnDeleteMeasFilter(id)
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
 newMeasFilter(){
	 this.editmode = "create";
	 this.getMeasforMeasFilters();
 }

 editMeasFilter(row){
	 let id = row.ID;
	 this.getMeasforMeasFilters();
	 this.measFilterService.getMeasFilterById(id)
 		 .subscribe(data => { this.testmeasfilters = data },
 		 err => console.error(err),
 		 () =>  this.editmode = "modify"
 		);
 	}

	deleteMeasFilter(id){
		this.measFilterService.deleteMeasFilter(id)
			.subscribe( data => {},
			err => console.error(err),
			() => {this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData()}
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
	 if(this.measfilterForm.valid) {
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
