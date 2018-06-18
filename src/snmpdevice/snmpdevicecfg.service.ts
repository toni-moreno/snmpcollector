import { HttpService } from '../core/http.service';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class SnmpDeviceService {

    constructor(public httpAPI: HttpService) {
        console.log('Task Service created.', httpAPI);
    }

    parseJSON(key,value) {
        console.log("KEY: "+key+" Value: "+value);
        if ( key == 'Port' ||
        key == 'Retries' ||
        key == 'Timeout' ||
        key == 'Repeat' ||
        key == 'Freq' ||
        key == 'MaxRepetitions'  ||
        key == 'UpdateFltFreq') {
            return parseInt(value);
        }
        if ( key == 'Active' ||
        key == 'SnmpDebug' ||
        key == 'DisableBulk' ||
        key == 'ConcurrentGather') return ( value === "true" || value === true);
        if ( key == 'ExtraTags' ||
             key == 'SystemOIDs')
             return  String(value).split(',');
        if ( key == 'MeasFilters' ||
        key == 'MeasurementGroups' ||
        key == 'DeviceVars') {
            if (value == "") return null;
            else return value;
        }
        return value;
    }

    addDevice(dev,mode?) {
        var url = '/api/cfg/snmpdevice/';
        if (mode) {
            url += mode
        }
        return this.httpAPI.post(url,JSON.stringify(dev,this.parseJSON))
        .map( (responseData) => responseData.json());
    }

    editDevice(dev, id, mode?,hideAlert?) {
        console.log("DEV: ",dev,mode);
        var url = '/api/cfg/snmpdevice/'+id;
        if (mode) {
            url += '/'+mode
        }
        console.log("URL",url);
        //TODO: Se tiene que coger el oldid para substituir en la configuración lo que toque!!!!
        return this.httpAPI.put(url,JSON.stringify(dev,this.parseJSON),null,hideAlert)
        .map( (responseData) => responseData.json());
    }

    deleteDevice(id : string, mode?,hideAlert?) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        var url = '/api/cfg/snmpdevice/'+id;
        if (mode) {
            url += '/'+mode
        }
        return this.httpAPI.delete(url, null, hideAlert)
        .map( (responseData) => {
            responseData.json();
        })
    };


    getDevices(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/snmpdevice')
        .map( (responseData) => {
            return responseData.json();
        })
 
    }
    getDevicesById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.httpAPI.get('/api/cfg/snmpdevice/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteSNMPDevice(id : string){
    return this.httpAPI.get('/api/cfg/snmpdevice/checkondel/'+id)
    .map( (responseData) =>
     responseData.json()
    ).map((deleteobject) => {
        console.log("MAP SERVICE",deleteobject);
        let result : any = {'ID' : id};
        _.forEach(deleteobject,function(value,key){
            result[value.TypeDesc] = [];
        });
        _.forEach(deleteobject,function(value,key){
            result[value.TypeDesc].Description=value.Action;
            result[value.TypeDesc].push(value.ObID);
        });
        return result;
    });
  };


    pingDevice(dev, hideAlert?) {
        console.log(dev);
        return this.httpAPI.post('/api/rt/agent/snmpconsole/ping/',JSON.stringify(dev,this.parseJSON),null,hideAlert)
        .map( (responseData) => responseData.json());
    }

    sendQuery(dev,getMode,oid, hideAlert?) {
        return this.httpAPI.post('/api/rt/agent/snmpconsole/query/'+getMode+'/oid/'+oid,JSON.stringify(dev,this.parseJSON),null,hideAlert)
        .map( (responseData) => responseData.json());
    }
}
