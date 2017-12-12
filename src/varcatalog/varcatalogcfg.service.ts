import { Injectable } from '@angular/core';
import { HttpService } from '../core/http.service';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class VarCatalogService {

    constructor(public httpAPI: HttpService) {
    }

    parseJSON(key,value) {
        if ( key == 'Port'  ||
        key == 'Timeout' ) {
          return parseInt(value);
        }
        return value; 
    }

    addVarCatalog(dev) {
        return this.httpAPI.post('/api/cfg/varcatalog',JSON.stringify(dev,this.parseJSON))
        .map( (responseData) => responseData.json());

    }

    editVarCatalog(dev, id, hideAlert?) {
        return this.httpAPI.put('/api/cfg/varcatalog/'+id,JSON.stringify(dev,this.parseJSON),null,hideAlert)
        .map( (responseData) => responseData.json());
    }

    getVarCatalog(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/varcatalog')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((varcatalog) => {
            console.log("MAP SERVICE",varcatalog);
            let result = [];
            if (varcatalog) {
                _.forEach(varcatalog,function(value,key){
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
    
    getVarCatalogById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.httpAPI.get('/api/cfg/varcatalog/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteVarCatalog(id : string){
      return this.httpAPI.get('/api/cfg/varcatalog/checkondel/'+id)
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

    deleteVarCatalog(id : string, hideAlert?) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/api/cfg/varcatalog/'+id, null, hideAlert)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
