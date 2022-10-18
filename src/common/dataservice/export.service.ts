import { HttpService } from '../../core/http.service'
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class ExportServiceCfg {

    constructor(public httpAPI: HttpService) {
        console.log('Task Service created.', httpAPI);
    }

    exportFastRecursive(type : string, id : string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/export/'+type+'/'+id)
        .map((res) => {
        return [new Blob([res['_body']],{type: "application/json"}),res.json()];
        })
    }

    bulkExport(values) {
      return this.httpAPI.post('/api/cfg/bulkexport',values, null, true)
      .map((res) => {
          console.log(res);
          return [new Blob([res['_body']],{type: "application/json"}),res.json()];
      })
    }

    exportRecursive(type : string, id : string, values) {
        console.log(values);
        // return an observable
        return this.httpAPI.post('/api/cfg/export/'+type+'/'+id, values, null, true)
        .map((res) => {
            return [new Blob([res['_body']],{type: "application/json"}),res.json()];
        })
    }
}