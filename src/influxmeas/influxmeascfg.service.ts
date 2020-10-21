import { HttpService } from '../core/http.service';;
import { Injectable } from '@angular/core';

import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class InfluxMeasService {

    constructor(public httpAPI: HttpService) {
        console.log('Task Service created.', httpAPI);
    }

    parseJSON(key,value) {
        if ( key == 'IndexAsValue' ) return ( value === "true" || value === true);
        return value;
    }

    addMeas(dev) {
        return this.httpAPI.post('/api/cfg/measurement',JSON.stringify(dev,this.parseJSON))
        .map( (responseData) => responseData.json());
    }

    editMeas(dev, id, hideAlert?) {
        console.log("DEV: ",dev);
        return this.httpAPI.put('/api/cfg/measurement/'+id,JSON.stringify(dev,this.parseJSON),null,hideAlert)
        .map( (responseData) => responseData.json());
    }

    getMeas(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/measurement')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((influxmeas) => {
            console.log("MAP SERVICE",influxmeas);
            let result = [];
            if (influxmeas) {
                _.forEach(influxmeas,function(value,key){
                    console.log("FOREACH LOOP",value,key);
                    if (value.GetMode == "indexed_mit") {
                        if (value.MultiTagOID.length > 0 ) {
                        value.TagOID = {multi: {}}
                         for (let t in value.MultiTagOID) {
                             value.TagOID.multi[t] = value.MultiTagOID[t].TagOID
                         }
                        }
                    }
                    if (value.GetMode == "indexed_multiple") {
                        value.IndexOID = {multi: {}}
                        value.TagOID = {multi: {}}
                        value.IndexTag = {multi: {}}
                        value.IndexTagFormat = {multi: {}}
                        value.IndexAsValue = { multi: {}}
                        for (let mm of value.MultiIndexCfg) {
                            value.IndexOID.multi[mm.Label] = mm.IndexOID
                            if (mm.GetMode == "indexed_mit") {
                                console.log(mm.GetMode, mm.Label)
                                for (let t in mm.MultiTagOID) {
                                    value.TagOID.multi[mm.Label+"["+t+"]"] = mm.MultiTagOID[t].TagOID    
                                } 
                            } else {
                                value.TagOID.multi[mm.Label] = mm.TagOID
                            }
                            value.IndexTag.multi[mm.Label] = mm.IndexTag
                            value.IndexTagFormat.multi[mm.Label] = mm.IndexTagFormat
                            value.IndexAsValue.multi[mm.Label] = mm.IndexAsValue ? mm.IndexAsValue : false
                        }
                    }
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
        return this.httpAPI.get('/api/cfg/measurement/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    getMeasByType(type : string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/measurement/type/'+type)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteInfluxMeas(id : string){
      return this.httpAPI.get('/api/cfg/measurement/checkondel/'+id)
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

    deleteMeas(id : string, hideAlert?) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/api/cfg/measurement/'+id,null,hideAlert)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
