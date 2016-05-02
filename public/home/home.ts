import { Component } from 'angular2/core';
import { CORE_DIRECTIVES } from 'angular2/common';
import { NgSwitch, NgSwitchWhen, NgSwitchDefault } from 'angular2/common';
import { Http, Headers } from 'angular2/http';
import { AuthHttp,JwtHelper } from 'angular2-jwt';
import { Router } from 'angular2/router';
import { SnmpDeviceCfgComponent } from './snmpdevicecfg.component';


@Component({
  selector: 'home',
  directives: [CORE_DIRECTIVES,NgSwitch,NgSwitchWhen,NgSwitchDefault,SnmpDeviceCfgComponent],
  templateUrl: '/public/home/home.html',
  styleUrls: [ '/public/home/home.css' ]
})

export class Home {
  jwt: string;
  decodedJwt: string;
  response: string;
  api: string;
  item_type: string;


  constructor(public router: Router, public http: Http, public authHttp: AuthHttp) {
    this.jwt = localStorage.getItem('jwt');
    this.decodedJwt = this.jwt;
    this.item_type= "snmpdevice";

  }

  logout() {
    localStorage.removeItem('jwt');
    this.router.parent.navigateByUrl('/login');
  }

  SNMPDevices() {
	  this.item_type = "snmpdevice";
	  alert('SnmpDevices!!!');
  }

  SNMPMetrics () {
	  this.item_type = "snmpmetrics";
	  alert('SnmpMetrics!!!');
  }
  IflxMeasurements() {
	  this.item_type = "ifluxmeas";
  }
  MetricGroups() {
	  this.item_type = "metricgroup";
  }
  InfluxServers() {
	  this.item_type = "influxserver";
  }
	
  /*
  callAnonymousApi() {
    this._callApi('Anonymous', 'http://localhost:3001/api/random-quote');
  }

  callSecuredApi() {
    this._callApi('Secured', 'http://localhost:3001/api/protected/random-quote');
  }

  _callApi(type, url) {
    this.response = null;
    if (type === 'Anonymous') {
      // For non-protected routes, just use Http
      this.http.get(url)
        .subscribe(
          response => this.response = response.text(),
          error => this.response = error.text()
        );
    }
    if (type === 'Secured') {
      // For protected routes, use AuthHttp
      this.authHttp.get(url)
        .subscribe(
          response => this.response = response.text(),
          error => this.response = error.text()
        );
    }
  }*/
}
