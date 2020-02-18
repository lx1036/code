import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {Router} from "@angular/router";
import {AdminSideNav, SideNavType} from "../../shared/sidenav.const";
import {SideNavCollapseStorage} from "../../shared/shared.const";
import {StorageService} from "../../shared/storage.service";
import {AuthService} from "../../shared/auth.service";
import {SideNavService} from "./sidenav.service";

@Component({
  selector: 'app-admin-sidenav',
  template: `
    <clr-vertical-nav [clrVerticalNavCollapsible]="true" [(clrVerticalNavCollapsed)]="collapsed">
      <ng-container *ngFor="let item of adminSideNav">
        <a clrVerticalNavLink *ngIf="item.type === sideNavType.NormalLink" [title]="item.a.title | translate" [routerLink]="item.a.link" [class.active]="getActive(item.a.link)">
          <clr-icon clrVerticalNavIcon [attr.shape]="item.a.icon.shape"></clr-icon>{{ item.a.text | translate }}
        </a>
        <div class="nav-divider" *ngIf="item.type === sideNavType.Divider"></div>
        <clr-vertical-nav-group *ngIf="item.type === sideNavType.GroupLink && (item.name !== 'edge-node' || this.authService.config['system.external-ip'])" routerLinkActive="active">
          <clr-icon clrVerticalNavIcon [title]="item.icon.title | translate" [class.is-solid]="item.icon.solid" [attr.shape]="item.icon.shape"></clr-icon>{{item.text | translate}}
          <clr-vertical-nav-group-children *clrIfExpanded="getExpand(item.links)">
            <ng-container *ngFor="let child of item.child">
              <a clrVerticalNavLink *ngIf="child.type !== sideNavType.Divider && (child.a.link !== 'apikey' || this.authService.config['enableApiKeys'])" [routerLink]="child.a.link" [class.active]="getActive(child.a.link, child.a.options)">
                <clr-icon *ngIf="child.a.icon" clrVerticalNavIcon [attr.shape]="child.a.icon.shape" [class.is-solid]="child.a.icon.solid"></clr-icon>{{child.a.text | translate}}
              </a>
              <div class="nav-divider" *ngIf="child.type === sideNavType.Divider"></div>
            </ng-container>
          </clr-vertical-nav-group-children>
        </clr-vertical-nav-group>
      </ng-container>
    </clr-vertical-nav>
  `,
  styleUrls: ["./sidenav.component.scss"]
})
export class SidenavComponent implements OnInit {
  public adminSideNav: any[];
  sideNavType = SideNavType;
  currentUrl = this.router.url;
  prefix = "/admin/";

  constructor(
    public authService: AuthService,
    public sideNavService: SideNavService,
    public router: Router,
    public cr: ChangeDetectorRef,
    public storage: StorageService,
  ) {
    this.adminSideNav = this.addMathLinks(AdminSideNav);
  }

  ngOnInit() {
    this._collapsed = this.storage.get(SideNavCollapseStorage) !== 'false';

    this.sideNavService.routerChange.subscribe(
      url => {
        this.currentUrl = url.split('?')[0];
        // 取消脏检查
        this.cr.detectChanges();
      }
    );
  }

  addMathLinks(sideNav: any[]) {
    sideNav.forEach(item => {
      if (item.type === SideNavType.GroupLink && item.child) {
        item.links = item.child.filter(child => child.a).map(child => `${this.prefix}${child.a.link}`);
      }
    });
    return sideNav;
  }

  public getActive(link: string, option?: any): boolean {
    if (option && option.exact !== undefined) {
      return option.exact ?
        this.currentUrl === `${this.prefix}${link}` :
        new RegExp(`${link}\\b`).test(this.currentUrl);
    } else {
      return this.currentUrl === `${this.prefix}${link}`;
    }
  }

  public getExpand(matchUrls: string[]): boolean {
    return new RegExp(matchUrls.map(url => `${url}\\b`).join('|')).test(this.currentUrl);
  }

  _collapsed = false;
  get collapsed() {
    return this._collapsed;
  }
  set collapsed(value: boolean) {
    this._collapsed = value;
    this.storage.save(SideNavCollapseStorage, value);
  }
}
