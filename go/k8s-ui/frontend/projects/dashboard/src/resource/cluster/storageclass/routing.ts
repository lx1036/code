

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {CLUSTER_ROUTE} from '../../cluster/routing';

import {StorageClassDetailComponent} from './detail/component';
import {StorageClassListComponent} from './list/component';

const STORAGECLASS_LIST_ROUTE: Route = {
  path: '',
  component: StorageClassListComponent,
  data: {
    breadcrumb: 'Storage Classes',
    parent: CLUSTER_ROUTE,
  },
};

const STORAGECLASS_DETAIL_ROUTE: Route = {
  path: ':resourceName',
  component: StorageClassDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: STORAGECLASS_LIST_ROUTE,
  },
};

@NgModule({
  imports: [
    RouterModule.forChild([STORAGECLASS_LIST_ROUTE, STORAGECLASS_DETAIL_ROUTE, DEFAULT_ACTIONBAR]),
  ],
  exports: [RouterModule],
})
export class StorageClassRoutingModule {}
