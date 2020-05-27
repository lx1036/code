

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {PersistentVolumeClaimDetailComponent} from './detail/component';
import {PersistentVolumeClaimListComponent} from './list/component';
import {PersistentVolumeClaimRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, PersistentVolumeClaimRoutingModule],
  declarations: [PersistentVolumeClaimListComponent, PersistentVolumeClaimDetailComponent],
})
export class PersistentVolumeClaimModule {}
