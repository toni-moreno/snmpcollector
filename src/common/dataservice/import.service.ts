import { HttpService } from '../../core/http.service'
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class ImportServiceCfg {

    constructor(public httpAPI: HttpService) {
        console.log('Task Service created.', httpAPI);
    }

    importItem(data) {
        console.log(data);
        let formData = new FormData();
        console.log(formData);
        formData.append('auto_rename',data.auto_rename );
        formData.append('over_write',data.over_write );
        formData.append('export_file', data.files[0], data.files[0].name);
        return this.httpAPI.postFile('/api/cfg/import/', formData)
        .map((res) => {
            return res.json();
        })
    }
}
