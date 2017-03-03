import { Component, EventEmitter, Input, Output, Pipe, PipeTransform } from '@angular/core';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { TooltipModule } from 'ng2-bootstrap';


@Component({
  selector: 'ng-table',
  template: `
    <table class="table dataTable" ngClass="{{config.className || ''}}"
           role="grid" style="width: 100%;">
      <thead>
        <tr role="row">
          <th *ngIf="showCustom == true" [ngStyle]="exportType ? {'min-width': '95px'} : {'min-width': '80px'}">
          </th>
          <th *ngFor="let column of columns" [ngTableSorting]="config" [column]="column"
              (sortChanged)="onChangeTable($event)" ngClass="{{column.className || ''}}" style="vertical-align: middle; text-align: center; width:auto !important;">
            {{column.title}}
            <i *ngIf="config && column.sort" class="pull-right glyphicon"
              [ngClass]="{'glyphicon-chevron-down': column.sort === 'desc', 'glyphicon-chevron-up': column.sort === 'asc'}"></i>
          </th>
        </tr>
      </thead>
      <tbody>
      <tr *ngIf="showFilterRow">
      <td *ngIf="showCustom == true" style="width:100px">
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
          <td *ngIf="showCustom == true">
          <i *ngIf="exportType" class="glyphicon glyphicon-download-alt" [tooltip]="'Export item'" (click)="exportItem(row, exportType)"></i>
          <i class="glyphicon glyphicon-eye-open" [tooltip]="'View item'" (click)="viewItem(row)"></i>
					<i class="glyphicon glyphicon-edit"  [tooltip]="'Edit item'" (click)="editItem(row)"></i>
    			<i class="glyphicon glyphicon glyphicon-remove"  [tooltip]="'Remove Item'" (click)="removeItem(row)"></i>
          </td>
          <td (click)="cellClick(row, column.name)" *ngFor="let column of columns; let i = index" container=body [tooltip]="row.tooltipInfo ? tooltipValues : (column.name === 'ID' ? row.Description : '')" [innerHtml]="sanitize(getData(row, column.name))" style="text-align:right">
          <template #tooltipValues>
            <h6>Index:{{row.Index }}</h6>
            <h6>Metric:{{column.name}}</h6>
            <hr>
            <div *ngFor="let test of (getData(row.tooltipInfo, column.name) | keyParser)" style="text-align:left !important">
            <span class="text-left" style="paddint-left: 10px"><b>{{test.key}}</b></span>
            <span class="text-right" *ngIf="test.key === 'CurTime'"> {{test.value | date:'d/M/y H:m:s'}}</span>
            <span class="text-right" *ngIf="test.key === 'LastTime'"> {{test.value | date:'d/M/y H:m:s'}}</span>
            <span class="text-right" *ngIf="test.key !== 'CurTime' && test.key !== 'LastTime'">{{test.value}} </span>
            </div>
          </template>

          </td>
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
  @Input() public showCustom: boolean;
  @Input() public exportType: string;

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

  public sanitize(html: string): SafeHtml {
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

  public getData(row: any, propertyName: string): string {
    return propertyName.split('.').reduce((prev: any, curr: string) => prev[curr], row);
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
}
