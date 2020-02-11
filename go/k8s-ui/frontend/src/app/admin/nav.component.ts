import {Component, OnInit} from '@angular/core';
import {AuthService} from "../shared/auth.service";
import {LangChangeEvent, TranslateService} from "@ngx-translate/core";
import {StorageService} from "../shared/storage.service";
import {Router} from "@angular/router";
import {LoginTokenKey} from "../shared/shared.const";

@Component({
  selector: 'app-admin-nav',
  template: `
    <header>
      <div class="branding">
        <a routerLink="/admin/reportform/overview" class="nav-link">
          <img src="assets/images/wayne-logo.svg" width="60px" alt="">
          <span class="title">{{getTitle()}}</span>
        </a>
      </div>
      <div class="header-actions">
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
        <app-dropdown size="small">
          <clr-icon shape="user" style="margin-right: 5px"></clr-icon>
          {{authService.currentUser?.display}}
          <clr-icon shape="caret down" size="12" style="margin-left: 5px;"></clr-icon>
          <app-dropdown-item>
            <span (click)="goFront()" *ngIf="authService.currentUser.admin">返回前台</span>
            <span href="javascript:void(0)" (click)="logout()">注销登录</span>
          </app-dropdown-item>
        </app-dropdown>
      </div>
    </header>
  `
})

export class NavComponent implements OnInit {
  currentLang: string;

  constructor(public authService: AuthService,
              public translate: TranslateService,
              private storage: StorageService,
              private router: Router) {
  }

  ngOnInit() {
    this.currentLang = this.translate.currentLang;
    this.translate.onLangChange.subscribe((event: LangChangeEvent) => {
      this.currentLang = event.lang;
    })
  }

  goFront() {
    this.router.navigateByUrl("/");
  }

  logout() {
    this.storage.remove(LoginTokenKey)
  }

  changeLang(lang: string) {
    this.translate.use(lang);
    this.storage.save('lang', lang);
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

  getTitle() {
    const imagePrefix = this.authService.config['system.title'];
    return imagePrefix ? imagePrefix : 'Wayne';
  }
}
