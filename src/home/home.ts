import { Component, ViewChild,ViewContainerRef } from '@angular/core';
import { NgSwitch, NgSwitchCase, NgSwitchDefault } from '@angular/common';
import { Router } from '@angular/router';
import { HttpAPI} from '../common/httpAPI';
import { Observable } from 'rxjs/Observable';
import { BlockUIService } from '../common/blockui/blockui-service';
import { BlockUIComponent } from '../common/blockui/blockui-component'; // error

declare var _:any;

@Component({
  selector: 'home',
  templateUrl: './home.html',
  styleUrls: [ './home.css' ],
  providers: [BlockUIService]
})

export class Home {

  @ViewChild('blocker', { read: ViewContainerRef }) container: ViewContainerRef;

  response: string;
  api: string;
  item_type: string;
  version: any;
  menuItems : Array<any> = [
  {'title': 'Influx Servers', 'selector' : 'influxserver'},
  {'title': 'OID Conditions', 'selector' : 'oidcondition'},
  {'title': 'SNMP Metrics', 'selector' : 'snmpmetric'},
  {'title': 'Influx Measurements', 'selector' : 'influxmeas'},
  {'title': 'Measurement Groups', 'selector' : 'measgroup'},
  {'title': 'Measurement Filters', 'selector' : 'measfilter'},
  {'title': 'Custom Filters', 'selector' : 'customfilter'},
  {'title': 'SNMP Devices', 'selector' : 'snmpdevice'},
  {'title': 'Runtime', 'selector' : 'runtime'},
  ];
  elapsedReload: string = '';
  lastReload: Date;

  constructor(public router: Router, public httpAPI: HttpAPI, private _blocker: BlockUIService) {
    this.item_type= "runtime";
    this.getFooterInfo();
  }

  logout() {
    this.httpAPI.post('/logout','')
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

  clickMenu(selected : string) : void {
    this.item_type = "";
    this.item_type = selected;
  }

  reloadConfig() {
    this._blocker.start(this.container, "Reloading Conf. Please wait...");
    this.httpAPI.get('/api/rt/agent/reload/')
    .subscribe(
    response => {
      this.lastReload = new Date();
      this.elapsedReload = response.json();
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
    this.getInfo(null)
    .subscribe(data => {
      this.version = data;
    },
    err => console.error(err),
    () =>  {}
    );
  }

  getInfo(filter_s: string) {
    // return an observable
    return this.httpAPI.get('/api/rt/agent/info/version/')
    .map( (responseData) => {
      return responseData.json()});
    }
  }
