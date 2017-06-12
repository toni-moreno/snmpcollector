import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ElapsedSecondsPipe }  from './elapsedseconds.pipe';

//import { ObjectParserPipe as keyParserPipe } from '../../../common/custom_pipe';

@NgModule({
  imports: [CommonModule],
  declarations: [ElapsedSecondsPipe],
  exports: [ElapsedSecondsPipe]
})
export class CustomPipesModule {
}
