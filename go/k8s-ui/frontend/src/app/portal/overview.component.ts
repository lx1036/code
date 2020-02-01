import {Component, OnInit} from '@angular/core';
import {CacheService} from '../shared/cache.service';
import {AppService} from '../shared/app.service';
import {HttpErrorResponse} from '@angular/common/http';
import {
  KubeApiTypeConfigMap,
  KubeApiTypeCronJob,
  KubeApiTypeDaemonSet,
  KubeApiTypeDeployment, KubeApiTypePersistentVolumeClaim, KubeApiTypeSecret, KubeApiTypeService,
  KubeApiTypeStatefulSet
} from '../shared/shared.const';

export interface NamespaceStatistics {
  ConfigMap: number;
  CronJob: number;
  DaemonSet: number;
  Deployment: number;
  PersistentVolumeClaim: number;
  Secret: number;
  Service: number;
  StatefulSet: number;
}

/**
 * @see https://clarity.design/
 */
@Component({
  selector: 'app-overview',
  template: `
    <div class="clr-row" style="padding-left: 1%">
        <div class="clr-col-lg-12 clr-col-md-12 clr-col-sm-12 clr-col-xs-12">
            <form>
                <section class="form-block form-box">
                    <div class="clr-row group">
                        <div class="clr-col-sm-2">
                            {{'TITLE.DEPARTMENT_NAME' | translate}}
                        </div>
                        <div class="clr-col-sm-10">
                            {{cacheService?.namespace.name}}
                        </div>
                    </div>

                    <div class="clr-row group">
                        <div class="clr-col-sm-2">
                            {{'TITLE.CREATE_TIME' | translate}}
                        </div>
                        <div class="clr-col-sm-10">
                            {{cacheService?.namespace.createTime | date:'yyyy-MM-dd HH:mm:ss'}}
                        </div>
                    </div>

                    <div class="clr-row group">
                        <div class="clr-col-sm-2">
                            <div class="flex" style="height:70px;flex-direction: column; justify-content:space-around;">
                                <span>{{'TITLE.CPU_USAGE' | translate}}</span>
                                <span>{{'TITLE.MEMORY_USAGE' | translate}}</span>
                            </div>
                        </div>
                        <div class="clr-col-sm-10">
                            <ng-container *ngFor="let cluster of clusters; let i = index">
                                <div *ngIf="i < showNumber" style="display: inline-flex;height: 70px;width: 150px; margin-right: 10px; flex-direction: column; justify-content: space-around;">
                                    <app-progress></app-progress>
                                </div>
                            </ng-container>
                            <div class="cluster-more" (click)="showMoreCluster()" *ngIf="clusters.length > showNumber">
                                <clr-icon shape="angle-double"></clr-icon>
                            </div>
                        </div>
                    </div>

                    <div class="clr-row group">
                        <app-card>
                            <div class="card-title">{{'MENU.DEPLOYMENT' | translate}}</div>
                            <p class="card-text">
                                {{resourceCount(deployment)}}
                            </p>
                        </app-card>
                        <app-card>
                            <div class="card-title">{{'MENU.STATEFULSET' | translate}}</div>
                            <p class="card-text">
                                {{resourceCount(statefulSet)}}
                            </p>
                        </app-card>
                        <app-card>
                            <div class="card-title">{{'MENU.DAEMONSET' | translate}}</div>
                            <p class="card-text">
                                {{resourceCount(daemonSet)}}
                            </p>
                        </app-card>
                        <app-card>
                            <div class="card-title">{{'MENU.CRONJOB' | translate}}</div>
                            <p class="card-text">
                                {{resourceCount(cronJob)}}
                            </p>
                        </app-card>
                    </div>

                    <div class="clr-row group">
                        <app-card>
                            <div class="card-title">{{'MENU.SERVICE' | translate}}</div>
                            <p class="card-text">
                                {{resourceCount(service)}}
                            </p>
                        </app-card>
                        <app-card>
                            <div class="card-title">{{'MENU.CONFIGMAP' | translate}}</div>
                            <p class="card-text">
                                {{resourceCount(configMap)}}
                            </p>
                        </app-card>
                        <app-card>
                            <div class="card-title">{{'MENU.SECRET' | translate}}</div>
                            <p class="card-text">
                                {{resourceCount(secret)}}
                            </p>
                        </app-card>
                        <app-card>
                            <div class="card-title">{{'MENU.PVC' | translate}}</div>
                            <p class="card-text">
                                {{resourceCount(persistentVolumeClaim)}}
                            </p>
                        </app-card>
                    </div>
                </section>
            </form>
        </div>
    </div>
  `
})
export class OverviewComponent implements OnInit {
  clusters: string[] = [];
  showNumber = 10;
  namespaceStatistics: NamespaceStatistics;

  readonly deployment = KubeApiTypeDeployment;
  readonly cronJob = KubeApiTypeCronJob;
  readonly statefulSet = KubeApiTypeStatefulSet;
  readonly daemonSet = KubeApiTypeDaemonSet;
  readonly service = KubeApiTypeService;
  readonly configMap = KubeApiTypeConfigMap;
  readonly secret = KubeApiTypeSecret;
  readonly persistentVolumeClaim = KubeApiTypePersistentVolumeClaim;

  constructor(public cacheService: CacheService,
              private appService: AppService) {}

  ngOnInit() {
    this.initResourceCount();
    this.initResourceUsage();
  }

  initResourceCount() {
    this.appService.listResourceCount(this.cacheService.namespaceId).subscribe((response: {data: NamespaceStatistics}) => {
      this.namespaceStatistics = response.data;
    }, (error: HttpErrorResponse) => {

    });
  }

  initResourceUsage() {

  }

  resourceCount(resource: string): number {
    if (this.namespaceStatistics) {
      return this.namespaceStatistics[resource] || 0;
    }

    return 0;
  }

  showMoreCluster() {

  }
}

