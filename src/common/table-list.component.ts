import { Component, Input, Output, EventEmitter, forwardRef, IterableDiffers, SimpleChanges } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Observable } from 'rxjs/Rx';

import { ItemsPerPageOptions } from './global-constants';
import { TableActions } from './table-actions';
import { AvailableTableActions } from './table-available-actions';
import { OnInit, OnChanges } from '@angular/core';

import { ChangeDetectionStrategy } from '@angular/core';
import { ChangeDetectorRef } from '@angular/core';

declare var _: any;

@Component({
  selector: 'table-list',
  styles: [`
		a { outline: none !important; }
	`],
  template: `
    <div class="row">
    <div class="col-md-8 text-left">
    <!--Filtering section-->
      <label [tooltip]="'Clear Filter'" container="body" (click)="onResetFilter()" style="margin-top: 10px"><i class="glyphicon glyphicon-trash text-primary"></i></label>
      <input *ngIf="config.filtering" placeholder="Filter all columns" required = "false" [(ngModel)]="myFilterValue" [ngTableFiltering]="config.filtering" class="form-control select-pages" (tableChanged)="onChangeTable(config)" />
      <span [ngClass]="length > 0 ? ['label label-info'] : ['label label-warning']" style="font-size : 100%">{{length}} Results</span>
    <!--Table Actions-->
    <ng-container *ngIf="tableRole === 'fulledit'">
      <button style ="margin-top: -5px;" type="button" (click)="customClick('new')" class="btn btn-primary"><i class="glyphicon glyphicon-plus"></i> New</button>
    </ng-container>
    <ng-container *ngIf="tableRole !== 'viewonly'">
      <button style ="margin-top: -5px;" type="button" (click)="enableEdit()" class="btn btn-primary"><i class="glyphicon glyphicon-edit"></i> {{editEnabled === false ? 'Enable Edit' : 'Disable Edit' }}</button>
    </ng-container>
      </div>
    <!--Items per page selection-->
    <div class="col-md-4 text-right">
      <span style="margin-left: 20px"> Items per page: </span>
      <select class="select-pages" style="width:auto" [ngModel]="itemsPerPage || 'All'" (ngModelChange)="changeItemsPerPage($event)">
        <option *ngFor="let option of itemsPerPageOptions" style="padding-left:2px" [value]="option.value">{{option.title}}</option>
      </select>
    </div>
    </div>
    <br>
    <!--Table available actions-->
    <table-actions [editEnabled]="editEnabled" [counterErrors]="counterErrors" [counterItems]="counterItems || 0" [itemsSelected]="selectedArray.length" [tableAvailableActions]="tableAvailableActions" (actionApply)="customClick('tableaction',$event,selectedArray)"></table-actions>
    <my-spinner [isRunning]="isRequesting"></my-spinner>
    <!--Table with data -->
    <ng-table *ngIf="isRequesting === false && data"
      [rows]="rows"
      [columns]="columns"
      [sanitizeCell]="sanitizeCell"
      [config]="config"
      [exportType]="typeComponent"
      [(checkedItems)]="selectedArray"
      [editMode]="editEnabled"
      (tableChanged)="onChangeTable(config)"
      (customClicked)="customClick($event.action, $event.row)"
      [tableRole]="tableRole"
      [roleActions]="roleActions">
    </ng-table>

    <!-- Pagination -->
    <pagination *ngIf="config.paging" class="pagination-sm" [ngModel]="page" [totalItems]="length" [itemsPerPage]="itemsPerPage" [maxSize]="maxSize" [boundaryLinks]="false" [rotate]="false" (pageChanged)="onChangeTable(config, $event)" (numPages)="numPages = $event">
    </pagination>
    <pre *ngIf="config.paging" class="card card-block card-header">Page: {{page}} / {{numPages}}</pre>
    `,
  styleUrls: ['../css/component-styles.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class TableListComponent implements OnInit, OnChanges {

  //Inputs
  @Input() typeComponent: string;
  @Input() columns: Array<any>;
  @Input() data: Array<any>;
  @Input() counterItems: any = 0;
  @Input() counterErrors: any = [];
  @Input() selectedArray: any = [];
  @Input() isRequesting: boolean = false;
  @Input() overrideEditEnabled : boolean = false;
  @Input() tableRole : any = 'fulledit';
  @Input() roleActions : any = [
    {'name':'export', 'type':'icon', 'icon' : 'glyphicon glyphicon-download-alt text-info', 'tooltip': 'Export item'},
    {'name':'view', 'type':'icon', 'icon' : 'glyphicon glyphicon-eye-open text-success', 'tooltip': 'View item'},
    {'name':'edit', 'type':'icon', 'icon' : 'glyphicon glyphicon-edit text-warning', 'tooltip': 'Edit item'},
    {'name':'remove', 'type':'icon', 'icon' : 'glyphicon glyphicon glyphicon-remove text-danger', 'tooltip': 'Remove item'}
  ]

  @Input() sanitizeCell: Function;
  @Input() public tableAvailableActions: any;

  @Output() public customClicked: EventEmitter<any> = new EventEmitter();

  //Vars
  private editEnabled: boolean = false;
  private rows: Array<any> = [];
  public page: number = 1;
  public itemsPerPage: number = 20;
  public itemsPerPageOptions: any = ItemsPerPageOptions;
  public maxSize: number = 5;
  public numPages: number = 1;
  public length: number = 0;
  public myFilterValue: any;

  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.columns },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  ngOnChanges(changes: SimpleChanges) {
    if (!this.data) this.data = [];
    this.onChangeTable(this.config);
    this.cd.markForCheck();
  }

  ngOnInit() {
    this.config.sorting = { columns: this.columns };
    this.onChangeTable(this.config)
  }

  constructor(public cd: ChangeDetectorRef) { }

  //Enable edit tables
  enableEdit() {
    this.editEnabled = !this.editEnabled;
    let obsArray = [];
    if (this.editEnabled === true) {
      if (this.overrideEditEnabled) this.customClick('editenabled',this.editEnabled);
      else this.tableAvailableActions = new AvailableTableActions(this.typeComponent).availableOptions;
    }
  }

  public changePage(page: any, data: Array<any> = this.data): Array<any> {
    //Check if we have to change the actual page
    let maxPage = Math.ceil(data.length / this.itemsPerPage);
    if (page.page > maxPage && page.page != 1) this.page = page.page = maxPage;
    let start = (page.page - 1) * page.itemsPerPage;
    let end = page.itemsPerPage > -1 ? (start + page.itemsPerPage) : data.length;
    return data.slice(start, end);
  }

  public changeSort(data: any, config: any): any {
    if (!config.sorting) {
      return data;
    }

    let columns = this.config.sorting.columns || [];
    let columnName: string = void 0;
    let sort: string = void 0;

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
    return data.sort((previous: any, current: any) => {
      if (previous[columnName] > current[columnName]) {
        return sort === 'desc' ? -1 : 1;
      } else if (previous[columnName] < current[columnName]) {
        return sort === 'asc' ? -1 : 1;
      }
      return 0;
    });
  }

  public changeFilter(data: any, config: any): any {
    let filteredData: Array<any> = data;
    this.columns.forEach((column: any) => {
      if (column.filtering) {
        filteredData = filteredData.filter((item: any) => {
          return item[column.name].match(column.filtering.filterString);
        });
      }
    });

    if (!config.filtering) {
      return filteredData;
    }

    if (config.filtering.columnName) {
      return filteredData.filter((item: any) =>
        item[config.filtering.columnName].match(this.config.filtering.filterString));
    }

    let tempArray: Array<any> = [];
    filteredData.forEach((item: any) => {
      let flag = false;
      this.columns.forEach((column: any) => {
        if (item[column.name] === null) {
          item[column.name] = ''
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

  changeItemsPerPage(items) {
    this.itemsPerPage = parseInt(items);
    let maxPage = Math.ceil(this.length / this.itemsPerPage);
    if (this.page > maxPage) this.page = maxPage;
    this.onChangeTable(this.config);
  }

  public onChangeTable(config: any, page: any = { page: this.page, itemsPerPage: this.itemsPerPage }): any {

    if (config) {
      if (config.filtering) {
        Object.assign(this.config.filtering, config.filtering);
      }
      if (config.sorting) {
        Object.assign(this.config.sorting, config.sorting);
      }
    }
    let filteredData = this.changeFilter(this.data, this.config);
    let sortedData = this.changeSort(filteredData, this.config);
    this.rows = page && config.paging ? this.changePage(page, sortedData) : sortedData;
    this.length = sortedData.length;
  }

  onResetFilter(): void {
    this.page = 1;
    this.myFilterValue = "";
    this.config.filtering = { filtering: { filterString: '' } };
    this.onChangeTable(this.config);
  }

  customClick(clicked: string, event: any = "", data: any = ""): void {
    this.customClicked.emit({ 'option': clicked, 'event': event, 'data': data });
  }

}
