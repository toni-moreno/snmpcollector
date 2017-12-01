import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TooltipModule } from 'ngx-bootstrap';
import { NgTableComponent } from './table/ng-table.component';
import { NgTableFilteringDirective } from './table/ng-table-filtering.directive';
import { NgTablePagingDirective } from './table/ng-table-paging.directive';
import { NgTableSortingDirective } from './table/ng-table-sorting.directive';
import { keyParserPipe } from './table/ng-table-keypipe';
import { CustomPipesModule }  from '../../custom-pipe-module';

//import { ObjectParserPipe as keyParserPipe } from '../../../common/custom_pipe';

@NgModule({
  imports: [CommonModule, TooltipModule, CustomPipesModule],
  declarations: [NgTableComponent, NgTableFilteringDirective, NgTablePagingDirective, NgTableSortingDirective, keyParserPipe],
  exports: [NgTableComponent, NgTableFilteringDirective, NgTablePagingDirective, NgTableSortingDirective]
})
export class Ng2TableModule {
}
