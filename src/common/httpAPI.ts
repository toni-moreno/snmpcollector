import {Injectable} from '@angular/core';
import {Http, Request, Response, Headers, RequestOptionsArgs, RequestMethod} from "@angular/http";
import {RequestArgs} from "@angular/http/src/interfaces";
import {Observable} from 'rxjs/Observable';
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/catch';
import 'rxjs/add/observable/throw';

import { Router } from '@angular/router';

@Injectable()
export class HttpAPI {
    protected headers: Headers;
    protected router : Router;
    protected headersUpload: Headers;
    public testi;

    constructor(private _http: Http, public _router : Router) {
        this.router = _router;
        this.headers = new Headers();
        this.headers.append('Content-Type', 'application/json');
        this.headers.append('Accept', 'application/json');
        this.headersUpload = new Headers();
        this.headersUpload.append('Content-Type', 'multipart/form-data')
        this.headersUpload.append('Accept', 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8')
    }

    get(url:string) : Observable<any> {
        return this._http.get(encodeURI(url))
            .catch(this.handleError.bind(this));
    }

    post(url:string, data:any, args?: RequestOptionsArgs) : Observable<any> {
        if (args == null) args = {};
        if (args.headers === undefined) args.headers = this.headers;
        return this._http.post(encodeURI(url), data, args)
            .catch(this.handleError.bind(this));
    }

    postFile(url:string, data:any, args?: RequestOptionsArgs) : Observable<any> {
        if (args == null) args = {};
        args.headers = this.headersUpload;
        return this._http.post(encodeURI(url), data, this.headersUpload)
            .catch(this.handleError.bind(this));
    }

    put(url:string, data:any, args?: RequestOptionsArgs) : Observable<any> {
        if (args == null) args = {};
        if (args.headers === undefined) args.headers = this.headers;
        return this._http.put(encodeURI(url), data, args)
            .catch(this.handleError.bind(this));
    }

    delete(url: string, data?: any, args?: RequestOptionsArgs): Observable<any> {
        if (args == null) args = {};
        args.url = encodeURI(url);
        args.method = RequestMethod.Delete;
        if (!args.headers) args.headers = this.headers;
        args.body  = data ? (data) : null;

        return this._http.request(new Request(<RequestArgs>args))
            .catch(this.handleError.bind(this))
    }

    /*private static json(res: Response): any {
        return res.text() === "" ? res : res.json();
    }*/

    private handleError(error:any) {
        if (error['status'] == 403) {
            this.router.navigate(['/login']);
        }else if (error['status'] == 0) {
            alert('Server seems not being running...');
        } else if (error['status'] == 404) {
            console.log(error);
            alert('CODE :'+error.status +'\n'+"ERROR: \t"+error['_body']);
        }
        //return Observable.bindNodeCallback(this.test);
        return Observable.throw(error);

    }
}
