

import {HttpParams} from '@angular/common/http';
import {ChangeDetectionStrategy, ChangeDetectorRef, Component, Input} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {CRDObject, CRDObjectList} from '@api/backendapi';
import {Observable} from 'rxjs';
import {map, takeUntil} from 'rxjs/operators';
import {ResourceListBase} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-crd-object-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class CRDObjectListComponent extends ResourceListBase<CRDObjectList, CRDObject> {
  @Input() endpoint: string;
  @Input() namespaced = false;

  constructor(
    private readonly crdObject_: NamespacedResourceService<CRDObjectList>,
    private readonly activatedRoute_: ActivatedRoute,
    notifications: NotificationsService,
    cdr: ChangeDetectorRef,
  ) {
    super(
      activatedRoute_.params.pipe(map(params => `customresourcedefinition/${params.crdName}`)),
      notifications,
      cdr,
    );
    this.id = ListIdentifier.crdObject;
    this.groupId = ListGroupIdentifier.none;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    this.activatedRoute_.params.pipe(takeUntil(this.unsubscribe_)).subscribe(params => {
      this.endpoint = EndpointManager.resource(Resource.crd, true).child(
        params.crdName,
        Resource.crdObject,
      );
    });
  }

  getResourceObservable(params?: HttpParams): Observable<CRDObjectList> {
    return this.crdObject_.get(this.endpoint, undefined, undefined, params);
  }

  map(crdObjectList: CRDObjectList): CRDObject[] {
    return crdObjectList.items;
  }

  getDisplayColumns(): string[] {
    return ['name', 'namespace', 'created'];
  }

  areMultipleNamespacesSelected(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
