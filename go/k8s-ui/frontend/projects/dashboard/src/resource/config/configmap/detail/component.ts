

import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {ConfigMapDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-config-map-detail',
  templateUrl: './template.html',
})
export class ConfigMapDetailComponent implements OnInit, OnDestroy {
  private configMapSubscription_: Subscription;
  private endpoint_ = EndpointManager.resource(Resource.configMap, true);
  configMap: ConfigMapDetail;
  isInitialized = false;

  constructor(
    private readonly configMap_: NamespacedResourceService<ConfigMapDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;
    const resourceNamespace = this.activatedRoute_.snapshot.params.resourceNamespace;

    this.configMapSubscription_ = this.configMap_
      .get(this.endpoint_.detail(), resourceName, resourceNamespace)
      .subscribe((d: ConfigMapDetail) => {
        this.configMap = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Config Map', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.configMapSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }

  getConfigMapData(cm: ConfigMapDetail): string {
    if (!cm) {
      return '';
    }

    return JSON.stringify(cm.data);
  }
}
