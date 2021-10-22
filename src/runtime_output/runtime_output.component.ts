
import { ChangeDetectionStrategy, Component, ViewChild, ChangeDetectorRef, OnDestroy } from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';
import { RuntimeOutputService } from './runtime_output.service';
import { ItemsPerPageOptions } from '../common/global-constants';

import { SpinnerComponent } from '../common/spinner';
import { SnmpDeviceService } from '../snmpdevice/snmpdevicecfg.service';

import { TestConnectionModal } from '../common/test-connection-modal';
import { RuntimeComponentConfig, TableRole, ExtraActions } from './runtime_output.data';

declare var _: any;
@Component({
  selector: 'runtime-output',
  providers: [RuntimeOutputService, SnmpDeviceService],
  templateUrl: './runtime_outputview.html',
  styleUrls: ['./runtime_outputeditor.css'],
})


export class RuntimeOutputComponent implements OnDestroy {
  @ViewChild('viewTestConnectionModal') public viewTestConnectionModal: TestConnectionModal;

  itemsPerPageOptions: any = ItemsPerPageOptions;
  public isRefreshing: boolean = true;

  public defaultConfig : any = RuntimeComponentConfig;
  public tableRole : any = TableRole;
  public extraActions: any = ExtraActions;
  public oneAtATime: boolean = true;
  editmode: string; //list , create, modify
  isRequesting: boolean = false;
  runtime_outputs: Array<any>;

  mySubscription: any;
  filter: string;
  measActive: number = 0;
  runtime_dev: any;
  subItem: any;
  islogLevelChanged: boolean = false;
  maxrep: any = '';

  //TABLE
  private data: Array<any> = [];
  public activeOutputs: number;
  public noConnectedDevices: number;
  public dataTable: Array<any> = [];
  public finalData: Array<Array<any>> = [];
  public columns: Array<any> = [];
  public finalColumns: Array<Array<any>> = [];
  public tmpcolumns: Array<any> = [];

  public refreshRuntime: any = {
    'Running': false,
    'LastUpdate': new Date()
  }
  public intervalStatus: any

  public rows: Array<any> = [];
  public page: number = 1;
  public itemsPerPage: number = 20;
  public maxSize: number;
  public numPages: number = 1;
  public length: number = 0;
  public myFilterValue: any;
  public activeFilter: boolean = false;
  public deactiveFilter: boolean = false;
  public noConnectedFilter: boolean = false;

  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.columns },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public runtimeOutputService: RuntimeOutputService, builder: FormBuilder, private ref: ChangeDetectorRef, public snmpDeviceService: SnmpDeviceService) {
    this.editmode = 'list';
    this.reloadData();
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
    console.log("GOT DATA", this.data)
    filteredData.forEach((item: any) => {
      let flag = false;
      this.columns.forEach((column: any) => {
        console.log(item, column.name)
        console.log(item[column.name])
        if (item[column.name] === null) {
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

  changeItemsPerPage(items) {
    if (items) this.itemsPerPage = parseInt(items);
    else this.itemsPerPage = this.length;
    let maxPage = Math.ceil(this.length / this.itemsPerPage);
    if (this.page > maxPage) this.page = maxPage;
    this.onChangeTable(this.config);
  }

  public onChangeTable(config: any, page: any = { page: this.page, itemsPerPage: this.itemsPerPage }): any {
    if (config.filtering) {
      Object.assign(this.config.filtering, config.filtering);
    }
    if (config.sorting) {
      Object.assign(this.config.sorting, config.sorting);
    }
    let filteredData = this.changeFilter(this.data, this.config);
    let sortedData = this.changeSort(filteredData, this.config);
    this.rows = page && this.config.paging ? this.changePage(page, sortedData) : sortedData;
    this.length = sortedData.length;
    this.activeOutputs = sortedData.filter((item) => { return item.Active }).length
    this.noConnectedDevices = sortedData.filter((item) => { if (item.Active === true && item.DeviceConnected === false) return true }).length
  }

  public onExtraActionClicked(data: any) {
    switch (data.action) {
      case 'SetActive':
        this.changeActiveOutput(data.row.ID, !data.row.Active)
        break;
      case 'SetEnqueuePolicy':
          this.changeEnqueuePolicy(data.row.ID, !data.row.EnqueueOnWriteError)
          break;
      case 'FlushBuffer':
        this.flushBuffer(data.row.ID);
        break;
      case 'ResetBuffer':
          this.resetBuffer(data.row.ID);
          break;
      default:
        break;
    }
    console.log(data);
  }

  onResetFilter(): void {
    this.page = 1;
    this.myFilterValue = "";
    this.config.filtering = { filtering: { filterString: '' } };
    this.onChangeTable(this.config);
  }


  tableCellParser (data: any, type: string) {
    if (type === "MULTISTRINGPARSER") {
      var test: any = '<ul class="list-unstyled">';
      for (var i of data) {
          test +="<li>"
          test +="<span class=\"badge\">"+ ( i["IType"] === 'T' ? "Tag" : "Field" ) +"</span><b>"+ i["IName"] + " :</b>" +  i["Value"];
          test += "</li>";
      }
      test += "</ul>"
      return test
    }
    return ""
  }

  showTable(id: number) {
    this.columns = [];
    this.measActive = id;
    this.myFilterValue = "";
    this.config.filtering = { filtering: { filterString: '' } };
    //Reload data and column table
    this.data = this.finalData[id];
    this.columns = this.finalColumns[id];
    //Reload config to enable sort
    this.config.sorting = { columns: this.columns };
    this.onChangeTable(this.config);
  }

  changeActiveOutput(id, event) {
    console.log("ID,event", id, event);

    this.runtimeOutputService.changeOutputActive(id, event)
      .subscribe(
      data => {
        _.forEach(this.runtime_outputs, function (d, key) {
          console.log(d, key)
          if (d.ID == id) {
            d.Active = !d.Active;
            return false
          }
        })
        console.log(this.runtime_outputs);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  changeEnqueuePolicy(id, event) {
    console.log("ID,event", id, event);
    this.runtimeOutputService.changeEnqueuePolicy(id, event)
      .subscribe(
      data => {
        _.forEach(this.runtime_outputs, function (d, key) {
          console.log(d, key)
          if (d.ID == id) {
            d.EnqueueOnWriteError = !d.EnqueueOnWriteError;
            return false
          }
        })
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  showTestConnectionModal(row: any) {
    this.snmpDeviceService.getDevicesById(row.ID)
      .subscribe(data => {
        this.viewTestConnectionModal.show(data);
      },
      err => console.error(err),
    );
  }

  flushBuffer(id) {
    console.log("ID", id);
    this.runtimeOutputService.flushbuffer(id)
      .subscribe(
      data => {
        console.log("buffer flushed done")
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  resetBuffer(id) {
    console.log("ID", id);
    this.runtimeOutputService.resetbuffer(id)
      .subscribe(
      data => {
        console.log("buffer reset done")
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  reloadData() {
    this.itemsPerPage = 20;
    this.isRequesting = true;
    if (this.mySubscription) {
      this.mySubscription.unsubscribe();
    }
    clearInterval(this.intervalStatus);
    this.editmode = "list"
    this.columns = this.defaultConfig['table-columns']
    this.filter = null;
    // now it's a simple subscription to the observable
    this.mySubscription = this.runtimeOutputService.getOutputRuntime(null)
      .subscribe(
      data => {
        this.mySubscription = null;
        this.runtime_outputs = data
        this.data = this.runtime_outputs;
        this.config.sorting.columns = this.columns,
          this.isRequesting = false;
        this.activeFilter = this.deactiveFilter = this.noConnectedFilter = false;
        this.onChangeTable(this.config);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  toogleActiveFilter(option: string) {
    if (this.activeFilter === false && option === 'active') {
      this.noConnectedFilter = false;
      this.deactiveFilter = false;
      this.activeFilter = true;
      this.data = this.runtime_outputs.filter((item) => { if (item.Active === true) return true })
    } else if (this.deactiveFilter === false && option === 'deactive') {
      this.noConnectedFilter = false;
      this.activeFilter = false;
      this.deactiveFilter = true;
      this.data = this.runtime_outputs.filter((item) => { if (item.Active === false) return true })
    } else if (this.noConnectedFilter === false && option === 'noconnected') {
      this.noConnectedFilter = true;
      this.activeFilter = false;
      this.deactiveFilter = false;
      this.data = this.runtime_outputs.filter((item) => { if (item.DeviceConnected === false && item.Active === true) return true })
    } else {
      this.data = this.runtime_outputs;
      this.noConnectedFilter = false;
      this.activeFilter = false;
      this.deactiveFilter = false;
    }
    this.onChangeTable(this.config);
  }

  ngOnDestroy() {
    clearInterval(this.intervalStatus);
    if (this.mySubscription) this.mySubscription.unsubscribe();
  }

}
