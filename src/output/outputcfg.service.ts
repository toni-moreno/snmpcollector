import { Injectable } from '@angular/core';
import { HttpService } from '../core/http.service';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class OutputService {

    constructor(public httpAPI: HttpService) {
    }

    parseJSON(key,value) {
        if ( key == 'BufferSize'||
        key == 'FlushInterval' ||
        key == 'MetricBatchSize' ) {
          return parseInt(value);
        }
        if (key == 'Backend') {
            return value.split('..')[0]
        }
        if ( key == 'Active' ||
        key == 'EnqueueOnError') return ( value === "true" || value === true);
        return value;
    }

    addOutput(dev) {
        return this.httpAPI.post('/api/cfg/output',JSON.stringify(dev,this.parseJSON))
        .map( (responseData) => responseData.json());

    }

    editOutput(dev, id, hideAlert?) {
        return this.httpAPI.put('/api/cfg/output/'+id,JSON.stringify(dev,this.parseJSON),null,hideAlert)
        .map( (responseData) => responseData.json());
    }

    getOutput(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/output')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((output) => {
            console.log("MAP SERVICE",output);
            let result = [];
            if (output) {
                _.forEach(output,function(value,key){
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
    getOutputById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.httpAPI.get('/api/cfg/output/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteOutput(id : string){
      return this.httpAPI.get('/api/cfg/output/checkondel/'+id)
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

    testOutput(output,hideAlert?) {
      // return an observable
      return this.httpAPI.post('/api/cfg/output/ping/',JSON.stringify(output,this.parseJSON), null, hideAlert)
      .map((responseData) => responseData.json());
    };

    deleteOutput(id : string, hideAlert?) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/api/cfg/output/'+id, null, hideAlert)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
