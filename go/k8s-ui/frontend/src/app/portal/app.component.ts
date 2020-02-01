import {
  AfterViewInit,
  Component,
  ElementRef,
  EventEmitter,
  Inject,
  OnDestroy,
  OnInit,
  Output,
  ViewChild
} from '@angular/core';
import {DOCUMENT} from '@angular/common';
import {NamespaceClient} from '../shared/client/v1/kubernetes/namespace';
import {CacheService} from '../shared/cache.service';
import {MessageHandlerService} from '../shared/message-handler.service';
import {AuthService} from '../shared/auth.service';
import {App} from '../shared/model/v1/app';
import {AppService} from '../shared/app.service';
import {NgForm} from '@angular/forms';

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

export const enum ActionType {
  ADD_NEW, EDIT
}

@Component({
  selector: 'app-create-edit-app',
  template: `
    <clr-modal [(clrModalOpen)]="createAppOpened">
      <h3 class="modal-title">{{appTitle}}</h3>
      <div class="modal-body">
        <form #appForm="ngForm" clrForm clrLayout="horizontal">
          <section class="form-block">
            <clr-input-container>
              <label class="required">{{'TITLE.NAME' | translate}}</label>
              <input clrInput type="text" id="app_name" [(ngModel)]="app.name" name="app_name" size="32" required
                     [placeholder]="'PLACEHOLDER.APP_NAME' | translate" [readonly]="actionType==1" pattern="[a-z]([-a-z0-9]*[a-z0-9])?" maxlength="32" (keyup)='handleValidation()'>
              <clr-control-helper>
                <span style="color: red;" *ngIf="!isNameValid">{{'RULE.REGEXT' | translate}}[a-z]([-a-z0-9]*[a-z0-9])?</span>
              </clr-control-helper>
              <clr-control-error>{{'RULE.REGEXT' | translate}}[a-z]([-a-z0-9]*[a-z0-9])?</clr-control-error>
            </clr-input-container>
            <div hidden class="form-group" style="padding-left: 135px;">
              <label for="app_metadata" class="clr-col-md-3 form-group-label-override">{{'TITLE.METADATA' | translate}}</label>
              <textarea id="app_metadata" [(ngModel)]="app.metaData" name="app_metadata" rows="3"></textarea>
            </div>
            <clr-textarea-container>
              <label>{{'TITLE.DESCRIPTION' | translate}}</label>
              <textarea clrTextarea id="app_description" [(ngModel)]="app.description" name="app_description" rows="3"> </textarea>
            </clr-textarea-container>
            <div class="modal-footer">
              <button type="button" class="btn btn-outline" (click)="onCancel()">{{'BUTTON.CANCEL' | translate}}</button>
              <button type="button" class="btn btn-primary" [disabled]="!isValid" (click)="onSubmit()">{{'BUTTON.CONFIRM' | translate}}</button>
            </div>
          </section>
        </form>
      </div>
    </clr-modal>
  `
})
export class CreateEditAppComponent implements OnInit {
  createAppOpened: boolean;
  @Output() create = new EventEmitter<boolean>();
  appTitle: string;
  isNameValid: boolean;
  app: App = new App();
  actionType: ActionType;
  @ViewChild('appForm', { static: true }) currentForm: NgForm;
  checkOnGoing = false;
  isSubmitOnGoing = false;

  constructor(private appService: AppService, ) {}

  ngOnInit() {
  }

  public get isValid(): boolean {
    return this.currentForm &&
      this.currentForm.valid &&
      !this.isSubmitOnGoing &&
      this.isNameValid &&
      !this.checkOnGoing;
  }

  newOrEditApp(id?: number) {

  }

  onCancel() {

  }

  handleValidation() {

  }

  onSubmit() {
    switch (this.actionType) {
      case ActionType.ADD_NEW:
        this.appService.create(this.app).subscribe(response => {}, err => {});
        break;
      case ActionType.EDIT:
        this.appService.update(this.app).subscribe(response => {}, err => {});
        break;
    }
  }
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

                <app-list-apps [apps]="changedApps"></app-list-apps>
            </app-box>
          </div>
        </div>
      </div>
    </div>

    <app-sidenav-namespace style="display: flex; order: -1"></app-sidenav-namespace>
    <app-create-edit-app (create)="createApp($event)"></app-create-edit-app>
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
  changedApps: App[];
  @ViewChild(CreateEditAppComponent, { static: false }) createEditApp: CreateEditAppComponent;

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

  openModal() {
    this.createEditApp.newOrEditApp();
  }

  createApp($event) {

  }

  editApp(app: App) {
    this.createEditApp.newOrEditApp(app.id);
  }
}
