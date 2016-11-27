import { Component, ChangeDetectionStrategy, ViewChild  } from '@angular/core';
import { FormBuilder,  Validators} from '@angular/forms';
import { SnmpMetricService } from './snmpmetriccfg.service';
import { ControlMessagesComponent } from '../common/control-messages.component'
import { ValidationService } from '../common/validation.service'

import { GenericModal } from '../common/generic-modal';

@Component({
  selector: 'snmpmetrics',
  providers: [SnmpMetricService],
  templateUrl: 'public/snmpmetric/snmpmetriceditor.html',
  styleUrls:['public/snmpmetric/snmpmetriceditor.css'],

})

export class SnmpMetricCfgComponent {
	@ViewChild('viewModal') public viewModal: GenericModal;
	@ViewChild('viewModalDelete') public viewModalDelete: GenericModal;

  editmode: string; //list , create, modify
  snmpmetrics: Array<any>;
  filter: string;
  snmpmetForm: any;
	testsnmpmetric: any;

	//Initialization data, rows, colunms for Table
	private data:Array<any> = [];
	public rows:Array<any> = [];
	public columns:Array<any> = [
	{title: 'ID', name: 'ID'},
	{title: 'FieldName', name: 'FieldName'},
	{title: 'BaseOID', name: 'BaseOID'},
	{title: 'DataSrcType', name: 'DataSrcType'},
	{title: 'GetRate', name: 'GetRate'},
	{title: 'Scale', name: 'Scale'},
	{title: 'Shift', name: 'Shift'},
	{title: 'IsTag', name: 'IsTag'}
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

  constructor(public snmpMetricService: SnmpMetricService,builder: FormBuilder) {
	  this.editmode='list';
	  this.reloadData();
	  this.snmpmetForm = builder.group({
			id: ['',Validators.required],
  		FieldName: ['', Validators.required],
			BaseOID: ['', Validators.compose([Validators.required, ValidationService.OIDValidator])],
			DataSrcType: ['', Validators.required],
			//Depending on datasrctype
			GetRate:['false', Validators.required],
			Scale: ['0', Validators.required],
			Shift: ['0', Validators.required],
      IsTag:['false',Validators.required],
			Description: ['']
		});
  }

	reloadData(){
  // now it's a simple subscription to the observable
    this.snmpMetricService.getMetrics(this.filter)
    .subscribe(
			data => {
				this.snmpmetrics = data;
				this.data=data;
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
	this.snmpMetricService.checkOnDeleteMetric(id)
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
 newMetric(){
	 this.editmode = "create";
 }
 editMetric(row){
	 let id = row.ID;
	 this.snmpMetricService.getMetricsById(id)
	 .subscribe(data => { this.testsnmpmetric = data },
	 err => console.error(err),
	 () =>  this.editmode = "modify"
   );
 }
 deleteSNMPMetric(id){
	 this.snmpMetricService.deleteMetric(id)
		 .subscribe( data => {},
		 err => console.error(err),
		 () => {this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData()}
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
