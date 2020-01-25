import {AfterViewInit, Component, ElementRef, Inject, OnDestroy, OnInit} from '@angular/core';
import {DOCUMENT} from '@angular/common';
import {NamespaceClient} from '../shared/client/v1/kubernetes/namespace';
import {CacheService} from '../shared/cache.service';
import {MessageHandlerService} from '../shared/message-handler.service';
import {AuthService} from '../shared/auth.service';

const showState = {
  name: {hidden: false},
  description: {hidden: false},
  create_time: {hidden: false},
  create_user: {hidden: false},
  action: {hidden: false}
};

interface ClusterCard {
  name: string;
  state: boolean;
}

@Component({
  selector: 'app-portal-app',
  template: `
    <div class="content-area" style="position: relative">
      <div class="clr-row">
        <div class="clr-col-lg-12 clr-col-md-12 clr-col-sm-12 clr-col-xs-12">
          <div class="clr-row flex-items-xs-between flex-items-xs-top" style="padding-left: 15px; padding-right: 15px;">
            <div class="cluster-outline" style="display: flex; flex-wrap: wrap;width: 100%;">
              <app-card *ngIf="authService.currentNamespacePermission.app.create || authService.currentUser.admin" (click)="openModal()" style="cursor: pointer;">
                <div style="flex: 1;display: flex; justify-content: center; align-items: center; color: #377aec; font-size: 20px;">
                  <svg style="width: 16px; height: 16px;fill: #377aec; margin-right: 5px;" viewBox="0, 0, 40 , 40" xmlns="http://www.w3.org/2000/svg">
                    <rect x="0" y="18.5" width="40" height="3" rx="1.5" ry="1.5"></rect>
                    <rect x="18.5" y="0" width="3" height="40" rx="1.5" ry="1.5"></rect>
                  </svg>
                  {{'TITLE.CREATE_APP' | translate}}
                </div>
              </app-card>
              <ng-container *ngFor="let cluster of clusters; let i = index">
                <app-card>

                </app-card>
              </ng-container>
            </div>

            <p class="card-show-p"></p>

            <app-box>
                <div class="table-search" style="padding: 0 15px;">
                    <div class="table-search-left">

                    </div>
                    <div class="table-search-right">

                    </div>
                </div>
            </app-box>
          </div>
        </div>
      </div>
    </div>

    <app-sidenav-namespace style="display: flex; order: -1"></app-sidenav-namespace>
<!--    <create-edit-app (create)="createApp($event)"></create-edit-app>-->
  `,
})
export class AppComponent implements OnInit, OnDestroy, AfterViewInit {
  showList: any[] = [];
  showState: object = showState;
  starredFilter: boolean;
  starredInherit: boolean; // starredInherit 用来传递给list
  resources: object = {};
  clusters: ClusterCard[] = [];
  allowNumber = 10;

  constructor(private namespaceClient: NamespaceClient,
              private cacheService: CacheService,
              @Inject(DOCUMENT) private document: any,
              private element: ElementRef,
              private messageHandlerService: MessageHandlerService,
              public authService: AuthService) {}



  ngOnInit() {
    this.initShow();
    this.starredFilter = localStorage.getItem('starred') === 'true';
    this.starredInherit = this.starredFilter;
    this.namespaceClient.getResourceUsage(this.cacheService.namespaceId).subscribe(response => {
      this.resources = response.data;
      Object.getOwnPropertyNames(this.resources).forEach(cluster => {
        this.clusters.push({name: cluster, state: false});
      });

      this.allowNumber = this.getClusterMaxNumber();
      for (let i = 0; i < this.allowNumber - 1; i++) {
        setTimeout(((idx) => {
          if (this.clusters[idx]) {
            this.clusters[idx].state = true;
          }
        }).bind(this, i), 200 * i);
      }
    }, error => this.messageHandlerService.handleError(error));
  }

  getClusterMaxNumber() {
    return Math.floor(this.element.nativeElement.querySelector('.cluster-outline').offsetWidth / 255);
  }

  initShow() {
    this.showList = [];
    Object.keys(this.showState).forEach(key => {
      if (!this.showState[key].hidden) {
        this.showList.push(key);
      }
    });
  }

  ngAfterViewInit(): void {
  }

  ngOnDestroy(): void {
  }
}
