import { Component, EventEmitter, Input, Output } from '@angular/core';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';

@Component({
  selector: 'ng-table',
  template: `
    <table class="table dataTable" ngClass="{{config.className || ''}}"
           role="grid" style="width: 100%;">
      <thead>
        <tr role="row">
          <th *ngIf="showCustom == true" style="min-width:78px">
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
          <i class="glyphicon glyphicon-eye-open" (click)="viewItem(row)"></i>
					<i class="glyphicon glyphicon-edit"  (click)="editItem(row)"></i>
    			<i class="glyphicon glyphicon glyphicon-remove"  (click)="removeItem(row)"></i>
          </td>
          <td (click)="cellClick(row, column.name)" *ngFor="let column of columns" [innerHtml]="sanitize(getData(row, column.name))" style="text-align:right"></td>
        </tr>
      </tbody>
    </table>
  `
})
export class NgTableComponent {
  // Table values
  @Input() public rows:Array<any> = [];
  @Input() public showCustom: boolean;

  @Input()
  public set config(conf:any) {
    if (!conf.className) {
      conf.className = 'table-striped table-bordered';
    }
    if (conf.className instanceof Array) {
      conf.className = conf.className.join(' ');
    }
    this._config = conf;
  }

  // Outputs (Events)
  @Output() public tableChanged:EventEmitter<any> = new EventEmitter();
  @Output() public cellClicked:EventEmitter<any> = new EventEmitter();
  @Output() public viewedItem:EventEmitter<any> = new EventEmitter();
  @Output() public editedItem:EventEmitter<any> = new EventEmitter();
  @Output() public removedItem:EventEmitter<any> = new EventEmitter();


  public showFilterRow:Boolean = false;

  @Input()
  public set columns(values:Array<any>) {
    values.forEach((value:any) => {
      if (value.filtering) {
        this.showFilterRow = true;
      }
      if (value.className && value.className instanceof Array) {
        value.className = value.className.join(' ');
      }
      let column = this._columns.find((col:any) => col.name === value.name);
      if (column) {
        Object.assign(column, value);
      }
      if (!column) {
        this._columns.push(value);
      }
    });
  }

  private _columns:Array<any> = [];
  private _config:any = {};

  public constructor(private sanitizer:DomSanitizer) {
  }

  public sanitize(html:string):SafeHtml {
    if ( typeof html === 'object') {
      var test: any = '<ul class="list-unstyled">';
      for (var item of html) {
        if(typeof item === 'object') {
          test += "<li>";
          for (var item2 in Object(item)) {
            if (typeof item[item2] === 'boolean') {
                if (item[item2]) test += ' <i class="glyphicon glyphicon-arrow-right"></i>'
                else test += ' <i class="glyphicon glyphicon-alert"></i>'
            } else test += item[item2];
          }
          test += "</li>";
        } else {
          test += "<li>"+item+"</li>";
        }
      }
      test+="</ul>"
      return test;
    } else if (typeof html === 'boolean') {
        if (html) return '<i class="glyphicon glyphicon-ok"></i>'
        else return '<i class="glyphicon glyphicon-remove"></i>'
    }
    else {
    return this.sanitizer.bypassSecurityTrustHtml(html);
  }
  }

  public get columns():Array<any> {
    return this._columns;
  }

  public get config():any {
    return this._config;
  }

  public get configColumns():any {
    let sortColumns:Array<any> = [];

    this.columns.forEach((column:any) => {
      if (column.sort) {
        sortColumns.push(column);
      }
    });

    return {columns: sortColumns};
  }

  public onChangeTable(column:any):void {
    this._columns.forEach((col:any) => {
      if (col.name !== column.name && col.sort !== false) {
        col.sort = '';
      }
    });
    this.tableChanged.emit({sorting: this.configColumns});
  }

  public getData(row:any, propertyName:string):string {
    return propertyName.split('.').reduce((prev:any, curr:string) => prev[curr], row);
  }

  public cellClick(row:any, column:any):void {
    this.cellClicked.emit({row, column});
  }
  public viewItem(row:any):void {
    this.viewedItem.emit(row);
  }
  public editItem(row:any):void {
    this.editedItem.emit(row);
  }
  public removeItem(row:any):void {
    this.removedItem.emit(row);
  }
}
