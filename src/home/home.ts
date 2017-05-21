import { Component, ViewChild,ViewContainerRef } from '@angular/core';
import { NgSwitch, NgSwitchCase, NgSwitchDefault } from '@angular/common';
import { Router } from '@angular/router';
import { HttpAPI} from '../common/httpAPI';
import { Observable } from 'rxjs/Observable';
import { BlockUIService } from '../common/blockui/blockui-service';
import { BlockUIComponent } from '../common/blockui/blockui-component'; // error
import { ImportFileModal } from '../common/dataservice/import-file-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { HomeService } from './home.service';
import { AboutModal } from './about-modal'
import { WindowRef } from '../common/windowref';
declare var _:any;

@Component({
  selector: 'home',
  templateUrl: './home.html',
  styleUrls: [ './home.css' ],
  providers: [BlockUIService, HomeService]
})

export class Home {

  @ViewChild('blocker', { read: ViewContainerRef }) container: ViewContainerRef;
  @ViewChild('importFileModal') public importFileModal : ImportFileModal;
  @ViewChild('exportBulkFileModal') public exportBulkFileModal : ExportFileModal;
  @ViewChild('aboutModal') public aboutModal : AboutModal;
  @ViewChild('RuntimeComponent') public rt : any;
  nativeWindow: any
  response: string;
  api: string;
  item_type: string;
  version: RInfo;
  configurationItems : Array<any> = [
  {'title': 'Influx Servers', 'selector' : 'influxserver'},
  {'title': 'OID Conditions', 'selector' : 'oidcondition'},
  {'title': 'SNMP Metrics', 'selector' : 'snmpmetric'},
  {'title': 'Influx Measurements', 'selector' : 'influxmeas'},
  {'title': 'Measurement Groups', 'selector' : 'measgroup'},
  {'title': 'Measurement Filters', 'selector' : 'measfilter'},
  {'title': 'Custom Filters', 'selector' : 'customfilter'},
  {'title': 'SNMP Devices', 'selector' : 'snmpdevice'},
  ];

  runtimeItems : Array<any> = [
  {'title': 'Device status', 'selector' : 'runtime'},
  ];

  mode : boolean = false;
  userIn : boolean = false;

  elapsedReload: string = '';
  lastReload: Date;

  constructor(private winRef: WindowRef,public router: Router, public httpAPI: HttpAPI, private _blocker: BlockUIService, public homeService: HomeService) {
    this.nativeWindow = winRef.nativeWindow;
    this.getFooterInfo();
    this.item_type= "runtime";
  }

  link(url: string) {
    this.nativeWindow.open(url);
  }


  logout() {
    this.homeService.userLogout()
    .subscribe(
    response => {
      this.router.navigate(['/login']);
    },
    error => {
      alert(error.text());
      console.log(error.text());
    }
    );
  }
  changeModeMenu() {
    this.mode = !this.mode
  }

  clickMenu(selected : string) : void {
    this.item_type = "";
    this.item_type = selected;
  }

  showImportModal() {
    this.importFileModal.initImport();
  }

  showExportBulkModal() {
    this.exportBulkFileModal.initExportModal(null, false);
  }

  showAboutModal() {
    this.aboutModal.showModal(this.version);
  }

  reloadConfig() {
    this._blocker.start(this.container, "Reloading Conf. Please wait...");
    if (this.rt) this.rt.updateRuntimeInfo(null,null,false);
    this.homeService.reloadConfig()
    .subscribe(
    response => {
      this.lastReload = new Date();
      this.elapsedReload = response;
      this._blocker.stop();
    },
    error => {
      this._blocker.stop();
      alert(error.text());
      console.log(error.text());
    }
    );
  }

  getFooterInfo() {
    this.homeService.getInfo()
    .subscribe(data => {
      this.version = data;
      this.userIn = true;
    },
    err => console.error(err),
    () =>  {}
    );
  }
}
