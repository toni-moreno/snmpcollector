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

  response: string;
  api: string;
  item_type: string;
  version: any;
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
  {'title': 'Runtime', 'selector' : 'runtime'},
  ];

  mode : boolean = true;
  userIn : boolean = false;

  elapsedReload: string = '';
  lastReload: Date;

  constructor(public router: Router, public httpAPI: HttpAPI, private _blocker: BlockUIService, public homeService: HomeService) {
    this.getFooterInfo();
    this.item_type= "runtime";
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
    console.log(this.mode)
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

  reloadConfig() {
    this._blocker.start(this.container, "Reloading Conf. Please wait...");
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
