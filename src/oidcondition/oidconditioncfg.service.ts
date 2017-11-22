import { HttpService } from '../core/http.service';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class OidConditionService {

    constructor(public httpAPI: HttpService) {
        console.log('Task Service created.', httpAPI);
    }

    addCondition(dev) {
        return this.httpAPI.post('/api/cfg/oidcondition',JSON.stringify(dev,function (key,value) {
            if ( key == 'IsMultiple' ) return ( value === "true" || value === true);
            return value;

        }))
        .map( (responseData) => responseData.json());
    }

    editCondition(dev, id, hideAlert?) {
        console.log("DEV: ",dev);
        return this.httpAPI.put('/api/cfg/oidcondition/'+id,JSON.stringify(dev,function (key,value) {
            if ( key == 'IsMultiple' ) return ( value === "true" || value === true);
            return value;

        }),null,hideAlert)
        .map( (responseData) => responseData.json());
    }

    getConditions(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/oidcondition')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((oidconditions) => {
            console.log("MAP SERVICE",oidconditions);
            let result = [];
            if (oidconditions) {
                _.forEach(oidconditions,function(value,key){
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
    getConditionsById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.httpAPI.get('/api/cfg/oidcondition/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteCondition(id : string){
      return this.httpAPI.get('/api/cfg/oidcondition/checkondel/'+id)
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

    deleteCondition(id : string, hideAlert?) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/api/cfg/oidcondition/'+id, null, hideAlert)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
