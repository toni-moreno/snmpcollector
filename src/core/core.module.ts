import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { XHRBackend, RequestOptions } from '@angular/http';

import { HttpService } from './http.service';
import { httpServiceFactory } from './http-service.factory';
import { DefaultRequestOptions } from './default-request.options';
import { LoaderService } from './loader/loader.service';
import { LoaderComponent } from './loader/loader.component';
import { AlertModule } from 'ngx-bootstrap';
import { Router } from '@angular/router';

@NgModule({
    imports: [
        CommonModule,
        AlertModule.forRoot()
    ],
    exports: [
        LoaderComponent
    ],
    declarations: [
        LoaderComponent
    ],
    providers: [
        LoaderService,
        {
            provide: HttpService,
            useFactory: httpServiceFactory,
            deps: [XHRBackend, RequestOptions, LoaderService, Router ]
        }
    ]
})

export class CoreModule { }
