import { Injectable } from '@angular/core';
import { HttpService } from '../core/http.service';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class InfluxServerService {

    constructor(public httpAPI: HttpService) {
    }

    parseJSON(key,value) {
        if ( key == 'Port'  ||
        key == 'Timeout' ||
        key == 'BufferSize' ) {
          return parseInt(value);
        }
        if ( key == 'EnableSSL' ||
        key == 'InsecureSkipVerify') return ( value === "true" || value === true);
        return value;
    }

    addInfluxServer(dev) {
        return this.httpAPI.post('/api/cfg/influxservers',JSON.stringify(dev,this.parseJSON))
        .map( (responseData) => responseData.json());

    }

    editInfluxServer(dev, id, hideAlert?) {
        return this.httpAPI.put('/api/cfg/influxservers/'+id,JSON.stringify(dev,this.parseJSON),null,hideAlert)
        .map( (responseData) => responseData.json());
    }

    getInfluxServer(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/influxservers')
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
        return this.httpAPI.get('/api/cfg/influxservers/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteInfluxServer(id : string){
      return this.httpAPI.get('/api/cfg/influxservers/checkondel/'+id)
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

    testInfluxServer(influxserver,hideAlert?) {
      // return an observable
      return this.httpAPI.post('/api/cfg/influxservers/ping/',JSON.stringify(influxserver,this.parseJSON), null, hideAlert)
      .map((responseData) => responseData.json());
    };

    deleteInfluxServer(id : string, hideAlert?) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/api/cfg/influxservers/'+id, null, hideAlert)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
