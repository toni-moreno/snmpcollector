import { BaseRequestOptions } from '@angular/http';

export class DefaultRequestOptions extends BaseRequestOptions {

    constructor () {
        super();
        this.headers.append('Content-Type', 'application/json');
    }
}
