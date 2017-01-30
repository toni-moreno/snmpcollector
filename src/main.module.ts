//Auth examples from: https://github.com/auth0-blog/angular2-authentication-sample
import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { RouterModule } from '@angular/router';
import { HttpModule } from '@angular/http';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { DropdownModule } from 'ng2-bootstrap';

// external libs

import { Ng2TableModule } from './common/ng-table/ng2-table';

import { Home } from './home/home';
import { Login } from './login/login';
import { App } from './app/app';
import { HttpAPI } from './common/httpAPI';

import { routes } from './app/app.routes';
//common
import { ControlMessagesComponent } from './common/control-messages.component';
import { MultiselectDropdownModule } from './common/multiselect-dropdown'
//snmpcollector components

import { SnmpDeviceCfgComponent } from './snmpdevice/snmpdevicecfg.component';
import { SnmpMetricCfgComponent } from './snmpmetric/snmpmetriccfg.component';
import { InfluxMeasCfgComponent } from './influxmeas/influxmeascfg.component';
import { MeasGroupCfgComponent } from './measgroup/measgroupcfg.component';
import { MeasFilterCfgComponent } from './measfilter/measfiltercfg.component';
import { InfluxServerCfgComponent } from './influxserver/influxservercfg.component';
import { RuntimeComponent } from './runtime/runtime.component';
import { CustomFilterCfgComponent } from './customfilter/customfiltercfg.component';

import { AccordionModule , PaginationModule ,TabsModule } from 'ng2-bootstrap';
import { TooltipModule } from 'ng2-bootstrap';
import { ModalModule } from 'ng2-bootstrap';
import { ModalDirective } from 'ng2-bootstrap';

import { GenericModal } from './common/generic-modal';
import { TestConnectionModal } from './common/test-connection-modal';
import { TestFilterModal } from './customfilter/test-filter-modal';

//others
import { ValidationService } from './common/validation.service';
//pipes
import { ObjectParserPipe } from './common/custom_pipe';

import { SpinnerComponent } from './common/spinner';

@NgModule({
  bootstrap: [App],
  declarations: [
    ObjectParserPipe,
    ControlMessagesComponent,
    SnmpDeviceCfgComponent,
    SnmpMetricCfgComponent,
    InfluxMeasCfgComponent,
    MeasGroupCfgComponent,
    MeasFilterCfgComponent,
    InfluxServerCfgComponent,
    CustomFilterCfgComponent,
    RuntimeComponent,
    GenericModal,
    SpinnerComponent,
    TestConnectionModal,
    TestFilterModal,
    Home,
    Login,
    App,
  ],
  imports: [
    HttpModule,
    BrowserModule,
    FormsModule,
    ReactiveFormsModule,
    MultiselectDropdownModule,
    AccordionModule.forRoot(),
    TooltipModule.forRoot(),
    ModalModule.forRoot(),
    PaginationModule.forRoot(),
    TabsModule.forRoot(),
    DropdownModule.forRoot(),
    Ng2TableModule,
    RouterModule.forRoot(routes, {
    //  useHash: true
    })
  ],
  providers: [
    HttpAPI,
    ValidationService,
  ]
})
export class AppModule {}
