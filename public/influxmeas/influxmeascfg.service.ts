import { Http,Headers } from '@angular/http';
import { Injectable } from '@angular/core';
import * as _ from 'lodash';

@Injectable()
export class InfluxMeasService {

    constructor(public http: Http) {
        console.log('Task Service created.', http);
    }

    addMeas(dev) {
        var headers = new Headers();
        headers.append("Content-Type", 'application/json');
        return this.http.post('/measurement',JSON.stringify(dev,function (key,value) {
            if ( key == 'Fields' ) {
              if (value == null || value == "")  return null;
              else return String(value).split(',');
            }
            if ( key == 'IndexAsValue' ) return ( value === "true" || value === true);
            return value;
        }), { headers: headers })
        .map( (responseData) => responseData.json());
    }

    editMeas(dev, id) {
        var headers = new Headers();
        headers.append("Content-Type", 'application/json');
        console.log("DEV: ",dev);
        return this.http.put('/measurement/'+id,JSON.stringify(dev,function (key,value) {
          if ( key == 'Fields' ) {
            if (value == null || value == "")  return null;
            else return String(value).split(',');
          }
          if ( key == 'IndexAsValue' ) return ( value === "true" || value === true);
          return value;

        }), {  headers: headers   })
        .map( (responseData) => responseData.json());
    }

    getMeas(filter_s: string) {
        // return an observable
        return this.http.get('/measurement')
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
        return this.http.get('/measurement/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    deleteMeas(id : string) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.http.delete('/measurement/'+id)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
