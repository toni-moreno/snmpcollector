import { Http,Headers } from 'angular2/http';
// normally this would be imported from 'angular2/core'
// but in our compiler we're pulling the dev version of angular2
import { Injectable } from 'angular2/core';
import 'rxjs/Rx';
import * as _ from 'lodash';

@Injectable()
export class SnmpMetricService {

    constructor(public http: Http) {
        console.log('Task Service created.', http);
    }

    addMetric(dev) {
        var headers = new Headers();
        headers.append("Content-Type", 'application/json');
        return this.http.post('/metric',JSON.stringify(dev,function (key,value) {
            if (key == 'Scale' || key == 'Shift') {
                return parseFloat(value);
            };
            if (key == 'GetRate'){
                if (value === 'true')   return true;
                else return false;
            }
            return value;

        }), { headers: headers })
        .map( (responseData) => responseData.json());
    }

    editMetric(dev, id) {
        var headers = new Headers();
        headers.append("Content-Type", 'application/json');
        console.log("DEV: ",dev);
        //TODO: Se tiene que coger el oldid para substituir en la configuración lo que toque!!!!
        return this.http.put('/metric/'+id,JSON.stringify(dev,function (key,value) {
            if (key == 'Scale' || key == 'Shift') {
                return parseFloat(value);
            }
            if (key == 'GetRate'){
                return Boolean(value);
            }
            return value;

        }), {  headers: headers   })
        .map( (responseData) => responseData.json());
    }

    getMetrics(filter_s: string) {
        // return an observable
        return this.http.get('/metric')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((snmpmetrics) => {
            console.log("MAP SERVICE",snmpmetrics);
            let result = [];
            if (snmpmetrics) {
                _.forEach(snmpmetrics,function(value,key){
                    console.log("FOREACH LOOP",value,key);
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
    getMetricsById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.http.get('/metric/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    deleteMetric(id : string) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.http.delete('/metric/'+id)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
