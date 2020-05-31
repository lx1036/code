

import {NgModule} from '@angular/core';
import {RouterModule} from '@angular/router';
import {NamespacedResourceService, ResourceService} from './resource';
import {UtilityService} from './utility';

@NgModule({
  imports: [RouterModule],
  providers: [ResourceService, NamespacedResourceService, UtilityService],
})
export class ResourceModule {}
