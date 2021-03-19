//Auth examples from: https://github.com/auth0-blog/angular2-authentication-sample
import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { RouterModule } from '@angular/router';
import { HttpModule } from '@angular/http';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { CoreModule } from './core/core.module';

// external libs

import { Ng2TableModule } from './common/ng-table/ng2-table';

import { Home } from './home/home';
import { Login } from './login/login';
import { App } from './app/app';

import { AppRoutes } from './app/app.routes';
//common
import { ControlMessagesComponent } from './common/control-messages.component';
import { MultiselectDropdownModule } from './common/multiselect-dropdown'
import { PasswordToggleDirective } from './common/custom-directives'
import { TableActions } from './common/table-actions';

//snmpcollector components
import { PollerLocationCfgComponent } from './pollerlocations/pollerlocations.component';
import { VarCatalogCfgComponent } from './varcatalog/varcatalogcfg.component';
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

import { AccordionModule , PaginationModule ,TabsModule } from 'ngx-bootstrap';
import { TooltipModule } from 'ngx-bootstrap';
import { ModalModule } from 'ngx-bootstrap';
import { ModalDirective } from 'ngx-bootstrap';
import { ProgressbarModule } from 'ngx-bootstrap';
import { BsDropdownModule } from 'ngx-bootstrap';

import { GenericModal } from './common/generic-modal';
import { ExportFileModal } from './common/dataservice/export-file-modal';
import { AboutModal } from './home/about-modal';
import { TreeView } from './common/dataservice/treeview';
import { TestConnectionModal } from './common/test-connection-modal';
import { TestFilterModal } from './customfilter/test-filter-modal';
import { ImportFileModal } from './common/dataservice/import-file-modal';
import { TableListComponent } from './common/table-list.component';

//others
import { WindowRef } from './common/windowref';
import { ValidationService } from './common/validation.service';
import { ExportServiceCfg } from './common/dataservice/export.service'
//pipes
import { ObjectParserPipe,SplitCommaPipe } from './common/custom_pipe';
import { ElapsedSecondsPipe } from './common/elapsedseconds.pipe';

import { CustomPipesModule } from './common/custom-pipe-module';

import { BlockUIComponent } from './common/blockui/blockui-component';
import { SpinnerComponent } from './common/spinner';


@NgModule({
  bootstrap: [App],
  declarations: [
    PasswordToggleDirective,
    ObjectParserPipe,
    //ElapsedSecondsPipe,
    SplitCommaPipe,
    TableActions,
    ControlMessagesComponent,
    SnmpDeviceCfgComponent,
    OidConditionCfgComponent,
    SnmpMetricCfgComponent,
    InfluxMeasCfgComponent,
    MeasGroupCfgComponent,
    MeasFilterCfgComponent,
    InfluxServerCfgComponent,
    CustomFilterCfgComponent,
    VarCatalogCfgComponent,
    PollerLocationCfgComponent,
    TableListComponent,
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
    CoreModule,
    CustomPipesModule,
    HttpModule,
    BrowserModule,
    FormsModule,
    ReactiveFormsModule,
    MultiselectDropdownModule,
    ProgressbarModule.forRoot(),
    AccordionModule.forRoot(),
    TooltipModule.forRoot(),
    ModalModule.forRoot(),
    PaginationModule.forRoot(),
    TabsModule.forRoot(),
    BsDropdownModule.forRoot(),
    Ng2TableModule,
    RouterModule.forRoot(AppRoutes)
  ],
  providers: [
    WindowRef,
    ExportServiceCfg,
    ValidationService,
    BlockUIService
  ],
  entryComponents: [BlockUIComponent]
})
export class AppModule {}
