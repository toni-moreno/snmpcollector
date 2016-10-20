import { Http,Headers } from '@angular/http';
import { Injectable } from '@angular/core';
import * as _ from 'lodash';

@Injectable()
export class MeasGroupService {

    constructor(public http: Http) {
        console.log('Task Service created.', http);
    }

    addMeasGroup(dev) {
        var headers = new Headers();
        headers.append("Content-Type", 'application/json');
        return this.http.post('/measgroups',JSON.stringify(dev,function (key,value) {
                if ( key == 'Measurements' ) {
                    if (!value) return null
                  else return String(value).split(',');
                }
                return value;
        }), { headers: headers })
        .map( (responseData) => responseData.json());
    }

    editMeasGroup(dev, id) {
        var headers = new Headers();
        headers.append("Content-Type", 'application/json');
        console.log("DEV: ",dev);
        //TODO: Se tiene que coger el oldid para substituir en la configuración lo que toque!!!!
        return this.http.put('/measgroups/'+id,JSON.stringify(dev,function (key,value) {
            if ( key == 'Measurements' ) {
              if (!value) return null
              else return String(value).split(',');
            }
            return value;
        }), {  headers: headers   })
        .map( (responseData) => responseData.json());
    }

    getMeasGroup(filter_s: string) {
        // return an observable
        return this.http.get('/measgroups')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((measgroups) => {
            console.log("MAP SERVICE",measgroups);
            let result = [];
            if (measgroups) {
                _.forEach(measgroups,function(value,key){
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
    getMeasGroupById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.http.get('/measgroups/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    deleteMeasGroup(id : string) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.http.delete('/measgroups/'+id)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
