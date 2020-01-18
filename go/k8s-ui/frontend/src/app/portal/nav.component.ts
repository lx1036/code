import {Component, OnInit} from '@angular/core';
import {CacheService} from '../shared/cache.service';
import {AuthService} from '../shared/auth.service';
import {Notification, NotificationLog} from '../shared/model/v1/notification';
import {TranslateService} from '@ngx-translate/core';
import {Namespace} from '../shared/model/v1/namespace';
import {Router} from '@angular/router';

@Component({
  selector: 'app-nav',
  template: `
    <header style="background-color: #1D2143" class="header">
      <div class="branding" style="min-width: auto">
        <a routerLink="/portal/namespace/{{cacheService.currentNamespace?.id}}/app" class="nav-link">
          <img src="assets/images/wayne-logo.svg" width="60px" alt="">
          <span class="title">{{getTitle()}}</span>
        </a>
      </div>

      <div class="header-actions">
        <app-dropdown size="small">
          <clr-icon ref="javascript:void(0)" shape="bell" [class.has-badge]="mind" style="margin-right: 5px"></clr-icon>
          <app-dropdown-item>
            <ng-container *ngIf="notificationLogs && notificationLogs.length > 0">
              <div *ngFor="let notificationLog of notificationLogs" ref="javascript:void(0)"
                   (click)="showNotification(notificationLog)" style="white-space: nowrap;">
                <label class="label label-info" [class.label-info]="notificationLog.isReaded"
                       [class.label-warning]="!notificationLog.isReaded">{{(notificationLog.isReaded ? 'MESSAGE.READED' : 'MESSAGE.UNREAD') | translate}}</label>
                {{notificationLog.notification.from.name}} {{'MESSAGE.SEND' | translate}}《{{notificationLog.notification.title}}
                》{{'OF' | translate}}{{notificationLog.notification.type}}
              </div>
            </ng-container>
            <span *ngIf="!notificationLogs || notificationLogs.length === 0">{{'MESSAGE.NONE' | translate}}</span>
          </app-dropdown-item>
        </app-dropdown>

        <app-dropdown size="small">
          <clr-icon shape="world" style="margin-right: 5px"></clr-icon>
          {{showLang(currentLang)}}
          <clr-icon shape="caret down" size="12" style="margin-left: 5px;"></clr-icon>
          <app-dropdown-item>
            <ng-container *ngFor="let lang of translate.getLangs()">
              <span (click)="changeLang(lang)">{{showLang(lang)}}</span>
            </ng-container>
          </app-dropdown-item>
        </app-dropdown>

        <app-dropdown [size]="authService.currentUser?.namespaces.length > 10 ? 'middle' : 'small'">
          <clr-icon shape="organization" style="margin-right: 5px"></clr-icon>
          {{cacheService.currentNamespace?.name}}
          <clr-icon shape="caret down" size="12" style="margin-left: 5px;"></clr-icon>
          <app-dropdown-item [title]="authService.currentUser?.namespaces.length > 10 ? '部门' : ''">
            <span style="white-space: nowrap;" *ngFor="let n of authService.currentUser?.namespaces"
                  (click)='switchNamespace(n)'>{{n.name}}</span>
          </app-dropdown-item>
        </app-dropdown>

        <app-dropdown size="small" last>
          <clr-icon shape="user" style="margin-right: 5px"></clr-icon>
          {{authService.currentUser?.display}}
          <clr-icon shape="caret down" size="12" style="margin-left: 5px;"></clr-icon>
          <app-dropdown-item>
            <span (click)="goBack()" *ngIf="authService.currentUser.admin">{{'ACCOUNT.GO_BACK' | translate}}</span>
            <span (click)="logout()">{{'ACCOUNT.LOGOUT' | translate}}</span>
          </app-dropdown-item>
        </app-dropdown>
      </div>
    </header>

    <clr-modal [(clrModalOpen)]="notificationModal" [clrModalSize]="'xl'" [clrModalClosable]="false">
      <a class="modal-title"><span class="label label-info">{{notification.type}}</span> {{notification.title}}</a>
      <div class="modal-body">
        <markdown [data]="notification.message"></markdown>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-primary"
                (click)="closeNotification()">{{'BUTTON.CONFIRM' | translate}}</button>
      </div>
    </clr-modal>
  `
})

export class NavComponent implements OnInit {
  notificationLogs: NotificationLog[];
  mind = false;
  currentLang: string;
  notificationModal = false;
  notification: Notification = new Notification();

  constructor(private router: Router,
              public cacheService: CacheService,
              public authService: AuthService,
              public translate: TranslateService) {
  }

  ngOnInit() {
  }

  getTitle() {
    const imagePrefix = this.authService.config['system.title'];
    return imagePrefix ? imagePrefix : 'Kubernetes-UI';
  }

  showNotification(notificationLog: NotificationLog) {

  }

  closeNotification() {

  }

  showLang(lang: string): string {
    switch (lang) {
      case 'en':
        return 'English';
      case 'zh-Hans':
        return '中文简体';
      default:
        return '';
    }
  }

  changeLang(lang: string) {

  }

  switchNamespace(namespace: Namespace) {
    this.router.navigateByUrl(`/portal/namespace/${namespace.id}/app`).then();
  }

  goBack() {

  }

  logout() {
  }

}
