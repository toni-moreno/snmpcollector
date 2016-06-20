import { Http,Headers } from 'angular2/http';
// normally this would be imported from 'angular2/core'
// but in our compiler we're pulling the dev version of angular2
import { Injectable } from 'angular2/core';
import 'rxjs/Rx';
import * as _ from 'lodash';

@Injectable()
export class SnmpDeviceService {

    constructor(public http: Http) {
        console.log('Task Service created.', http);
    }

    addDevice(dev) {
        var headers = new Headers();
        headers.append("Content-Type", 'application/json');
        return this.http.post('/snmpdevice',JSON.stringify(dev,function (key,value) {
            if (	key == 'Port' ||
            key == 'Retries' ||
            key == 'Timeout' ||
            key == 'Repeat' ||
            key == 'Freq' ) {
                return parseInt(value);
            }
            if ( key == 'SnmpDebug' ) return ( value === "true");
            if ( key == 'Extratags' ) return  value.split(',');
            return value;

        }), {  headers: headers   })
        .map( (res) => res.json())
        .subscribe(
        data=> {console.log('data post',data);},
        err => console.log(err),
        () => console.log('complete')
        );
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
