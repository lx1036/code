

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../common/components/module';
import {SharedModule} from '../shared.module';

import {SearchComponent} from './component';
import {SearchRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, SearchRoutingModule],
  declarations: [SearchComponent],
})
export class SearchModule {}
