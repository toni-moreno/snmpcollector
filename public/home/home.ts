import { Component } from '@angular/core';
import { NgSwitch, NgSwitchCase, NgSwitchDefault } from '@angular/common';
import { Http, Headers } from '@angular/http';
import { AuthHttp,JwtHelper } from 'angular2-jwt';
import { Router } from '@angular/router';



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


  constructor(public router: Router, public http: Http, public authHttp: AuthHttp) {
    this.jwt = localStorage.getItem('id_token');
    console.log('creating home!! id_token:'+this.jwt);
    this.decodedJwt = this.jwt ;
    this.item_type= "snmpdevice";

  }

  logout() {
    localStorage.removeItem('id_token');
    this.router.navigate(['']);
  }

  SNMPDevices() {
	  this.item_type = "snmpdevice";
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
  
  InfluxServers() {
	  this.item_type = "influxserver";
  }

}
