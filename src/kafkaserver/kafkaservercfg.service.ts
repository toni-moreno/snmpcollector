import { Injectable } from '@angular/core';
import { HttpService } from '../core/http.service';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class KafkaServerService {

    constructor(public httpAPI: HttpService) {
    }

    parseJSON(key,value) {
        if ( key == 'RequiredAcks'  ||
        key == 'MaxRetry'  ||
        key == 'CompressionCodec' ||
        key == 'MaxMessageBytes') {
          return parseInt(value);
        }
        if ( key == 'EnableTLS' ||
        key == 'Socks5ProxyEnabled' ||
        key == 'InsecureSkipVerify' ||
        key == 'ExcludeTopicTag') {
            return ( value === "true" || value === true);
        }
        if ( key == 'Brokers' || key == 'Keys')
        return  String(value).split(',');

        return value;
    }

    addKafkaServer(dev) {
        return this.httpAPI.post('/api/cfg/kafkaservers',JSON.stringify(dev,this.parseJSON))
        .map( (responseData) => responseData.json());

    }

    editKafkaServer(dev, id, hideAlert?) {
        return this.httpAPI.put('/api/cfg/kafkaservers/'+id,JSON.stringify(dev,this.parseJSON),null,hideAlert)
        .map( (responseData) => responseData.json());
    }

    getKafkaServer(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/kafkaservers')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((kafkaservers) => {
            console.log("MAP SERVICE",kafkaservers);
            let result = [];
            if (kafkaservers) {
                _.forEach(kafkaservers,function(value,key){
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
    getKafkaServerById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.httpAPI.get('/api/cfg/kafkaservers/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteKafkaServer(id : string){
      return this.httpAPI.get('/api/cfg/kafkaservers/checkondel/'+id)
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

    testKafkaServer(kafkaserver,hideAlert?) {
      // return an observable
      return this.httpAPI.post('/api/cfg/kafkaservers/ping/',JSON.stringify(kafkaserver,this.parseJSON), null, hideAlert)
      .map((responseData) => responseData.json());
    };

    deleteKafkaServer(id : string, hideAlert?) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/api/cfg/kafkaservers/'+id, null, hideAlert)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
