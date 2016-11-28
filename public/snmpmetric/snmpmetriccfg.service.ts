import { Http,Headers } from '@angular/http';
import { Injectable } from '@angular/core';
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
            if (key == 'Scale' ||
            key == 'Shift') {
                return parseFloat(value);
            };
            if (key == 'GetRate' ||
            key == 'IsTag' ) return ( value === "true" || value === true);
            return value;

        }), { headers: headers })
        .map( (responseData) => responseData.json());
    }

    editMetric(dev, id) {
        var headers = new Headers();
        headers.append("Content-Type", 'application/json');
        console.log("DEV: ",dev);
        return this.http.put('/metric/'+id,JSON.stringify(dev,function (key,value) {
            if (key == 'Scale' ||
            key == 'Shift') {
                return parseFloat(value);
            };
            if (key == 'GetRate' ||
            key == 'IsTag' ) return ( value === "true" || value === true);
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
    getMetricsById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.http.get('/metric/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteMetric(id : string){
      return this.http.get('/metric/checkondel/'+id)
      .map( (responseData) =>
       responseData.json()
      ).map((deleteobject) => {
          console.log("MAP SERVICE",deleteobject);
          let result : any = {'ID' : id};
          _.forEach(deleteobject,function(value,key){
              result[value.Type] = [];
          });
          _.forEach(deleteobject,function(value,key){
              result[value.Type].Description=value.Action;
              result[value.Type].push(value.ObID);
          });
          return result;
      });
    };

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
