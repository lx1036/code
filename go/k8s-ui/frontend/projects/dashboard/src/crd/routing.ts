

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';

import {CRDDetailComponent} from './detail/component';
import {CRDListComponent} from './list/component';
import {DEFAULT_ACTIONBAR, PIN_DEFAULT_ACTIONBAR} from '../common/components/actionbars/routing';
import {CRDObjectDetailComponent} from './crdobject/component';
import {SCALE_DEFAULT_ACTIONBAR} from '../common/components/actionbars/routing';

const CRD_LIST_ROUTE: Route = {
  path: '',
  children: [
    {
      path: '',
      component: CRDListComponent,
      data: {breadcrumb: 'Custom Resource Definitions'},
    },
    DEFAULT_ACTIONBAR,
  ],
};

const CRD_DETAIL_ROUTE: Route = {
  path: '',
  children: [
    {
      path: ':crdName',
      component: CRDDetailComponent,
      data: {breadcrumb: '{{ crdName }}', parent: CRD_LIST_ROUTE.children[0]},
    },
    PIN_DEFAULT_ACTIONBAR,
  ],
};

const CRD_NAMESPACED_OBJECT_DETAIL_ROUTE: Route = {
  path: ':crdName/:namespace/:objectName',
  children: [
    {
      path: '',
      component: CRDObjectDetailComponent,
      data: {
        breadcrumb: '{{ objectName }}',
        routeParamsCount: 2,
        parent: CRD_DETAIL_ROUTE.children[0],
      },
    },
    SCALE_DEFAULT_ACTIONBAR,
  ],
};

const CRD_CLUSTER_OBJECT_DETAIL_ROUTE: Route = {
  path: ':crdName/:objectName',
  children: [
    {
      path: '',
      component: CRDObjectDetailComponent,
      data: {
        breadcrumb: '{{ objectName }}',
        routeParamsCount: 1,
        parent: CRD_DETAIL_ROUTE.children[0],
      },
    },
    SCALE_DEFAULT_ACTIONBAR,
  ],
};

@NgModule({
  imports: [
    RouterModule.forChild([
      CRD_LIST_ROUTE,
      CRD_DETAIL_ROUTE,
      CRD_NAMESPACED_OBJECT_DETAIL_ROUTE,
      CRD_CLUSTER_OBJECT_DETAIL_ROUTE,
    ]),
  ],
})
export class CRDRoutingModule {}
