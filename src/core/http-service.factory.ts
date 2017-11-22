import { XHRBackend } from '@angular/http';
import { DefaultRequestOptions } from '../core/default-request.options';
import { HttpService } from '../core/http.service';
import { LoaderService } from '../core/loader/loader.service';
import { Router } from '@angular/router';

function httpServiceFactory(backend: XHRBackend, options: DefaultRequestOptions, loaderService: LoaderService, router : Router ) {
    return new HttpService(backend, options, loaderService, router);
}

export { httpServiceFactory };
