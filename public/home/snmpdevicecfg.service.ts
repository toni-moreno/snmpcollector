import { Http } from 'angular2/http';
  // normally this would be imported from 'angular2/core'
  // but in our compiler we're pulling the dev version of angular2
import { Injectable } from 'angular2/core';
import { SnmpDeviceCfg } from './snmpdevicecfg';
import 'rxjs/Rx';
import _ from 'lodash';

@Injectable()
export class SnmpDeviceService {

 constructor(public http: Http) {
   console.log('Task Service created.', http);
 }

 getDevices(filter_s: string) {
    // return an observable
    return this.http.get('/snmpdevice')
    .map( (responseData) => {
      return responseData.json();
    })
    .map((snmpdevs) => {
	console.log("MAP SERVICE",snmpdevs);
      	let result = [];
      	if (snmpdevs) {
         _.forEach(snmpdevs,function(value,key){
		console.log("FOREACH LOOP",value,key);
		value.id = key;
		if(filter_s && filter_s.length > 0 ) {
			console.log("maching: "+key+ "filter: "+filter_s);
			var re = new RegExp(filter_s, 'gi');
			if (key.match(re)){
				result.push(value);
			}
			console.log(key.match(re));
		} else {
			result.push(value);
		}
	    });
        }
      return result;
    });
  }
}
