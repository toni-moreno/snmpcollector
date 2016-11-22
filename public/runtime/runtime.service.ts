import { Http,Headers } from '@angular/http';
import { Injectable } from '@angular/core';
import * as _ from 'lodash';

@Injectable()
export class RuntimeService {

    constructor(public http: Http) {
        console.log('Task Service created.', http);
    }


    getRuntime(filter_s: string) {
        // return an observable
        return this.http.get('/runtime/info')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((runtime_devs) => {
            console.log("snmpdevs: ",runtime_devs);
            console.log("MAP SERVICE",runtime_devs);
            let result = [];
            if (runtime_devs) {
                _.forEach(runtime_devs,function(value,key){
                  console.log("KEY: ",key);
                  console.log("FOREACH LOOP",value,key);
                    if(filter_s && filter_s.length > 0 ) {
                        console.log("maching: "+key+ "| filter: "+filter_s);
                        var re = new RegExp(filter_s, 'gi');
                        if (key.match(filter_s)){
                            result.push({'ID': key, 'value' :value});
                        }
                        console.log(key.match(re));
                    } else {
                            result.push({'ID': key, 'value' :value});
                    }
                });
            }
            console.log("result:",result);
            return result;
        });
    }

    getRuntimeById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.http.get('/runtime/info/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    changeDeviceActive(id : string, event : boolean) {
        // return an observable
        console.log("ID: ",id);
        console.log("RECIEVED EVENT: ", event);

        if (event) {
            return this.http.put('/runtime/activatedev/'+id,id)
            .map( (responseData) =>
                responseData.json()
            )
        } else {
            return this.http.put('/runtime/deactivatedev/'+id,id)
            .map( (responseData) =>
                responseData.json()
            )
        }
    };

    changeStateDebug(id : string, event : boolean) {
        // return an observable
        console.log("ID: ",id);
        console.log("RECIEVED EVENT: ", event);

        if (event) {
            return this.http.put('/runtime/actsnmpdbg/'+id,id)
            .map( (responseData) =>
                responseData.json()
            )
        } else {
            return this.http.put('/runtime/deactsnmpdbg/'+id,id)
            .map( (responseData) =>
                responseData.json()
            )
        }
    };



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
