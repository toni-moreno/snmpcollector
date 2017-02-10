import { HttpAPI } from '../common/httpAPI'
import { Injectable } from '@angular/core';

import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class InfluxMeasService {

    constructor(public httpAPI: HttpAPI) {
        console.log('Task Service created.', httpAPI);
    }

    addMeas(dev) {
        return this.httpAPI.post('/api/cfg/measurement',JSON.stringify(dev,function (key,value) {

            if ( key == 'IndexAsValue' ) return ( value === "true" || value === true);
            return value;
        }))
        .map( (responseData) => responseData.json());
    }

    editMeas(dev, id) {
        console.log("DEV: ",dev);
        return this.httpAPI.put('/api/cfg/measurement/'+id,JSON.stringify(dev,function (key,value) {

          if ( key == 'IndexAsValue' ) return ( value === "true" || value === true);
          return value;

      }))
        .map( (responseData) => responseData.json());
    }

    getMeas(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/measurement')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((influxmeas) => {
            console.log("MAP SERVICE",influxmeas);
            let result = [];
            if (influxmeas) {
                _.forEach(influxmeas,function(value,key){
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

    getMeasById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.httpAPI.get('/api/cfg/measurement/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    getMeasByType(type : string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/measurement/type/'+type)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteInfluxMeas(id : string){
      return this.httpAPI.get('/api/cfg/measurement/checkondel/'+id)
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

    deleteMeas(id : string) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/api/cfg/measurement/'+id)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
