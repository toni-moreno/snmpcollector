import { HttpService } from '../core/http.service';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class MeasFilterService {

    constructor(public httpAPI: HttpService) {
        console.log('Task Service created.', httpAPI);
    }

    parseJSON(key,value) {
        if ( key == 'EnableAlias' ) return ( value === "true" || value === true);
        if ( key == 'IDMeasurementCfg') {
            if ( value == "" ) return null
        }
        return value;    
    }

    addMeasFilter(dev) {
        return this.httpAPI.post('/api/cfg/measfilters',JSON.stringify(dev,this.parseJSON))
        .map( (responseData) => responseData.json());
    }

    editMeasFilter(dev, id, hideAlert?) {
        console.log("DEV: ",dev);
        return this.httpAPI.put('/api/cfg/measfilters/'+id,JSON.stringify(dev,this.parseJSON),null,hideAlert)
        .map( (responseData) => responseData.json());
    }

    getMeasFilter(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/measfilters')
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
        return this.httpAPI.get('/api/cfg/measfilters/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteMeasFilter(id : string){
      return this.httpAPI.get('/api/cfg/measfilters/checkondel/'+id)
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

    deleteMeasFilter(id : string, hideAlert?) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/api/cfg/measfilters/'+id, null, hideAlert)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
