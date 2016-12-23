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

  constructor(public router: Router, public httpAPI: HttpAPI) {
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

  reloadConfig() {
    this.httpAPI.get('/runtime/agent/reloadconf')
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

  InfluxServers() {
	  this.item_type = "influxserver";
  }

  SNMPMetrics () {
	  this.item_type = "snmpmetric";
  }

  InfluxMeasurements() {
    this.item_type = "influxmeas";
  }

  MeasGroups() {
    this.item_type = "measgroup";
  }

  MeasFilters() {
    this.item_type = "measfilter";
  }

  SNMPDevices() {
	  this.item_type = "snmpdevice";
  }

  Runtime() {
    this.item_type = "runtime";
  }


  getFooterInfo() {
    this.getInfo(null)
    .subscribe(data => {
      this.version = data;
      this.item_type= "runtime";
    },
     err => console.error(err),
     () =>  {}
     );
  }

  getInfo(filter_s: string) {
      // return an observable
      return this.httpAPI.get('/runtime/version')
      .map( (responseData) => {
          return responseData.json()});
  }
}
