

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {StorageClassDetailComponent} from './detail/component';
import {StorageClassListComponent} from './list/component';
import {StorageClassRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, StorageClassRoutingModule],
  declarations: [StorageClassListComponent, StorageClassDetailComponent],
})
export class StorageClassModule {}
