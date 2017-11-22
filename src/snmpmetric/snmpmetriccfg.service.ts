import { HttpService } from '../core/http.service';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class SnmpMetricService {

    constructor(public httpAPI: HttpService) {
        console.log('Task Service created.', httpAPI);
    }

    addMetric(dev) {
        return this.httpAPI.post('/api/cfg/metric',JSON.stringify(dev,function (key,value) {
            if (key == 'Scale' ||
            key == 'Shift') {
                return parseFloat(value);
            };
            if (key == 'GetRate' ||
            key == 'IsTag' ) return ( value === "true" || value === true);
            return value;

        }))
        .map( (responseData) => responseData.json());
    }

    editMetric(dev, id, hideAlert?) {
        console.log("DEV: ",dev);
        return this.httpAPI.put('/api/cfg/metric/'+id,JSON.stringify(dev,function (key,value) {
            if (key == 'Scale' ||
            key == 'Shift') {
                return parseFloat(value);
            };
            if (key == 'GetRate' ||
            key == 'IsTag' ) return ( value === "true" || value === true);
            return value;

        }),null, hideAlert)
        .map( (responseData) => responseData.json());
    }

    getMetrics(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/metric')
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
        return this.httpAPI.get('/api/cfg/metric/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteMetric(id : string){
      return this.httpAPI.get('/api/cfg/metric/checkondel/'+id)
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

    deleteMetric(id : string, hideAlert?) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/api/cfg/metric/'+id, null, hideAlert)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
