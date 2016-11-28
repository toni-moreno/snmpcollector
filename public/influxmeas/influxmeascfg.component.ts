import { Component, ChangeDetectionStrategy, ViewChild} from '@angular/core';
import {  FormBuilder,  Validators} from '@angular/forms';
import { InfluxMeasService } from './influxmeascfg.service';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';
import { SnmpMetricService } from '../snmpmetric/snmpmetriccfg.service';

import { GenericModal } from '../common/generic-modal';

@Component({
  selector: 'influxmeas',
  providers: [InfluxMeasService, SnmpMetricService],
  templateUrl: 'public/influxmeas/influxmeaseditor.html',
  styleUrls:['public/influxmeas/influxmeaseditor.css'],
})

export class InfluxMeasCfgComponent {
	@ViewChild('viewModal') public viewModal: GenericModal;
	@ViewChild('viewModalDelete') public viewModalDelete: GenericModal;

  editmode: string; //list , create, modify
  influxmeas: Array<any>;
  filter: string;
  influxmeasForm: any;
	testinfluxmeas: any;
	snmpmetrics: Array<any>;
 	selectmetrics: IMultiSelectOption[];
	deleteobject: Object;

	//Initialization data, rows, colunms for Table
	private data:Array<any> = [];
	public rows:Array<any> = [];
	public columns:Array<any> = [
	{title: 'ID', name: 'ID'},
	{title: 'Name', name: 'Name'},
	{title: 'GetMode', name: 'GetMode'},
	{title: 'Index OID', name: 'IndexOID'},
	{title: 'Index Tag', name: 'IndexTag'},
	{title: 'Index as Value', name: 'IndexAsValue'},
	{title: 'Metric Fields', name: 'Fields'}
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

  constructor(public influxMeasService: InfluxMeasService, public metricMeasService: SnmpMetricService, builder: FormBuilder) {
	  this.editmode='list';

	  this.reloadData();
	  this.influxmeasForm = builder.group({
			id: ['',Validators.required],
  		Name: ['', Validators.required],
			GetMode: ['', Validators.required],
			IndexOID: [''],
			IndexTag: [''],
      IndexAsValue: ['false'],
			Fields: ['', Validators.required],
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

	onChange(value){
		this.influxmeasForm.controls['Fields'].patchValue(value);
	}

  reloadData(){
  // now it's a simple subscription to the observable
    this.influxMeasService.getMeas(this.filter)
      .subscribe(
				data => {
					this.influxmeas = data
				 	this.data = data;
					this.onChangeTable(this.config);
					},
		  	err => console.error(err),
		  	() => {console.log('DONE'); }
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
	this.influxMeasService.checkOnDeleteInfluxMeas(id)
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

 newMeas(){
	 this.editmode = "create";
	 this.getMetricsforMeas();
 }

 editMeas(row){
	 let id = row.ID;
	 this.getMetricsforMeas();
	 this.influxMeasService.getMeasById(id)
 		 .subscribe(data => {
       this.testinfluxmeas = data
       //Update metrics fields
	 		this.influxmeasForm.controls['Fields'].patchValue(this.testinfluxmeas.Fields);
      },
 		 err => console.error(err),
 		 () =>  this.editmode = "modify"
 		);
 	}

	deleteInfluxMeas(id){
		this.influxMeasService.deleteMeas(id)
			.subscribe( data => {},
			err => console.error(err),
			() => {this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData()}
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
      this.influxmeasForm.reset();
		}
 }

 updateInfluxMeas(oldId){
	 console.log(oldId);
	 console.log(this.influxmeasForm.value.id);
	 if(this.influxmeasForm.valid) {
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
			data => {
        this.snmpmetrics = data;
				this.selectmetrics = [];
        this.influxmeasForm.controls['Fields'].reset();
				for (let entry of data) {
					console.log(entry)
					this.selectmetrics.push({'id' : entry.ID , 'name' : entry.ID});
				}
			},
			err => console.error(err),
			() => console.log('DONE')
		);
	}

}
