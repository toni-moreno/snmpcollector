import { Http,Headers } from '@angular/http';
import { Injectable } from '@angular/core';
import * as _ from 'lodash';

@Injectable()
export class MeasFilterService {

    constructor(public http: Http) {
        console.log('Task Service created.', http);
    }

    addMeasFilter(dev) {
        var headers = new Headers();
        headers.append("Content-Type", 'application/json');
        return this.http.post('/measfilters',JSON.stringify(dev,function (key,value) {
                if ( key == 'EnableAlias' ) {
                  if (value == "true") return true;
                  else return false;
                }
                return value;
        }), { headers: headers })
        .map( (responseData) => responseData.json());
    }

    editMeasFilter(dev, id) {
        var headers = new Headers();
        headers.append("Content-Type", 'application/json');
        console.log("DEV: ",dev);
        //TODO: Se tiene que coger el oldid para substituir en la configuración lo que toque!!!!
        return this.http.put('/measfilters/'+id,JSON.stringify(dev,function (key,value) {
            if ( key == 'EnableAlias' ) {
              if (value == "true") return true;
              else return false;
            }
            return value;

        }), {  headers: headers   })
        .map( (responseData) => responseData.json());
    }

    getMeasFilter(filter_s: string) {
        // return an observable
        return this.http.get('/measfilters')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((measfilter) => {
            console.log("MAP SERVICE",measfilter);
            let result = [];
            if (measfilter) {
                _.forEach(measfilter,function(value,key){
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
    getMeasFilterById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.http.get('/measfilters/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    deleteMeasFilter(id : string) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.http.delete('/measfilters/'+id)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
