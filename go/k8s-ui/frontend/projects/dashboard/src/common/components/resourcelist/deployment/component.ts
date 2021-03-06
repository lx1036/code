

import {HttpParams} from '@angular/common/http';
import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  ComponentFactoryResolver,
  Input,
} from '@angular/core';
import {Deployment, DeploymentList, Event, Metric} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-deployment-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DeploymentListComponent extends ResourceListWithStatuses<DeploymentList, Deployment> {
  @Input() endpoint = EndpointManager.resource(Resource.deployment, true).list();
  @Input() showMetrics = false;
  cumulativeMetrics: Metric[];

  constructor(
    private readonly deployment_: NamespacedResourceService<DeploymentList>,
    notifications: NotificationsService,
    resolver: ComponentFactoryResolver,
    cdr: ChangeDetectorRef,
  ) {
    super('deployment', notifications, cdr, resolver);
    this.id = ListIdentifier.deployment;
    this.groupId = ListGroupIdentifier.workloads;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.timelapse, 'kd-muted', this.isInPendingState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  getResourceObservable(params?: HttpParams): Observable<DeploymentList> {
    return this.deployment_.get(this.endpoint, undefined, undefined, params);
  }

  map(deploymentList: DeploymentList): Deployment[] {
    this.cumulativeMetrics = deploymentList.cumulativeMetrics;
    return deploymentList.deployments;
  }

  isInErrorState(resource: Deployment): boolean {
    return resource.pods.warnings.length > 0;
  }

  isInPendingState(resource: Deployment): boolean {
    return (
      resource.pods.warnings.length === 0 &&
      (resource.pods.pending > 0 || resource.pods.running !== resource.pods.desired)
    );
  }

  isInSuccessState(resource: Deployment): boolean {
    return (
      resource.pods.warnings.length === 0 &&
      resource.pods.pending === 0 &&
      resource.pods.running === resource.pods.desired
    );
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'pods', 'created', 'images'];
  }

  hasErrors(deployment: Deployment): boolean {
    return deployment.pods.warnings.length > 0;
  }

  getEvents(deployment: Deployment): Event[] {
    return deployment.pods.warnings;
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
