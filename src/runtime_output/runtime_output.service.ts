import { HttpService } from '../core/http.service';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class RuntimeOutputService {

    constructor(public httpAPI: HttpService) {
    }


    getOutputRuntime(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/rt/output/info')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((runtime_outputs) => {
            let result = [];
            console.log(runtime_outputs)
            if (runtime_outputs) {
                _.forEach(runtime_outputs,function(value,key){
                    let tmp : any = {};
                  tmp.ID = key;
                  _.forEach(value, function(val,key) {
                    if (key == "Stats") {
                        _.forEach(val, function(val2, key) {
                            tmp[key] = val2;
                        })
                    } else {
                        tmp[key] = val;
                    }
                  });
                  result.push(tmp);
                  //result.push({'ID': key, 'value' :value});
                });
            }
            return result;
        });
    }

    getOutputRuntimeById(id : string) {
        // return an observable
        return this.httpAPI.get('/api/rt/output/info/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    /* Actions */
    // Active / Deactive

    changeOutputActive(id : string, event : boolean) {
        // return an observable
        console.log(event)
        if (event) {
            return this.httpAPI.get('/api/rt/output/buffer/'+id+'/active/')
            .map( (responseData) =>
                responseData.json()
            )
        } else {
            return this.httpAPI.get('/api/rt/output/buffer/'+id+'/deactive/')
            .map( (responseData) =>
                responseData.json()
            )
        }
    };

    changeEnqueuePolicy(id : string, event : boolean) {
        // return an observable
        if (event) {
            return this.httpAPI.get('/api/rt/output/buffer/'+id+'/enqueue/')
            .map( (responseData) =>
                responseData.json()
            )
        } else {
            return this.httpAPI.get('/api/rt/output/buffer/'+id+'/not_enqueue/')
            .map( (responseData) =>
                responseData.json()
            )
        }
    };

    flushbuffer(id : string) {
        // return an observable
        return this.httpAPI.get('/api/rt/output/buffer/'+id+'/flushbuffer/')
        .map( (responseData) =>
            responseData.json()
        )
    };

    resetbuffer(id : string) {
        // return an observable
        return this.httpAPI.get('/api/rt/output/buffer/'+id+'/resetbuffer/')
        .map( (responseData) =>
            responseData.json()
        )
    };
}
