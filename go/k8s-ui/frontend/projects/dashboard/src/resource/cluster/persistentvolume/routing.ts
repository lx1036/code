

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {CLUSTER_ROUTE} from '../routing';

import {PersistentVolumeDetailComponent} from './detail/component';
import {PersistentVolumeListComponent} from './list/component';

const PERSISTENTVOLUME_LIST_ROUTE: Route = {
  path: '',
  component: PersistentVolumeListComponent,
  data: {
    breadcrumb: 'Persistent Volumes',
    parent: CLUSTER_ROUTE,
  },
};

const PERSISTENTVOLUME_DETAIL_ROUTE: Route = {
  path: ':resourceName',
  component: PersistentVolumeDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: PERSISTENTVOLUME_LIST_ROUTE,
  },
};

@NgModule({
  imports: [
    RouterModule.forChild([
      PERSISTENTVOLUME_LIST_ROUTE,
      PERSISTENTVOLUME_DETAIL_ROUTE,
      DEFAULT_ACTIONBAR,
    ]),
  ],
  exports: [RouterModule],
})
export class PersistentVolumeRoutingModule {}
