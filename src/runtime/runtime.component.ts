
import { ChangeDetectionStrategy, Component, ViewChild, ChangeDetectorRef, OnDestroy } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { RuntimeService } from './runtime.service';
import { ItemsPerPageOptions } from '../common/global-constants';

import { SpinnerComponent } from '../common/spinner';
import { SnmpDeviceService } from '../snmpdevice/snmpdevicecfg.service';

import { TestConnectionModal } from '../common/test-connection-modal';

declare var _:any;
@Component({
  selector: 'runtime',
  providers: [RuntimeService,SnmpDeviceService],
  templateUrl: './runtimeview.html',
  styleUrls: ['./runtimeeditor.css'],
})


export class RuntimeComponent implements OnDestroy {
  @ViewChild('viewTestConnectionModal') public viewTestConnectionModal: TestConnectionModal;

  itemsPerPageOptions : any = ItemsPerPageOptions;
  public isRefreshing: boolean = true;

  public oneAtATime: boolean = true;
  editmode: string; //list , create, modify
  dataArray : any = [];
  dataArray2 : any = [];
  isRequesting : boolean = false;
  runtime_devs: Array<any>;
  runtime_devs2: Array<any>;

  public extraActions: Array<any> =  [
    { 'title': 'SetActive', 'type' : 'boolean', 'content': {'enabled': '<i class="glyphicon glyphicon-pause"></i>', 'disabled' : '<i class="glyphicon glyphicon-play"></i>'}, 'property': 'DeviceActive'},
    { 'title': 'SnmpReset', 'type' : 'button', 'content': {'enabled': 'Reset'}}
  ]

  public rt_columns: Array<any> = [
    { title: 'ID', name: 'ID'},
    { title: 'TagMap', name: 'TagMap',tooltip: 'Num Measurements configured'},
    { title: 'SysDesc', name: 'SysDescription'},
    { title: '#Meas', name: 'NumMeasurements',tooltip: 'Num Measurements configured'},
    { title: '#Metrics', name: 'NumMetrics'},
    { title: 'Get.Errs' ,name:'Counter7',tooltip: 'SnmpOIDGetErrors:number of  oid with errors for all measurements '},
    { title: 'M.Errs',name:'Counter14',tooltip: 'MeasurementSentErrors: number of measuremenets  formatted with errors '},
    { title: 'G.Time',name:'Counter16',tooltip: 'CicleGatherDuration time: elapsed time taken to get all measurement info', transform : 'elapsedseconds'},
    { title: 'F.Time',name:'Counter18',tooltip: 'CicleGatherDuration time: elapsed time taken to compute all applicable filters on the device', transform : 'elapsedseconds'}
  ];
  mySubscription : any;
  filter: string;
  measActive: number = 0;
  runtime_dev: any;
  subItem: any;
  islogLevelChanged: boolean = false;
  newLogLevel: string = null;
  loglLevelArray: Array<string> = [
    'panic',
    'fatal',
    'error',
    'warning',
    'info',
    'debug'
  ];
  maxrep: any = '';
  counterDef: CounterType[] = [
  /*0*/    { "show":false, "id": "SnmpGetQueries","label": "SnmpGet Queries","type":"counter", "tooltip": "number of snmp queries"},
  /*1*/    { "show":false, "id": "SnmpWalkQueries","label": "SnmpWalk Queries","type":"counter", "tooltip": "number of snmp walks"},
  /*2*/    { "show":false, "id": "SnmpGetErrors","label": "SnmpGet Errors","type":"counter", "tooltip": "Get Error"},
  /*3*/    { "show":false, "id": "SnmpWalkErrors","label": "SnmpWalk Errors","type":"counter", "tooltip": "Walk Error"},
  /*4*/    { "show":false, "id": "SnmpQueryTimeouts","label": "Snmp Errors by Timeout","type":"counter", "tooltip": "number of registered errors by timeouts"},
  /*5*/    { "show":true, "id": "SnmpOIDGetAll","label": "OID Gets ALL","type":"counter", "tooltip": "All Gathered snmp metrics ( sum of snmpget oid's and all received oid's in snmpwalk queries)"},
  /*6*/    { "show":true, "id": "SnmpOIDGetProcessed","label": "OID Processed","type":"counter", "tooltip": "Gathered and processed snmp metrics after filters are applied ( not always sent to the backend it depens on the report flag)"},
  /*7*/    { "show":true, "id": "SnmpOIDGetErrors","label": "OID With Errors","type":"counter", "tooltip": "number of  oid with errors for all measurements"},
  /*8*/    { "show":false, "id": "EvalMetricsAll","label": "Evaluated Metrics","type":"counter", "tooltip": "number of evaluated metrics"},
  /*9*/    { "show":false, "id": "EvalMetricsOk","label": "Evaluated Metrics","type":"counter", "tooltip": "number of evaluated metrics without errors."},
  /*10*/    { "show":false, "id": "EvalMetricsErrors","label": "Evaluated Metrics","type":"counter", "tooltip": "number of evaluated metrics with some kind of error."},
  /*11*/    { "show":true, "id": "MetricSent","label": "Metric Sent","type":"counter", "tooltip": "number of metrics sent (taken as fields) for all measurements"},
  /*12*/    { "show":true, "id": "MetricSentErrors","label": "Metric Sent Errors","type":"counter", "tooltip": "number of metrics  (taken as fields) with errors forall measurements"},
  /*13*/    { "show":true, "id": "MeasurementSent","label": "Measurement sent","type":"counter", "tooltip": "(number of  measurements build to send as a sigle request sent to the backend)"},
  /*14*/    { "show":true, "id": "MeasurementSentErrors","label": "Measurement sent Errors","type":"counter", "tooltip": "(number of measuremenets  formatted with errors )"},
  /*15*/    { "show":false, "id": "CicleGatherStartTime","label": "Cicle Gather Start Time","type":"time", "tooltip": "Last gather time "},
  /*16*/    { "show":true, "id": "CicleGatherDuration","label": "Cicle Gather Duration","type":"duration", "tooltip": "elapsed time taken to get all measurement info"},
  /*17*/    { "show":false, "id": "FilterStartTime","label": "Filter update Start Time","type":"time", "tooltip": "Last Filter time"},
  /*18*/    { "show":true, "id": "FilterDuration","label": "Filter update Duration","type":"duration", "tooltip": "elapsed time taken to compute all applicable filters on the device"},
  /*19*/    { "show":false, "id": "BackEndSentStartTime","label": "BackEnd (influxdb) Sent Start Time","type":"time", "tooltip": "Last sent time"},
  /*20*/    { "show":true, "id": "BackEndSentDuration","label": "BackEnd (influxdb) Sent Duration","type":"duration", "tooltip": "elapsed time taken to send data to the db backend"},
  ];

  isObject(val) { return typeof val === 'object'; }
  isArray(val) { return typeof val === 'array' }

  //TABLE
  // CHECK DATA, our data is array of arrays?!
  private data: Array<any> = [];
  public activeDevices : number;
  public noConnectedDevices : number;
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
  public noConnectedFilter  : boolean = false;

  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.columns },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public runtimeService: RuntimeService, builder: FormBuilder, private ref: ChangeDetectorRef, public snmpDeviceService: SnmpDeviceService) {
    this.editmode = 'list';
    this.reloadData();
  }


  public changePage(page: any, data: Array<any> = this.data): Array<any> {
    //Check if we have to change the actual page
    let maxPage =  Math.ceil(data.length/this.itemsPerPage);
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

  changeItemsPerPage (items) {
    if (items) this.itemsPerPage = parseInt(items);
    else this.itemsPerPage=this.length;
    let maxPage =  Math.ceil(this.length/this.itemsPerPage);
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
    this.activeDevices = sortedData.filter((item) => {return item.DeviceActive}).length
    this.noConnectedDevices = sortedData.filter((item) => { if(item.DeviceActive === true && item.DeviceConnected  === false) return true }).length
  }

  public onExtraActionClicked(data:any) {
    switch (data.action) {
      case 'SetActive' :
      this.changeActiveDevice(data.row.ID, !data.row.DeviceActive)
      break;
      case 'SnmpReset' :
      this.forceSnmpReset(data.row.ID);
      break;
      default:
      break;
    }
    console.log(data);
  }

  onResetFilter() : void {
    this.page = 1;
    this.myFilterValue = "";
    this.config.filtering = {filtering: { filterString: '' }};
    this.onChangeTable(this.config);
  }

  initRuntimeInfo(id: string, meas: number, isRequesting?: boolean) {
    //Reset params
    this.editmode = 'view';
    if (isRequesting) {
      this.isRequesting = isRequesting;
      this.runtime_dev = null;
    }
    this.isRefreshing = false;
    this.refreshRuntime.Running = false;
    clearInterval(this.intervalStatus);
    this.measActive = meas || 0;
    if (!this.mySubscription) {
      this.loadRuntimeById(id, this.measActive);
    }
  }

  updateRuntimeInfo(id: string, selectedMeas: number, status: boolean) {
    clearInterval(this.intervalStatus);
    this.refreshRuntime.Running = status;
    this.measActive = selectedMeas || this.measActive;
    if (this.refreshRuntime.Running) {
      this.isRefreshing = true;
      this.refreshRuntime.LastUpdate = new Date();
      //Cargamos interval y dejamos actualizando la información:
      this.intervalStatus = setInterval(() => {
        this.isRefreshing = false;
        setTimeout(() => {
          this.isRefreshing = true;
        }, 2000);
        this.refreshRuntime.LastUpdate = new Date();
        this.loadRuntimeById(id, this.measActive);
        this.ref.markForCheck();
      }, Math.max(5000,this.runtime_dev['Freq'] * 1000)); //lowest update rate set to 5 sec
    } else {
      this.isRefreshing = false;
      clearInterval(this.intervalStatus);
    }
  }

  loadRuntimeById(id: string, selectedMeas: number) {
    this.mySubscription = this.runtimeService.getRuntimeById(id)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.mySubscription = null;
        this.finalColumns = [];
        this.finalData = [];
        this.runtime_dev = data;
        this.runtime_dev.ID = id;
        //Generate Columns
        if (data['Measurements'] && data['Measurements'].length > 0) {
          for (let measKey of data['Measurements']) {
            console.log('measKey: ',measKey)
            //Generate the Coluns array, go over it only once using break
            //Save it as array of arrays on finalColumns
            if (Object.keys(measKey['MetricTable']['Header']).length !== 0) {
              this.tmpcolumns = [];
              this.tmpcolumns.push({ title: 'Index', name: 'Index' });
              for (let fieldName in measKey['MetricTable']['Header']) {
                let mode = measKey['MetricTable']['Header'][fieldName]
                let micon = ''
                let tt = ''
                 switch(mode) {
                   case 0: //never send
                    micon = 'remove-circle text-danger'
                    tt = 'this field won\'t be sent'
                   break;
                   case 1: //always send
                    micon = 'ok-circle text-success'
                    tt = 'this always will send'
                   break;
                   case 2: //send i non zero
                    micon = 'ban-circle text-warning'
                    tt = 'this field only will be sent if non zero'
                   break

                 }
                let tmpColumn: any = { title: fieldName, name: fieldName , icon: micon, tooltip: tt }
                this.tmpcolumns.push(tmpColumn);
              }
            this.finalColumns.push(this.tmpcolumns);
            } else {
              this.finalColumns.push([]);
            }
            //Go over the array again and get the DATA
            //indexKey contains the Index, must generate the same on multiples arrays
            for (let rowid in measKey['MetricTable']['Row']) {
              let row = measKey['MetricTable']['Row'][rowid]
              let tmpTable: any = { tooltipInfo: {} };
              tmpTable['valid']=row['Valid'];
              tmpTable.Index = rowid;
              for (let metricid in row['Data'] ) {
                let metric = row['Data'][metricid]
                let fieldName = metric.FieldName
                tmpTable[fieldName] = metric.CookedValue;
                tmpTable['tooltipInfo'][fieldName] = metric;
              }
              this.dataTable.push(tmpTable);
            }
            this.finalData.push(this.dataTable);
            this.dataTable = [];
          }
          this.showTable(selectedMeas ? this.measActive : 0);
        }
        if (!this.refreshRuntime.Running) this.updateRuntimeInfo(id, this.measActive, true);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  showTable(id: number) {
    this.columns = [];
    this.measActive = id;
    this.myFilterValue = "";
    this.config.filtering = {filtering: { filterString: '' }};
    //Reload data and column table
    this.data = this.finalData[id];
    this.columns = this.finalColumns[id];
    //Reload config to enable sort
    this.config.sorting = { columns: this.columns };
    this.onChangeTable(this.config);
  }

  changeActiveDevice(id, event) {
    console.log("ID,event", id, event);

    this.runtimeService.changeDeviceActive(id, event)
      .subscribe(
      data => {
        _.forEach(this.runtime_devs,function(d,key){
          console.log(d,key)
          if (d.ID == id) {
            d.DeviceActive = !d.DeviceActive;
            return false
          }
        })
        console.log(this.runtime_devs);
        if (this.runtime_dev != null) {
          this.runtime_dev.DeviceActive = !this.runtime_dev.DeviceActive;
        }

      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  changeStateDebug(id, event) {
    console.log("ID,event", id, event);
    this.runtimeService.changeStateDebug(id, event)
      .subscribe(
      data => {
        this.runtime_dev.StateDebug = !this.runtime_dev.StateDebug;
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  onChangeLogLevel(level) {
    this.islogLevelChanged = true;
    this.newLogLevel = level;
  }

  changeLogLevel(id) {
    console.log("ID,event");
    this.runtimeService.changeLogLevel(id, this.newLogLevel)
      .subscribe(
      data => {
        this.runtime_dev.CurLogLevel = this.newLogLevel;
        this.islogLevelChanged = false;
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  downloadLogFile(id) {
    console.log("Download Log file from device", id);
    this.runtimeService.downloadLogFile(id)
      .subscribe(
      data => {
        saveAs(data, id + ".log")
        console.log("download done")
      },
      err => {
        console.error(err)
        console.log("Error downloading the file.")
      },
      () => console.log('Completed file download.')
      );
  }

  showTestConnectionModal(row : any) {
    this.snmpDeviceService.getDevicesById(row.ID)
      .subscribe(data => {
        this.viewTestConnectionModal.show(data);
      },
      err => console.error(err),
    );
  }

  forceFltUpdate(id) {
    console.log("ID,event", id, event);
    this.runtimeService.forceFltUpdate(id)
      .subscribe(
      data => {
        console.log("filter update done")
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  forceSnmpReset(id) {
    console.log("ID,event", id, event);
    this.runtimeService.forceSnmpReset(id)
      .subscribe(
      data => {
        console.log("reset done")
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }


  setSnmpMaxRepetitions(id) {
    if (this.maxrep.toString().match(/^([1-9]+\d*)$/) && this.maxrep < 256) {
        console.log("Set Snmp Max Repeticions : "+ this.maxrep +" On device: "+id)
        this.runtimeService.setSnmpMaxRepetitions(id,this.maxrep)
          .subscribe(
          data => {
            console.log("set max repetitions done")
          },
          err => console.error(err),
          () => console.log('DONE')
          );
    } else {
      alert(this.maxrep+' is not an unsigned 8 bits integer')
    }
    console.log("ID,event", id, event, this.maxrep);
  }

  reloadData() {
    this.itemsPerPage=20;
    this.isRequesting = true;
    if (this.mySubscription) {
      this.mySubscription.unsubscribe();
    }
    clearInterval(this.intervalStatus);
    this.editmode = "list"
    this.columns = this.rt_columns
    this.filter = null;
    // now it's a simple subscription to the observable
    this.mySubscription = this.runtimeService.getRuntime(null)
      .subscribe(
      data => {
        this.mySubscription = null;
        this.runtime_devs = data
        this.data = this.runtime_devs;
        this.config.sorting.columns=this.columns,
        this.isRequesting = false;
        this.activeFilter = this.deactiveFilter = this.noConnectedFilter = false;
        this.onChangeTable(this.config);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  toogleActiveFilter(option : string) {
    if (this.activeFilter === false && option === 'active') {
      this.noConnectedFilter = false;
      this.deactiveFilter = false;
      this.activeFilter = true;
      this.data = this.runtime_devs.filter((item) => { if(item.DeviceActive === true) return true })
    } else if (this.deactiveFilter === false && option === 'deactive' ) {
      this.noConnectedFilter = false;
      this.activeFilter = false;
      this.deactiveFilter = true;
      this.data = this.runtime_devs.filter((item) => { if(item.DeviceActive === false) return true })
    } else if (this.noConnectedFilter === false && option === 'noconnected'){
      this.noConnectedFilter = true;
      this.activeFilter = false;
      this.deactiveFilter = false;
      this.data = this.runtime_devs.filter((item) => { if(item.DeviceConnected === false) return true })
    } else {
      this.data = this.runtime_devs;
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
