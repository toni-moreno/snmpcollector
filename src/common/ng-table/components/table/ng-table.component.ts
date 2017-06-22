import { Component, EventEmitter, Input, Output, Pipe, PipeTransform } from '@angular/core';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { TooltipModule } from 'ng2-bootstrap';
import { ElapsedSecondsPipe } from '../../../elapsedseconds.pipe';


@Component({
  selector: 'ng-table',
  template: `
    <table class="table dataTable" ngClass="{{config.className || ''}}"
           role="grid" style="width: 100%;">
      <thead>
        <tr role="row">
        <th *ngIf="editMode == true">
        <p>Select</p>
        <label (click)="selectAllItems(true)" class="glyphicon glyphicon-check text-success"></label>-
        <label (click)="selectAllItems(false)" class="glyphicon glyphicon-unchecked text-danger"></label>
        </th>
          <th *ngIf="showCustom == true" [ngStyle]="exportType ? {'min-width': '95px'} : {'min-width': '80px'}">
          </th>
          <th *ngIf="showStatus == true" [ngStyle]="exportType ? {'min-width': '95px'} : {'min-width': '80px'}">
          </th>
          <ng-container *ngIf="checkRows">
          <th style="width: 80px; text-align: center; padding-bottom: 15px;" [ngClass]="rows.length === 0 ? ['bg-warning'] : ''">
            <i *ngIf="rows.length == 0" class="glyphicon glyphicon-warning-sign label label-warning" container=body [tooltip]="'No results'"> </i>
          </th>
          </ng-container>
          <th *ngFor="let column of columns" [ngTableSorting]="config" [column]="column"
              (sortChanged)="onChangeTable($event)" ngClass="{{column.className || ''}}" style="vertical-align: middle; text-align: center; width:auto !important;" container="body" [tooltip]="column.tooltip">
            {{column.title}} <i *ngIf="column.icon" [ngClass]="'glyphicon glyphicon-'+column.icon"></i>
            <i *ngIf="config && column.sort" class="pull-right glyphicon"
              [ngClass]="{'glyphicon-chevron-down': column.sort === 'desc', 'glyphicon-chevron-up': column.sort === 'asc'}"></i>
          </th>
          <ng-container *ngIf="extraActions">
          <th *ngFor="let action of extraActions" style="vertical-align: middle; text-align: center; width:auto !important;">
            {{action.title}}
          </th>
          </ng-container>
        </tr>
      </thead>
      <tbody>
      <tr *ngIf="showFilterRow">
      <td *ngIf="showCustom == true" style="width:100px">
      </td>
      <td *ngIf="showStatus == true" style="width:100px">
      </td>

        <td *ngFor="let column of columns">
          <input *ngIf="column.filtering" placeholder="{{column.filtering.placeholder}}"
                 [ngTableFiltering]="column.filtering"
                 class="form-control"
                 style="width: auto;"
                 (tableChanged)="onChangeTable(config)"/>
        </td>
      </tr>
        <tr *ngFor="let row of rows">
        <td *ngIf="editMode == true">
        <i [ngClass]="checkItems(row.ID) ? ['glyphicon glyphicon-unchecked text-danger'] : ['glyphicon glyphicon-check text-success']" (click)="selectItem(row)"></i>
        </td>
          <td *ngIf="showCustom == true">
          <i *ngIf="exportType" class="glyphicon glyphicon-download-alt" [tooltip]="'Export item'" (click)="exportItem(row, exportType)"></i>
          <i class="glyphicon glyphicon-eye-open" [tooltip]="'View item'" (click)="viewItem(row)"></i>
					<i class="glyphicon glyphicon-edit"  [tooltip]="'Edit item'" (click)="editItem(row)"></i>
    			<i class="glyphicon glyphicon glyphicon-remove"  [tooltip]="'Remove Item'" (click)="removeItem(row)"></i>
          </td>
          <td *ngIf="showStatus == true" style="min-width: 170px">
          <label style="display: inline; margin-right: 2px" container="body" [tooltip]="'View '+ row.ID" class="label label-primary glyphicon glyphicon-eye-open" (click)="viewItem(row)"></label>
          <span style="border-right: 1px solid #1B809E; padding-right: 6px">
          <label style="display: inline; margin-right: 2px" container="body" [tooltip]="'Test connection '+ row.ID" class="label label-primary glyphicon glyphicon glyphicon-flash" (click)="testConnection(row)"></label>
          </span>
          <span style="padding-left: 12px">
          <label style="display: inline; margin-right: 2px" container="body" [tooltip]="row.DeviceActive ?  'Active' : 'Not active'" [ngClass]="row.DeviceActive ?  'glyphicon glyphicon-play label label-success' : 'glyphicon glyphicon-pause label label-danger'"></label>
          </span>
          <label style="display: inline; margin-right: 2px" container="body" [tooltip]="row.DeviceConnected ?  'Connected' : 'Not connected'" [ngClass]="row.DeviceConnected ?  'glyphicon glyphicon-globe label label-success' : 'glyphicon glyphicon-warning-sign label label-danger'"></label>
          </td>
          <ng-container *ngIf="checkRows">
          <td style="width: 80px; text-align: center; padding-top: 15px;" [ngClass]="row.valid === true? ['bg-success'] : ['bg-warning']">
            <i [ngClass]="row.valid === true ? ['glyphicon glyphicon-ok-sign label label-success'] : ['glyphicon glyphicon-warning-sign label label-warning']"> </i>
          </td>
          </ng-container>
          <td [ngClass]="row.tooltipInfo ? (row.tooltipInfo[column.name] ? (row.tooltipInfo[column.name]['Valid'] === true ? ['bg-success'] : ['bg-danger']) : '') : ''" (click)="cellClick(row, column.name)" *ngFor="let column of columns; let i = index" container=body [tooltip]="row.tooltipInfo ? tooltipValues : (column.name === 'ID' ? row.Description : '')" [innerHtml]="sanitize(row[column.name],column.transform)" style="text-align:right">

          <template #tooltipValues>
            <h6>Index:{{row.Index }}</h6>
            <h6>Metric:{{column[name]}}</h6>
            <hr>
            <div *ngFor="let test of (row.tooltipInfo[column.name] | keyParser)" style="text-align:left !important">
            <span class="text-left" style="paddint-left: 10px"><b>{{test.key}}</b></span>
            <span class="text-right" *ngIf="test.key === 'CurTime'"> {{test.value | date:'d/M/y H:m:s'}}</span>
            <span class="text-right" *ngIf="test.key === 'LastTime'"> {{test.value | date:'d/M/y H:m:s'}}</span>
            <span class="text-right" *ngIf="test.key !== 'CurTime' && test.key !== 'LastTime'">{{test.value}} </span>
            </div>
          </template>

          </td>
          <ng-container *ngIf="extraActions">
            <td *ngFor="let action of extraActions" style="text-align: center">
              <button *ngIf="action.type == 'boolean'" (click)="extraActionClick(row,action.title,!row[action.property])"
                [ngClass]="row[action.property] ?  ['btn btn-success'] : ['btn btn-danger']" [innerHtml]="row[action.property] ? action.content['enabled'] : action.content['disabled']">
              </button>
              <button *ngIf="action.type == 'button'" class="btn btn-primary" (click)="extraActionClick(row,action.title)">
                <span>{{action.content['enabled']}}</span>
              </button>

            </td>
          </ng-container>
        </tr>
      </tbody>
    </table>
  `,
  styles: [`
    :host >>> .displayinline{
      display: inline !important;
      padding-left: 5px;
    }
`]
})
export class NgTableComponent {
  // Table values
  @Input() public rows: Array<any> = [];
  @Input() public showCustom: boolean = false;
  @Input() public showStatus: boolean = false;
  @Input() public editMode: boolean = false;
  @Input() public exportType: string;
  @Input() public extraActions: Array<any>;
  @Input() checkedItems: Array<any>;
  @Input() checkRows: Array<any>;

  @Input()
  public set config(conf: any) {
    if (!conf.className) {
      conf.className = 'table-striped table-bordered';
    }
    if (conf.className instanceof Array) {
      conf.className = conf.className.join(' ');
    }
    this._config = conf;
  }

  public reportMetricStatus: Array<Object> = [
    { value: 0, name: 'Never Report', icon: 'glyphicon glyphicon-remove-circle', class: 'text-danger' },
    { value: 1, name: 'Report', icon: 'glyphicon glyphicon-ok-circle', class: 'text-success' },
    { value: 2, name: 'Report if not zero', icon: 'glyphicon glyphicon-ban-circle', class: 'text-warning' }
  ];

  // Outputs (Events)
  @Output() public tableChanged: EventEmitter<any> = new EventEmitter();
  @Output() public cellClicked: EventEmitter<any> = new EventEmitter();
  @Output() public viewedItem: EventEmitter<any> = new EventEmitter();
  @Output() public editedItem: EventEmitter<any> = new EventEmitter();
  @Output() public removedItem: EventEmitter<any> = new EventEmitter();
  @Output() public exportedItem: EventEmitter<any> = new EventEmitter();
  @Output() public testedConnection: EventEmitter<any> = new EventEmitter();
  @Output() public extraActionClicked: EventEmitter<any> = new EventEmitter();

  public showFilterRow: Boolean = false;

  @Input()
  public set columns(values: Array<any>) {
    this._columns = [];
    values.forEach((value: any) => {
      if (value.filtering) {
        this.showFilterRow = true;
      }
      if (value.className && value.className instanceof Array) {
        value.className = value.className.join(' ');
      }
      let column = this._columns.find((col: any) => col.name === value.name);
      if (column) {
        Object.assign(column, value);
      }
      if (!column) {
        this._columns.push(value);
      }
    });
  }

  private _columns: Array<any> = [];
  private _config: any = {};

  public constructor(private sanitizer: DomSanitizer) {
  }

  public sanitize(html: string, transform?: any ): SafeHtml {
    if  (transform === "elapsedseconds") {
      let test = new ElapsedSecondsPipe().transform(html,'3');
      html = test.toString();
    }
    if (typeof html === 'object') {
      var test: any = '<ul class="list-unstyled">';
      for (var item of html) {
        if (typeof item === 'object') {
          test += "<li>";
          for (var item2 in Object(item)) {
            if (typeof item[item2] === 'boolean') {
              if (item[item2]) test += ' <i class="glyphicon glyphicon-arrow-right"></i>'
              else test += ' <i class="glyphicon glyphicon-alert"></i>'
            } else if (typeof item[item2] === 'number') {
              test += '<i class="' + this.reportMetricStatus[item[item2]]['icon'] + ' ' + this.reportMetricStatus[item[item2]]['class'] + ' displayinline"></i>'
            } else if (item2 === 'TagID') {
              test += '<h4 class="text-success displayinline">'+item[item2] +' - </h4>';
            } else {
              test +='<span>'+item[item2]+'</span>';
            }
          }
          test += "</li>";
        } else {
          test += "<li>" + item + "</li>";
        }
      }
      test += "</ul>"
      return test;
    } else if (typeof html === 'boolean') {
      if (html) return '<i class="glyphicon glyphicon-ok"></i>'
      else return '<i class="glyphicon glyphicon-remove"></i>'
    }
    else {
      return this.sanitizer.bypassSecurityTrustHtml(html);
    }
  }

  public get columns(): Array<any> {
    return this._columns;
  }

  public get config(): any {
    return this._config;
  }

  public get configColumns(): any {
    let sortColumns: Array<any> = [];

    this.columns.forEach((column: any) => {
      if (column.sort) {
        sortColumns.push(column);
      }
    });
    return { columns: sortColumns };
  }

  public onChangeTable(column: any): void {
    this._columns.forEach((col: any) => {
      if (col.name !== column.name && col.sort !== false) {
        col.sort = '';
      }
    });
    this.tableChanged.emit({ sorting: this.configColumns });
  }

  selectAllItems(selectAll) {
    //Creates the form array
    if (selectAll === true) {
      for (let row of this.rows) {
        if (this.checkItems(row.ID)) {
          this.checkedItems.push(row);
        }
      }
    } else {
      for (let row of this.rows) {
        let index = this.findIndexItem(row.ID);
          if (index) this.deleteItem(index);
      }
    }
  }

  checkItems(item: any) : boolean {
    //Extract the ID from finalArray and loaded Items:
    let exist = true;
    for (let a of this.checkedItems) {
      if (item === a.ID) {
        exist = false;
      }
    }
    return exist;
  }

  selectItem(row : any) : void {
    if (this.checkItems(row.ID)) {
      this.checkedItems.push(row);
    }
    else {
      let index = this.findIndexItem(row.ID);
      this.deleteItem(index);
    }
  }
  //Remove item from Array
  deleteItem(index) {
    this.checkedItems.splice(index,1);
  }

  findIndexItem(ID) : any {
    for (let a in this.checkedItems) {
      if (ID === this.checkedItems[a].ID) {
        return a;
      }
    }
  }

  public cellClick(row: any, column: any): void {
    this.cellClicked.emit({ row, column });
  }
  public viewItem(row: any): void {
    this.viewedItem.emit(row);
  }
  public editItem(row: any): void {
    this.editedItem.emit(row);
  }
  public removeItem(row: any): void {
    this.removedItem.emit(row);
  }
  public exportItem(row: any, exportType : any) : void {
    this.exportedItem.emit({row, exportType});
  }
  public testConnection(row: any) : void {
    this.testedConnection.emit(row);
  }
  public extraActionClick(row: any, action: any, property? : any) : void {
    this.extraActionClicked.emit({row , action, property});
  }
}
