//Auth examples from: https://github.com/auth0-blog/angular2-authentication-sample
import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { RouterModule } from '@angular/router';
import { HttpModule } from '@angular/http';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

// external libs

import { Ng2TableModule } from './common/ng-table/ng2-table';

import { Home } from './home/home';
import { Login } from './login/login';
import { App } from './app/app';

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

import { AccordionModule , PaginationModule ,TabsModule } from 'ng2-bootstrap/ng2-bootstrap';
import { TooltipModule } from 'ng2-bootstrap/ng2-bootstrap';
import { ModalModule } from 'ng2-bootstrap/ng2-bootstrap';
import { AlertModule } from 'ng2-bootstrap/ng2-bootstrap';
import { GenericModal } from './common/generic-modal'
//others
import { ValidationService } from './common/validation.service';
//pipes
import { ObjectParserPipe } from './common/custom_pipe'

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
    RuntimeComponent,
    GenericModal,
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
    AccordionModule,
    TooltipModule,
    ModalModule,
    AlertModule,
    PaginationModule,
    TabsModule,
    Ng2TableModule,
    RouterModule.forRoot(routes, {
    //  useHash: true
    })
  ],
  providers: [
    ValidationService,
  ]
})
export class AppModule {}
