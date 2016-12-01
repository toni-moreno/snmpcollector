import { Component } from '@angular/core';
import { NgSwitch, NgSwitchCase, NgSwitchDefault } from '@angular/common';
import { Http, Headers } from '@angular/http';
import { Router } from '@angular/router';
import * as _ from 'lodash';
import { contentHeaders } from '../common/headers';



@Component({
  selector: 'home',
  templateUrl: 'public/home/home.html',
  styleUrls: [ 'public/home/home.css' ]
})

export class Home {

  response: string;
  api: string;
  item_type: string;
  version: any;


  constructor(public router: Router, public http: Http) {

    this.item_type= "runtime";
    this.getFooterInfo();
  }

  logout() {
    this.http.post('/logout', { headers: contentHeaders })
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
    this.http.get('/runtime/agent/reloadconf', { headers: contentHeaders })
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
    .subscribe(data => { this.version = data;},
     err => console.error(err),
     () =>  {}
     );
  }

  getInfo(filter_s: string) {
      // return an observable
      return this.http.get('/runtime/version')
      .map( (responseData) => {
          return responseData.json();
      });
  }
}
