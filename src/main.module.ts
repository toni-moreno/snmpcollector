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

import { AppRoutes } from './app/app.routes';
//common
import { ControlMessagesComponent } from './common/control-messages.component';
import { MultiselectDropdownModule } from './common/multiselect-dropdown'
import { PasswordToggleDirective } from './common/custom-directives'

//snmpcollector components

import { SnmpDeviceCfgComponent } from './snmpdevice/snmpdevicecfg.component';
import { OidConditionCfgComponent } from './oidcondition/oidconditioncfg.component';
import { SnmpMetricCfgComponent } from './snmpmetric/snmpmetriccfg.component';
import { InfluxMeasCfgComponent } from './influxmeas/influxmeascfg.component';
import { MeasGroupCfgComponent } from './measgroup/measgroupcfg.component';
import { MeasFilterCfgComponent } from './measfilter/measfiltercfg.component';
import { InfluxServerCfgComponent } from './influxserver/influxservercfg.component';
import { RuntimeComponent } from './runtime/runtime.component';
import { CustomFilterCfgComponent } from './customfilter/customfiltercfg.component';
import { BlockUIService } from './common/blockui/blockui-service';

import { AccordionModule , PaginationModule ,TabsModule } from 'ng2-bootstrap';
import { TooltipModule } from 'ng2-bootstrap';
import { ModalModule } from 'ng2-bootstrap';
import { ModalDirective } from 'ng2-bootstrap';

import { GenericModal } from './common/generic-modal';
import { ExportFileModal } from './common/dataservice/export-file-modal';
import { AboutModal } from './home/about-modal';
import { TreeView } from './common/dataservice/treeview';
import { TestConnectionModal } from './common/test-connection-modal';
import { TestFilterModal } from './customfilter/test-filter-modal';
import { ImportFileModal } from './common/dataservice/import-file-modal'

//others
import { WindowRef } from './common/windowref';
import { ValidationService } from './common/validation.service';
import { ExportServiceCfg } from './common/dataservice/export.service'
//pipes
import { ObjectParserPipe,SplitCommaPipe } from './common/custom_pipe';
import { ElapsedSecondsPipe } from './common/elapsedseconds.pipe';

import { BlockUIComponent } from './common/blockui/blockui-component';
import { SpinnerComponent } from './common/spinner';


@NgModule({
  bootstrap: [App],
  declarations: [
    PasswordToggleDirective,
    ObjectParserPipe,
    SplitCommaPipe,
    ElapsedSecondsPipe,
    ControlMessagesComponent,
    SnmpDeviceCfgComponent,
    OidConditionCfgComponent,
    SnmpMetricCfgComponent,
    InfluxMeasCfgComponent,
    MeasGroupCfgComponent,
    MeasFilterCfgComponent,
    InfluxServerCfgComponent,
    CustomFilterCfgComponent,
    RuntimeComponent,
    GenericModal,
    AboutModal,
    ExportFileModal,
    ImportFileModal,
    BlockUIComponent,
    TreeView,
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
    RouterModule.forRoot(AppRoutes)
  ],
  providers: [
    WindowRef,
    HttpAPI,
    ExportServiceCfg,
    ValidationService,
    BlockUIService
  ],
  entryComponents: [BlockUIComponent]
})
export class AppModule {}
