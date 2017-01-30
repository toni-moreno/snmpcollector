import { Component } from '@angular/core';
import { NgSwitch, NgSwitchCase, NgSwitchDefault } from '@angular/common';
import { Router } from '@angular/router';
import { HttpAPI} from '../common/httpAPI';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Component({
  selector: 'home',
  templateUrl: './home.html',
  styleUrls: [ './home.css' ]
})

export class Home {

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

  constructor(public router: Router, public httpAPI: HttpAPI) {
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
    this.httpAPI.get('/api/rt/agent/reload/')
    .subscribe(
    response => {
      alert(response.json())
    },
    error => {
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
