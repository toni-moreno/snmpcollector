import { Component, Input, Output, Pipe, PipeTransform, ViewChild, EventEmitter } from '@angular/core';
import { ModalDirective } from 'ng2-bootstrap';
import { Validators, FormGroup, FormControl, FormArray, FormBuilder } from '@angular/forms';
import { ExportServiceCfg } from './export.service'
import { TreeView} from './treeview';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../multiselect-dropdown';

//Services
import { InfluxServerService } from '../../influxserver/influxservercfg.service';
import { SnmpDeviceService } from '../../snmpdevice/snmpdevicecfg.service';
import { InfluxMeasService } from '../../influxmeas/influxmeascfg.service';
import { OidConditionService } from '../../oidcondition/oidconditioncfg.service';
import { SnmpMetricService } from '../../snmpmetric/snmpmetriccfg.service';
import { MeasGroupService } from '../../measgroup/measgroupcfg.service';
import { MeasFilterService } from '../../measfilter/measfiltercfg.service';
import { CustomFilterService } from '../../customfilter/customfilter.service';
import { VarCatalogService } from '../../varcatalog/varcatalogcfg.service';

@Component({
  selector: 'export-file-modal',
  template: `
      <div bsModal #childModal="bs-modal" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="myLargeModalLabel" aria-hidden="true">
          <div class="modal-dialog" style="width:90%">
            <div class="modal-content" >
              <div class="modal-header">
                <button type="button" class="close" (click)="childModal.hide()" aria-label="Close">
                  <span aria-hidden="true">&times;</span>
                </button>
                <h4 class="modal-title" *ngIf="exportObject != null">{{titleName}} <b>{{ exportObject.ID }}</b> - <label [ngClass]="['label label-'+colorsObject[exportType]]">{{exportType}}</label></h4>
                <h4 class="modal-title" *ngIf="exportObject == null">Export</h4>
              </div>
              <div class="modal-body">

              <div *ngIf="prepareExport === false">
              <div class="row">
              <div class="col-md-2">
              <div class="panel-heading">
              1.Select type:
              </div>
                <div class="panel panel-default" *ngFor="let items of objectTypes; let i = index" style="margin-bottom: 0px" >
                  <div class="panel-heading" (click)="loadSelection(i, items.Type)" role="button">
                  <i [ngClass]="selectedType ? (selectedType.Type === items.Type ? ['glyphicon glyphicon-eye-open'] : ['glyphicon glyphicon-eye-close'] ) : ['glyphicon glyphicon-eye-close']"  style="padding-right: 10px"></i>
                  <label [ngClass]="['label label-'+items.Class]"> {{items.Type}}  </label>
                  </div>
                </div>
                </div>
                <div class="col-md-5">
                <div *ngIf="selectedType">
                  <div class="panel-heading">
                    <div>
                      2. Select Items of type <label [ngClass]="['label label-'+selectedType.Class]"> {{selectedType.Type}}</label>  <span class="badge" style="margin-left: 10px">{{resultArray.length}} Results</span>
                    </div>
                    <div dropdown class="text-left" style="margin-top: 10px">
                    <span class="dropdown-toggle-split">Filter by</span>
                    <ss-multiselect-dropdown style="border: none" class="text-primary" [options]="listFilterProp" [texts]="myTexts" [settings]="propSettings" [(ngModel)]="selectedFilterProp" (ngModelChange)="onChange(filter)"></ss-multiselect-dropdown>
                      <input type=text [(ngModel)]="filter" placeholder="Filter items..." (ngModelChange)="onChange($event)">
                      <label [tooltip]="'Clear Filter'" container="body" (click)="filter=''; onChange(filter)"><i class="glyphicon glyphicon-trash text-primary"></i></label>
                    </div>
                    <div class="text-right">
                    <label class="label label-success" (click)=selectAllItems(true)>Select All</label>
                    <label class="label label-danger" (click)=selectAllItems(false)>Deselect All</label>
                    </div>
                  </div>
                  <div style="max-height: 400px; overflow-y:auto">
                    <div *ngFor="let res of resultArray;  let i = index" style="margin-bottom: 0px" >
                      <treeview [keyProperty]="selectedFilterProp" [showType]="false" [visible]="false" [title]="res.ID" [object]="res" [alreadySelected]="checkItems(res.ID, selectedType.Type)" [type]="selectedType.Type" [visibleToogleEnable]="true" [addClickEnable]="true" (addClicked)="selectItem($event,index)"  style="margin-bottom:0px !important">{{res}}</treeview>
                    </div>
                  </div>
                  </div>
                  </div>
                  <div class="col-md-5">
                  <div *ngIf="finalArray.length !== 0">
                    <div class="panel-heading"> 3. Items ready to export: <span class="badge">{{finalArray.length}}</span>
                    <div class="text-right">
                      <label class="label label-danger" (click)="finalArray = []">Clear All</label>
                    </div>
                    </div>
                    <div style="max-height: 400px; overflow-y:auto">
                      <div *ngFor="let res of finalArray;  let i = index" class="col-md-12">
                      <i class="text-danger glyphicon glyphicon-remove-sign col-md-1" role="button" style="margin-top: 15px;" (click)="removeItem(i)"> </i>
                        <treeview [visible]="false" [title]="res.ObjectID" [object]="res" [type]="res.ObjectTypeID" [recursiveToogle]="true" [index] = "i" (recursiveClicked)="toogleRecursive($event)" class="col-md-11">{{res}}</treeview>
                      </div>
                    </div>
                    </div>
                    </div>
                </div>
                </div>
              <div *ngIf="prepareExport === true">
              <div *ngIf="exportResult === true" style="overflow-y: scroll; max-height: 350px">
                <h4 class="text-success"> <i class="glyphicon glyphicon-ok-circle" style="padding-right:10px"></i>Succesfully exported {{exportedItem.Objects.length}} items </h4>
                <div *ngFor="let object of exportedItem.Objects; let i=index">
                  <treeview [visible]="false" [visibleToogleEnable]="true" [title]="object.ObjectID" [type]="object.ObjectTypeID" [object]="object.ObjectCfg"> </treeview>
              </div>
              </div>
              <div  *ngIf="exportForm && exportResult === false">
              <form class="form-horizontal" *ngIf="bulkExport === false">
              <div class="form-group">
                <label class="col-sm-2 col-offset-sm-2" for="Recursive">Recursive</label>
                <i placement="top" style="float: left" class="info control-label glyphicon glyphicon-info-sign" tooltipAnimation="true" tooltip="Select if the export request will include all related componentes"></i>
                <div class="col-sm-9">
                <select name="recursiveObject" class="form-control" id="recursiveObject" [(ngModel)]="recursiveObject">
                  <option value="true">True</option>
                  <option value="false">False</option>
                </select>
                </div>
              </div>
              </form>
              <form [formGroup]="exportForm" class="form-horizontal"  >
                    <div class="form-group">
                      <label for="FileName" class="col-sm-2 col-offset-sm-2 control-FileName">FileName</label>
                      <i placement="top" style="float: left" class="info control-label glyphicon glyphicon-info-sign" tooltipAnimation="true" tooltip="Desired file name"></i>
                      <div class="col-sm-9">
                      <input type="text" class="form-control" placeholder="file.json" formControlName="FileName" id="FileName">
                      </div>
                    </div>
                    <div class="form-group">
                      <label for="Author" class="col-sm-2 control-Author">Author</label>
                      <i placement="top" style="float: left" class="info control-label glyphicon glyphicon-info-sign" tooltipAnimation="true" tooltip="Author of the export"></i>
                      <div class="col-sm-9">
                      <input type="text" class="form-control" placeholder="snmpcollector" formControlName="Author" id="Author">
                      </div>
                    </div>
                    <div class="form-group">
                      <label for="Tags" class="col-sm-2 control-Tags">Tags</label>
                      <i placement="top" style="float: left" class="info control-label glyphicon glyphicon-info-sign" tooltipAnimation="true" tooltip="Related tags to identify exported data"></i>
                      <div class="col-sm-9">
                      <input type="text" class="form-control" placeholder="cisco,catalyst,..." formControlName="Tags" id="Tags">
                      </div>
                    </div>

                    <div class="form-group">
                      <label for="FileName" class="col-sm-2 control-FileName">Description</label>
                      <i placement="top" style="float: left" class="info control-label glyphicon glyphicon-info-sign" tooltipAnimation="true" tooltip="Description of the exported file"></i>
                      <div class="col-sm-9">
                      <textarea class="form-control" style="width: 100%" rows="2" formControlName="Description" id="Description"> </textarea>
                      </div>
                    </div>
              </form>
              </div>
              </div>
              </div>
              <div class="modal-footer" *ngIf="showValidation === true">
               <button type="button" class="btn btn-default" (click)="childModal.hide()">Close</button>
               <button *ngIf="exportResult === false && prepareExport === true" type="button" class="btn btn-primary" (click)="exportBulkItem()">{{textValidation ? textValidation : Save}}</button>
               <button *ngIf="prepareExport === false" type="button" class="btn btn-primary" [disabled]="finalArray.length === 0" (click)="showExportForm()">Continue</button>
             </div>
            </div>
          </div>
        </div>`,
        styleUrls: ['./import-modal-styles.css'],
        providers: [ExportServiceCfg, InfluxServerService, SnmpDeviceService, SnmpMetricService, InfluxMeasService, OidConditionService,MeasGroupService, MeasFilterService, CustomFilterService, VarCatalogService, TreeView]
})

export class ExportFileModal {
  @ViewChild('childModal') public childModal: ModalDirective;
  @Input() titleName: any;
  @Input() customMessage: string;
  @Input() showValidation: boolean;
  @Input() textValidation: string;
  @Input() prepareExport : boolean = true;
  @Input() bulkExport: boolean = false;

  @Output() public validationClicked: EventEmitter<any> = new EventEmitter();

  public validationClick(myId: string): void {
    this.validationClicked.emit(myId);
  }

  public builder: any;
  public exportForm: any;

  constructor(builder: FormBuilder, public exportServiceCfg : ExportServiceCfg,
    public influxServerService: InfluxServerService, public metricMeasService: SnmpMetricService,
    public influxMeasService: InfluxMeasService, public oidConditionService : OidConditionService,
    public snmpDeviceService: SnmpDeviceService, public measGroupService: MeasGroupService,
    public measFilterService: MeasFilterService, public customFilterService: CustomFilterService,
    public varCatalogService: VarCatalogService) {

    this.builder = builder;
  }

//COMMON
  createStaticForm() {
    this.exportForm = this.builder.group({
      FileName: [this.prepareExport ? this.exportObject.ID+'_'+this.exportType+'_'+this.nowDate+'.json' : 'bulkexport_'+this.nowDate+'.json' , Validators.required],
      Author: ['snmpcollector', Validators.required],
      Tags: [''],
      Recursive: [true, Validators.required],
      Description: ['Autogenerated', Validators.required]
    });
  }

  //Single Object Export:
  exportObject: any = null;
  exportType: any = null;

   //Single Object
  public colorsObject : Object = {
   "snmpdevicecfg" : 'danger',
   "influxcfg" : 'info',
   "measfiltercfg": 'warning',
   "oidconditioncfg" : 'success',
   "customfiltercfg" : 'default',
   "measurementcfg" : 'primary',
   "snmpmetriccfg" : 'warning',
   "measgroupcfg" : 'success',
   "varcatalogcfg" : 'default'
   };

  //Control to load exported result
  exportResult : boolean = false;
  exportedItem : any;

  //Others
  nowDate : any;
  recursiveObject : boolean = true;

  //Bulk Export - Result Array from Loading data:

  resultArray : any = [];
  dataArray : any = [];
  selectedFilterProp : any = [];
  listFilterProp: IMultiSelectOption[] = [];
  private propSettings: IMultiSelectSettings = {
      singleSelect: true,
  };

  //Bulk Export - SelectedType
  selectedType : any = null;
  finalArray : any = [];
  filter : any = "";
  //Bulk Objects
  public objectTypes : any = [
   {'Type':"snmpdevicecfg", 'Class' : 'danger', 'Visible': false},
   {'Type':"influxcfg" ,'Class' : 'info', 'Visible': false},
   {'Type':"measfiltercfg", 'Class' : 'warning','Visible': false},
   {'Type':"oidconditioncfg", 'Class' : 'success', 'Visible': false},
   {'Type':"customfiltercfg", 'Class' : 'default', 'Visible': false},
   {'Type':"measurementcfg", 'Class' : 'primary', 'Visible': false},
   {'Type':"snmpmetriccfg", 'Class' : 'warning', 'Visible': false},
   {'Type':"measgroupcfg", 'Class' : 'success', 'Visible': false},
   {'Type':"varcatalogcfg", 'Class' : 'default', 'Visible': false}
   ]

   //Reset Vars on Init
  clearVars() {
    this.finalArray = [];
    this.resultArray = [];
    this.selectedType = null;
    this.exportResult = false;
    this.exportedItem = [];
    this.exportType = null;
    this.exportObject = null;
  }

  //Init Modal, depending from where is called
  initExportModal(exportObject: any, prepareExport? : boolean) {
    this.clearVars();
    if (prepareExport === false) {
      this.prepareExport = false;
    } else {
      this.prepareExport = true;
    };
    //Single export
    if (this.prepareExport === true) {
      this.exportObject = exportObject.row;
      this.exportType = exportObject.exportType;
      //Sets the FinalArray to export the items, in this case only be 1
      this.finalArray = [{
        'ObjectID' : this.exportObject.ID,
        'ObjectTypeID' :  this.exportType,
        'Options' : {
          Recursive: this.recursiveObject
        }
      }]
    //Bulk export
  } else {
    this.exportObject = exportObject;
    this.exportType = null;
  }
   
    this.nowDate = this.getCustomTimeString()
    this.createStaticForm();
    this.childModal.show();
  }

  getCustomTimeString(){
    let date  = new Date();
    let day = ('0' + date.getDate()).slice(-2);
    let year  = date.getFullYear().toString();
    let month  = ('0' + (date.getMonth()+1)).slice(-2);
    let ymd =year+month+day;
    let hm  =  date.getHours().toString()+date.getMinutes().toString();
    return ymd + '_' + hm;
  }

  onChange(event){
    let tmpArray = this.dataArray.filter((item: any) => {
      if (item[this.selectedFilterProp]) return item[this.selectedFilterProp].toString().match(event);
      else if (event === "" && !item[this.selectedFilterProp]) return item;
    });
    this.resultArray = tmpArray;
  }
  changeFilterProp(prop){
    this.selectedFilterProp = prop;
  }

  //Load items from selection type
   loadSelection(i, type) {
     for (let a of this.objectTypes) {
       if(type !== this.objectTypes[i].Type) {
         this.objectTypes[i].Visible = false;
       }
     }
     this.objectTypes[i].Visible = true;
     this.selectedType = this.objectTypes[i];
     this.filter = "";
     this.loadItems(type,null);
   }

   checkItems(checkItem: any,type) : boolean {
     //Extract the ID from finalArray and loaded Items:
     let exist = true;
     for (let a of this.finalArray) {
       if (checkItem === a.ObjectID) {
         exist = false;
       }
     }
     return exist;
   }

   //Common function to find given object property inside an array
   findIndexItem(checkArray, checkItem: any) : any {
     for (let a in checkArray) {
       if (checkItem === checkArray[a].ObjectID) {
         return a;
       }
     }
   }

   selectAllItems(selectAll) {
     //Creates the form array
     if (selectAll === true) {
       for (let a of this.resultArray) {
         if (this.checkItems(a.ID, this.selectedType)) {
           this.finalArray.push({ "ObjectID" : a.ID, ObjectTypeID: this.selectedType.Type, "Options" : {'Recursive': false }});
         }
       }
     } else {
       for (let a of this.resultArray) {
         let index = this.findIndexItem(this.finalArray, a.ID);
           if (index) this.removeItem(index);
       }
     }
   }

   //Select item to add it to the FinalArray or delete it if its alreay selected
  selectItem(event) {
    if (this.checkItems(event.ObjectID, event.ObjectTypeID)) {
      this.finalArray.push(event);
    }
    else {
      let index = this.findIndexItem(this.finalArray, event.ObjectID);
      this.removeItem(index);
    }
  }
  //Remove item from Array
  removeItem(index) {
    this.finalArray.splice(index,1);
  }

  //Change Recursive option on the FinalArray objects
  toogleRecursive(event) {
    this.finalArray[event.Index].Options.Recursive = event.Recursive;
  }

  showExportForm() {
    this.prepareExport = true;
  }

  exportBulkItem(){
    if (this.bulkExport === false) this.finalArray[0].Options.Recursive = this.recursiveObject;

    let finalValues = {"Info": this.exportForm.value, "Objects" : this.finalArray}
    this.exportServiceCfg.bulkExport(finalValues)
    .subscribe(
      data => {
        this.exportedItem = data[1];
        saveAs(data[0],data[1].Info.FileName);
        this.exportResult = true;
      },
      err => console.error(err),
      () => console.log("DONE"),
    );
  }

  //SINGLE EXPORT

/*  exportItem(){
    this.exportServiceCfg.bulkExport(this.finalArray[0])
    .subscribe(
      data => {
        this.exportedItem = data[1];
        saveAs(data[0],data[1].Info.FileName);
        this.exportResult = true;
      },
      err => console.error(err),
      () => console.log("DONE"),
    );
  }
  */

//Load items functions from services depending on items selected Type
  loadItems(type, filter?) {
    this.resultArray = [];
    this.selectedFilterProp = ["ID"];
    this.listFilterProp = [];

    switch (type) {
      case 'snmpdevicecfg':
       this.snmpDeviceService.getDevices(filter)
       .subscribe(
       data => {
         //Load items on selection
         this.dataArray = data;
         this.resultArray = this.dataArray;
         for (let i in this.dataArray[0]) {
           this.listFilterProp.push({ 'id': i, 'name': i });
         }
       },
       err => {console.log(err)},
       () => {console.log("DONE")}
       );

      break;
      case 'influxcfg':
       this.influxServerService.getInfluxServer(filter)
       .subscribe(
       data => {
         this.dataArray=data;
         this.resultArray = this.dataArray;
         for (let i in this.dataArray[0]) {
           this.listFilterProp.push({ 'id': i, 'name': i });
         }
       },
       err => {console.log(err)},
       () => {console.log("DONE")}
       );
      break;
      case 'oidconditioncfg':
       this.oidConditionService.getConditions(filter)
       .subscribe(
       data => {
         this.dataArray=data;
         this.resultArray = this.dataArray;
         for (let i in this.dataArray[0]) {
           this.listFilterProp.push({ 'id': i, 'name': i });
         }
       },
       err => {console.log(err)},
       () => {console.log("DONE")}
       );
      break;
      case 'measfiltercfg':
       this.measFilterService.getMeasFilter(filter)
       .subscribe(
       data => {
         this.dataArray=data;
         this.resultArray = this.dataArray;
         for (let i in this.dataArray[0]) {
           this.listFilterProp.push({ 'id': i, 'name': i });
         }
       },
       err => {console.log(err)},
       () => {console.log("DONE")}
       );
      break;
      case 'customfiltercfg':
       this.customFilterService.getCustomFilter(filter)
       .subscribe(
       data => {
         this.dataArray=data;
         this.resultArray = this.dataArray;
         for (let i in this.dataArray[0]) {
           this.listFilterProp.push({ 'id': i, 'name': i });
         }
       },
       err => {console.log(err)},
       () => {console.log("DONE")}
       );
      break;
      case 'measurementcfg':
       this.influxMeasService.getMeas(filter)
       .subscribe(
       data => {
         this.dataArray=data;
         this.resultArray = this.dataArray;
         for (let i in this.dataArray[0]) {
           this.listFilterProp.push({ 'id': i, 'name': i });
         }
       },
       err => {console.log(err)},
       () => {console.log("DONE")}
       );
      break;
      case 'snmpmetriccfg':
       this.metricMeasService.getMetrics(filter)
       .subscribe(
       data => {
         this.dataArray=data;
         this.resultArray = this.dataArray;
         for (let i in this.dataArray[0]) {
           this.listFilterProp.push({ 'id': i, 'name': i });
         }
       },
       err => {console.log(err)},
       () => {console.log("DONE")}
       );
      break;
      case 'measgroupcfg':
       this.measGroupService.getMeasGroup(filter)
       .subscribe(
       data => {
         this.dataArray=data;
         this.resultArray = this.dataArray;
         for (let i in this.dataArray[0]) {
           this.listFilterProp.push({ 'id': i, 'name': i });
         }
       },
       err => {console.log(err)},
       () => {console.log("DONE")}
       );
      break;
      case 'varcatalogcfg':
      this.varCatalogService.getVarCatalog(filter)
      .subscribe(
      data => {
        this.dataArray=data;
        this.resultArray = this.dataArray;
        for (let i in this.dataArray[0]) {
          this.listFilterProp.push({ 'id': i, 'name': i });
        }
      },
      err => {console.log(err)},
      () => {console.log("DONE")}
      );
     break;
      default:
      break;
    }
  }
  //Common Functions
  isArray(myObject) {
    return myObject instanceof Array;
  }

  isObject(myObject) {
    return typeof myObject === 'object'
  }

  hide() {
    this.childModal.hide();
  }

}
