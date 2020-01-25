import {Component, OnInit} from '@angular/core';
import {AuthService} from '../shared/auth.service';
import {CacheService} from '../shared/cache.service';
import {TranslateService} from '@ngx-translate/core';
import {Router} from '@angular/router';
import {StorageService} from '../shared/storage.service';
import {SideNavCollapseStorage} from '../shared/shared.const';

@Component({
  selector: 'app-sidenav-namespace',
  template: `
      <clr-vertical-nav [clrVerticalNavCollapsible]="true" [(clrVerticalNavCollapsed)]="collapsed">
          <a clrVerticalNavLink [title]="'MENU.DEPARTMENT' | translate"
             routerLink="/portal/namespace/{{cacheService.currentNamespace?.id}}/overview" routerLinkActive="active">
              <clr-icon clrVerticalNavIcon shape="help-info"></clr-icon>
              {{'MENU.DEPARTMENT' | translate}}
          </a>
          <a clrVerticalNavLink [title]="'MENU.PRODUCT' | translate"
             routerLink="/portal/namespace/{{cacheService.currentNamespace?.id}}/app"
             routerLinkActive="active">
              <clr-icon clrVerticalNavIcon shape="applications"></clr-icon>
              {{'MENU.PRODUCT' | translate}}
          </a>
          <a clrVerticalNavLink title="APIKeys"
             routerLink="/portal/namespace/{{cacheService.currentNamespace?.id}}/apikey"
             routerLinkActive="active" *ngIf="this.authService.config['enableApiKeys'] && (authService.currentNamespacePermission.apiKey.read || authService.currentUser.admin)">
              <clr-icon clrVerticalNavIcon shape="key"></clr-icon>
              APIKeys
          </a>
          <a clrVerticalNavLink title="Webhooks"
             routerLink="/portal/namespace/{{cacheService.currentNamespace?.id}}/webhook"
             routerLinkActive="active"
             *ngIf="authService.currentNamespacePermission.webHook.read || authService.currentUser.admin">
              <clr-icon clrVerticalNavIcon shape="pin"></clr-icon>
              Webhooks
          </a>
          <a clrVerticalNavLink [title]="'MENU.MEMBER' | translate"
             *ngIf="authService.currentNamespacePermission.namespaceUser.read || authService.currentUser.admin"
             routerLink="/portal/namespace/{{cacheService.currentNamespace?.id}}/users"
             routerLinkActive="active">
              <clr-icon clrVerticalNavIcon shape="user"></clr-icon>
              {{'MENU.MEMBER' | translate}}
          </a>
          <a href="javascript:;" style="flex: 1;"></a>
          <app-sidenav-footer></app-sidenav-footer>
      </clr-vertical-nav>
  `
})

export class SidenavNamespaceComponent implements OnInit {
  constructor(public authService: AuthService,
              public cacheService: CacheService,
              public translate: TranslateService,
              public storage: StorageService) {
  }

  _collapsed = false;
  get collapsed() {
    return this._collapsed;
  }
  set collapsed(value: boolean) {
    this._collapsed = value;
    this.storage.save(SideNavCollapseStorage, value);
  }

  ngOnInit() {
    this._collapsed = this.storage.get(SideNavCollapseStorage) !== 'false';
  }
}
