

import {NgModule} from '@angular/core';

import {SharedModule} from '../../shared.module';
import {AutofocusDirective} from './autofocus/directive';

const directives = [AutofocusDirective];

@NgModule({
  imports: [SharedModule],
  declarations: [...directives],
  exports: [...directives],
})
export class DirectivesModule {}
