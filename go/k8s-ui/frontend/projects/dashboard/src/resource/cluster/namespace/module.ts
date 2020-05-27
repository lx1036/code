

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {ActionbarComponent} from './detail/actionbar/component';
import {NamespaceDetailComponent} from './detail/component';
import {NamespaceListComponent} from './list/component';
import {NamespaceRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, NamespaceRoutingModule],
  declarations: [NamespaceListComponent, NamespaceDetailComponent, ActionbarComponent],
})
export class NamespaceModule {}
