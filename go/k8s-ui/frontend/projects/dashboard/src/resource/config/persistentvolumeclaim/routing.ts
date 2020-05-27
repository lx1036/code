

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {CONFIG_ROUTE} from '../routing';

import {PersistentVolumeClaimDetailComponent} from './detail/component';
import {PersistentVolumeClaimListComponent} from './list/component';

const PERSISTENTVOLUMECLAIM_LIST_ROUTE: Route = {
  path: '',
  component: PersistentVolumeClaimListComponent,
  data: {
    breadcrumb: 'Persistent Volume Claims',
    parent: CONFIG_ROUTE,
  },
};

const PERSISTENTVOLUMECLAIM_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: PersistentVolumeClaimDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: PERSISTENTVOLUMECLAIM_LIST_ROUTE,
  },
};

@NgModule({
  imports: [
    RouterModule.forChild([
      PERSISTENTVOLUMECLAIM_LIST_ROUTE,
      PERSISTENTVOLUMECLAIM_DETAIL_ROUTE,
      DEFAULT_ACTIONBAR,
    ]),
  ],
  exports: [RouterModule],
})
export class PersistentVolumeClaimRoutingModule {}
