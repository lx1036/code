

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {PersistentVolumeDetailComponent} from './detail/component';
import {PersistentVolumeSourceComponent} from './detail/source/component';
import {PersistentVolumeListComponent} from './list/component';
import {PersistentVolumeRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, PersistentVolumeRoutingModule],
  declarations: [
    PersistentVolumeListComponent,
    PersistentVolumeDetailComponent,
    PersistentVolumeSourceComponent,
  ],
})
export class PersistentVolumeModule {}
