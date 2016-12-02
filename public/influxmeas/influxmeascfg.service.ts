import { HttpAPI } from '../common/httpAPI'
import { Injectable } from '@angular/core';
import * as _ from 'lodash';

@Injectable()
export class InfluxMeasService {

    constructor(public httpAPI: HttpAPI) {
        console.log('Task Service created.', httpAPI);
    }

    addMeas(dev) {
        return this.httpAPI.post('/measurement',JSON.stringify(dev,function (key,value) {
            if ( key == 'Fields' ) {
              if (value == null || value == "")  return null;
              else return String(value).split(',');
            }
            if ( key == 'IndexAsValue' ) return ( value === "true" || value === true);
            return value;
        }))
        .map( (responseData) => responseData.json());
    }

    editMeas(dev, id) {
        console.log("DEV: ",dev);
        return this.httpAPI.put('/measurement/'+id,JSON.stringify(dev,function (key,value) {
          if ( key == 'Fields' ) {
            if (value == null || value == "")  return null;
            else return String(value).split(',');
          }
          if ( key == 'IndexAsValue' ) return ( value === "true" || value === true);
          return value;

      }))
        .map( (responseData) => responseData.json());
    }

    getMeas(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/measurement')
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
        return this.httpAPI.get('/measurement/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteInfluxMeas(id : string){
      return this.httpAPI.get('/measurement/checkondel/'+id)
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
        return this.httpAPI.delete('/measurement/'+id)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
