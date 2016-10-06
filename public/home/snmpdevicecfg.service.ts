import { Http,Headers } from '@angular/http';
import { Injectable } from '@angular/core';
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
            console.log("KEY:"+key+" Value"+value);
            if (	key == 'Port' ||
            key == 'Retries' ||
            key == 'Timeout' ||
            key == 'Repeat' ||
            key == 'Freq' ) {
                return parseInt(value);
            }
            if ( key == 'SnmpDebug' ) return ( value === "true");
            if ( key == 'Extratags' ) return  String(value).split(',');
            //TODO Actualizar en Angular2 Final con Multiselect
            if ( key == 'MeasFilters') return value.split(',');
            if ( key == 'MetricGroups') return value.split(',');
            return value;

        }), { headers: headers })
        .map( (responseData) => responseData.json());
    }

    editDevice(dev, id) {
        var headers = new Headers();
        headers.append("Content-Type", 'application/json');
        console.log("DEV: ",dev);
        //TODO: Se tiene que coger el oldid para substituir en la configuración lo que toque!!!!
        return this.http.put('/snmpdevice/'+id,JSON.stringify(dev,function (key,value) {
            if (key == 'Port' ||
            key == 'Retries' ||
            key == 'Timeout' ||
            key == 'Repeat' ||
            key == 'Freq' ) {
                return parseInt(value);
            }
            if ( key == 'SnmpDebug' ) return ( value === "true");
            if ( key == 'Extratags' ) return  String(value).split(',');
            return value;

        }), {  headers: headers   })
        .map( (responseData) => responseData.json());
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
                    console.log("FOREACH LOOP",value,value.ID);
                    if(filter_s && filter_s.length > 0 ) {
                        console.log("maching: "+value.ID+ "filter: "+filter_s);
                        var re = new RegExp(filter_s, 'gi');
                        if (value.ID.match(re)){
                            result.push(value);
                        }
                        console.log(value.ID.match(re));
                    } else {
                        result.push(value);
                    }
                });
            }
            return result;
        });
    }
    getDevicesById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.http.get('/snmpdevice/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    deleteDevice(id : string) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.http.delete('/snmpdevice/'+id)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
