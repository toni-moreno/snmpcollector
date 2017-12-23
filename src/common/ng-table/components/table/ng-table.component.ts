import { Component, EventEmitter, Input, Output, Pipe, PipeTransform } from '@angular/core';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { TooltipModule } from 'ngx-bootstrap';
import { ElapsedSecondsPipe } from '../../../elapsedseconds.pipe';


@Component({
  selector: 'ng-table',
  templateUrl: './ng-table.html',
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
  @Input() sanitizeCell: Function;
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

      let output: string
      if (typeof this.sanitizeCell === "function" ) {
        output = this.sanitizeCell(html,transform)
        if ( output.length > 0 ) {
          return output
        }
      }

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
