import { Component } from '@angular/core';
import { NgSwitch, NgSwitchCase, NgSwitchDefault } from '@angular/common';
import { Http, Headers } from '@angular/http';
import { AuthHttp,JwtHelper } from 'angular2-jwt';
import { Router } from '@angular/router';
import * as _ from 'lodash';



@Component({
  selector: 'home',
  templateUrl: 'public/home/home.html',
  styleUrls: [ 'public/home/home.css' ]
})

export class Home {
  jwt: string;
  decodedJwt: string;
  response: string;
  api: string;
  item_type: string;
  username: string;
  version: any;


  constructor(public router: Router, public http: Http, public authHttp: AuthHttp) {
    this.jwt = localStorage.getItem('id_token');
    this.username = localStorage.getItem('username');
    console.log('creating home!! id_token:'+this.jwt);

    this.decodedJwt = this.jwt ;
    this.item_type= "snmpdevice";
    this.getFooterInfo();
  }

  logout() {
    localStorage.removeItem('username');
    localStorage.removeItem('id_token');
    this.router.navigate(['']);
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
  SNMPDevicesTable() {
    this.item_type = "snmpdevice_table";
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
