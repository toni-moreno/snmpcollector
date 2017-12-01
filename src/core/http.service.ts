import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import 'rxjs/Rx';
import {
    Http,
    RequestOptions,
    RequestOptionsArgs,
    Response,
    Request,
    Headers,
    XHRBackend
} from '@angular/http';
import { Router } from '@angular/router';


import { DefaultRequestOptions } from './default-request.options';

import { LoaderService } from './loader/loader.service';

@Injectable()
export class HttpService extends Http {

    public router : Router;

    protected headersUpload: Headers;

      constructor(
        backend: XHRBackend,
        defaultOptions: DefaultRequestOptions,
        private loaderService: LoaderService,
        public _router : Router
    ) {
        super(backend, defaultOptions);
        this.router = _router;
    }

    get(url: string, options?: RequestOptionsArgs ): Observable<any> {
        return super.get(this.getFullUrl(url), this.requestOptions(options))
            .catch(this.onCatch.bind(this))
            .do((res: Response) => {
            }, (error: any) => {
                this.onError(error);
            })
            .finally(() => {
                this.onEnd();
            });
    }

    postFile(url:string, data:any, options?: RequestOptionsArgs) : Observable<any> {
        if (options == null) options = {};
        options.headers = this.headersUpload;
        return super.post(this.getFullUrl(url), data, options)
            .catch(this.onCatch.bind(this))
            .do((res: Response) => {
                this.onSuccess(res);
            }, (error: any) => {
                this.onError(error);
            })
            .finally(() => {
                this.onEnd();
            });
    }

    post(url: string, data:any, options?: RequestOptionsArgs, hideAlert? : boolean ): Observable<any> {
        return super.post(this.getFullUrl(url), data, this.requestOptions(options))
            .catch(this.onCatch.bind(this))
            .do((res: Response) => {
                if (!hideAlert) this.onSuccess(res);
            }, (error: any) => {
                this.onError(error);
            })
            .finally(() => {
                this.onEnd();
            });
    }

    put(url: string, data:any, options?: RequestOptionsArgs, hideAlert? : boolean): Observable<any> {
        return super.put(this.getFullUrl(url), data, this.requestOptions(options))
            .catch(this.onCatch.bind(this))
            .do((res: Response) => {
                if (!hideAlert) this.onSuccess(res);
            }, (error: any) => {
                this.onError(error);
            })
            .finally(() => {
                this.onEnd();
            });
    }

    delete(url: string, options?: RequestOptionsArgs, hideAlert? : boolean ): Observable<any> {
        return super.delete(this.getFullUrl(url), this.requestOptions(options))
            .catch(this.onCatch.bind(this))
            .do((res: Response) => {
                if (!hideAlert) this.onSuccess(res);
            }, (error: any) => {
                this.onError(error);
            })
            .finally(() => {
                this.onEnd();
            });
    }

    private requestOptions(options?: RequestOptionsArgs, file?): RequestOptionsArgs {

        if (options == null) {
            options = new DefaultRequestOptions();
        }
        if (options.headers == null) {
            options.headers = new Headers();
        }
        return options;
    }

    private getFullUrl(url: string): string {
        return encodeURI(url);
    }

    private onCatch(error: any, caught: Observable<any>): Observable<any> {
        if (error['status'] == 403) {
            this.router.navigate(['/login']);
            this.loaderService.show(error,'danger');
        }else if (error['status'] == 0) {
            //alert('Server seems not being running...');
            this.loaderService.show('Server seems not being running...','danger');
        } else if (error['status'] == 404) {
            this.loaderService.show(error,'danger');
            //alert('CODE :'+error.status +'\n'+"ERROR: \t"+error['_body']);
        }
        return Observable.throw(error);
    }

    private onSuccess(res: Response): void {
        this.loaderService.show(res,'success');

    }

    private onError(res: Response): void {
        console.log('Error, status code: ' + res.status);
    }

    onEnd() {
        console.log("finished");
    }

}
