import {Component, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {ClrDatagridStateInterface} from '@clr/angular';
import {PageState} from '../../shared/page-state';
import {Notification} from '../../shared/model/v1/notification';
import {NotificationService} from '../../shared/notification.service';
import {MessageHandlerService} from '../../shared/message-handler.service';


@Component({
  selector: 'app-create-notification',
  template: `
    <clr-modal [(clrModalOpen)]="opened" [clrModalSize]="'xl'">
      <h3 class="modal-title">创建通知</h3>
      <div class="modal-body">
        <form>
          <section class="form-block">
            <div class="form-group">
              <label class="clr-col-md-3 form-group-label-override required">标题</label>
              <label class="tooltip tooltip-validation tooltip-md tooltip-bottom-left">
                <input type="text" size="1024" [(ngModel)]="notify.title" [ngModelOptions]="{standalone: true}" required>
              </label>
            </div>
            <div class="form-group">
              <label class="clr-col-md-3 form-group-label-override required">类型</label>
              <div class="select">
                <select [(ngModel)]="notify.type" [ngModelOptions]="{standalone: true}" type="string">
                  <option [ngValue]="'公告'">公告</option>
                  <option [ngValue]="'警告'">警告</option>
                </select>
              </div>
            </div>
          </section>
        </form>
        <textarea [(ngModel)]="notify.message"></textarea>
        <markdown [data]="notify.message"></markdown>
      </div>
      <div class="modal-footer">
        <button class="btn btn-outline" type="button" (click)="cancelCreateNotification()">{{'BUTTON.CANCEL' | translate}}</button>
        <button type="button" class="btn btn-primary" [disabled]="!isValid()" (click)="createNotification()">创建</button>
      </div>
    </clr-modal>
  `
})
export class CreateNotificationComponent implements OnInit {
  opened = false;
  notify: Notification = new Notification();
  @Output() create = new EventEmitter<boolean>();

  constructor(private notificationService: NotificationService,
              private messageHandlerService: MessageHandlerService) {}

  ngOnInit() {
  }

  newOrEditNotification() {
    this.opened = true;
    this.notify = new Notification();
  }

  cancelCreateNotification() {
    this.opened = false;
    this.notify = new Notification();
  }

  createNotification() {
    this.notificationService.create(this.notify)
    .subscribe(
      response => {
        // this.messageHandlerService.showSuccess(`创建通知成功！`);
        this.create.emit(true);
        this.opened = false;
      },
      error => this.messageHandlerService.handleError(error)
    );
  }

  isValid(): boolean {
    return this.notify.title.length > 0;
  }
}


@Component({
  selector: 'app-list-notification',
  template: `
    <clr-datagrid (clrDgRefresh)="refresh($event)">
      <clr-dg-column>
        <ng-container *clrDgHideableColumn="{hidden: false}">ID</ng-container>
      </clr-dg-column>
      <clr-dg-column style="min-width: 206px;">
        <ng-container *clrDgHideableColumn="{hidden: false}">类型</ng-container>
      </clr-dg-column>
      <clr-dg-column style="min-width: 206px;">
        <ng-container *clrDgHideableColumn="{hidden: false}">标题</ng-container>
      </clr-dg-column>
      <clr-dg-column style="min-width: 206px;">
        <ng-container *clrDgHideableColumn="{hidden: false}">创建人</ng-container>
      </clr-dg-column>
      <clr-dg-column style="min-width: 206px;">
        <ng-container *clrDgHideableColumn="{hidden: false}">状态</ng-container>
      </clr-dg-column>
      <clr-dg-column style="min-width: 206px;">
        <ng-container *clrDgHideableColumn="{hidden: false}">{{'TITLE.CREATE_TIME' | translate}}</ng-container>
      </clr-dg-column>

      <clr-dg-row *ngFor="let n of notifications; let i = index" [clrDgItem]="notification">
        <clr-dg-action-overflow>
          <button class="action-item" (click)="showPushNotifyModal(n)">广播</button>
        </clr-dg-action-overflow>
        <clr-dg-cell>{{n.id}}</clr-dg-cell>
        <clr-dg-cell class="copy">{{n.type}}</clr-dg-cell>
        <clr-dg-cell class="copy">{{n.title}}</clr-dg-cell>
        <clr-dg-cell class="copy">{{n.user.name}}</clr-dg-cell>
        <clr-dg-cell class="copy">
          <div *ngIf="n.isPublished">已广播</div>
          <div *ngIf="!n.isPublished">未广播</div>
        </clr-dg-cell>
        <clr-dg-cell>{{n.createTime | date:'yyyy-MM-dd HH:mm:ss'}}</clr-dg-cell>
      </clr-dg-row>

      <clr-dg-footer>
<!--        <app-paginate-->
<!--          [(currentPage)]="currentPage"-->
<!--          [total]="page.totalCount"-->
<!--          [pageSizes]="[10, 20, 50]"-->
<!--          (sizeChange)="pageSizeChange($event)"></app-paginate>-->
      </clr-dg-footer>
    </clr-datagrid>

    <clr-modal [(clrModalOpen)]="notificationModal" [clrModalSize]="'xl'">
      <h3 class="modal-title">是否广播如下内容:</h3>
      <div class="modal-body">
        <markdown ngPreserveWhitespaces [data]="notification.message"></markdown>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-outline" (click)="cancelPushNotify()">{{'BUTTON.CANCEL' | translate}}</button>
        <button type="button" class="btn btn-primary" (click)="pushNotify()">广播</button>
      </div>
    </clr-modal>
  `
})
export class ListNotificationComponent implements OnInit {
  @Input() notifications: Notification[];
  @Output() paginate = new EventEmitter<ClrDatagridStateInterface>();

  notification: Notification = new Notification();
  notificationModal = false;
  state: ClrDatagridStateInterface;

  constructor() {
  }

  ngOnInit() {
  }

  showPushNotifyModal(notify: Notification) {
    this.notification = notify;
    this.notificationModal = true;
  }

  refresh(state: ClrDatagridStateInterface) {
    this.state = state;
    this.paginate.emit(state);
  }

  pushNotify() {

  }

  cancelPushNotify() {

  }
}



@Component({
  selector: 'app-notification',
  template: `
    <div class="content-container" style="position: relative">
      <div class="content-area">
        <div class="clr-row">
          <div class="clr-col-lg-12 clr-col-md-12 clr-col-sm-12 clr-col-xs-12">
            <div class="clr-row flex-items-xs-between flex-items-xs-top" style="padding-left: 15px; padding-right: 15px;">
              <h2 class="header-title">{{ resourceLabel }}列表</h2>
            </div>

            <div class="clr-row flex-items-xs-between" style="height:32px;">
              <div class="option-left">
                <button class="btn btn-link" (click)="openCreateModal()"><clr-icon shape="add"></clr-icon>创建通知</button>
                <app-create-notification (create)="createNotification($event)"></app-create-notification>
              </div>

              <app-list-notification [notifications]="notifications" (updated)="updateNotification($event)" (paginate)="retrieve($event)"></app-list-notification>
            </div>
          </div>
        </div>
      </div>
    </div>
  `
})
export class NotificationComponent implements OnInit {
  resourceLabel = '通知';
  notifications: Notification[];
  pageState: PageState = new PageState();


  @ViewChild(CreateNotificationComponent, { static: false })
  createNotificationComponent: CreateNotificationComponent;

  constructor(private notificationService: NotificationService, ) {
  }

  ngOnInit() {
  }

  openCreateModal(): void {
    this.createNotificationComponent.newOrEditNotification();
  }

  createNotification(created: boolean) {
    if (created) {
      this.retrieve();
    }
  }

  updateNotification(updated: boolean) {
  }

  retrieve(state?: ClrDatagridStateInterface): void {
    this.notificationService.query().subscribe(
      (response: {data: Notification[]}) => {
        this.notifications = response.data;
      }, error => {});
  }

}
