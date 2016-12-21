import { Injectable } from '@angular/core';
import * as _ from 'lodash';
import { HttpAPI } from '../common/httpAPI'

@Injectable()
export class InfluxServerService {

    constructor(public httpAPI: HttpAPI) {
    }

    addInfluxServer(dev) {
        return this.httpAPI.post('/influxservers',JSON.stringify(dev,function (key,value) {
                if ( key == 'Port' ) {
                  return parseInt(value);
                }
                return value;
        }))
        .map( (responseData) => responseData.json());

    }

    editInfluxServer(dev, id) {
        return this.httpAPI.put('/influxservers/'+id,JSON.stringify(dev,function (key,value) {
            if ( key == 'Port' ) {
              return parseInt(value);
            }
            return value;

        }))
        .map( (responseData) => responseData.json());
    }

    getInfluxServer(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/influxservers')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((influxservers) => {
            console.log("MAP SERVICE",influxservers);
            let result = [];
            if (influxservers) {
                _.forEach(influxservers,function(value,key){
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
    getInfluxServerById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.httpAPI.get('/influxservers/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteInfluxServer(id : string){
      return this.httpAPI.get('/influxservers/checkondel/'+id)
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

    deleteInfluxServer(id : string) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/influxservers/'+id)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
