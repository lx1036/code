import {Component, Input, OnInit} from '@angular/core';
import {Subscription} from "rxjs";
import {Event, NavigationEnd, Router} from "@angular/router";
import {BreadcrumbService} from "./breadcrumb.service";

@Component({
  selector: 'app-shared-breadcrumb',
  template: `
    <ul class="breadcrumb">
      <li class="item" *ngFor="let url of urls; let last = last;" [class.active]="last || !friendlyName(url).avail">
        <a href="javascript:;" (click)="navigateTo(url, friendlyName(url).avail)">{{friendlyName(url).name}}</a>
      </li>
    </ul>
  `
})
export class BreadcrumbComponent implements OnInit {
  @Input() prefix = '';
  urls: string[] = [];
  private routerSubscription: Subscription;

  constructor(public router: Router, private breadcrumbService: BreadcrumbService) {
  }

  ngOnInit() {
    this.generateTrail(this.router.url);
    this.router.events.subscribe((event: Event) => {
      if (event instanceof NavigationEnd) {
        this.urls = [];
        this.generateTrail(event.url)
      }
    })
  }

  generateTrail(url: string) {
    if (url === '') {
      return;
    }
    if (!this.breadcrumbService.isRouteHidden(url)) {
      this.urls.unshift(url);
    }
    if (url.indexOf('/') > -1) {
      this.generateTrail(url.substr(0, url.lastIndexOf('/')));
    } else if (this.prefix.length > 0) {
      this.urls.unshift(this.prefix);
    }
  }

  navigateTo(url: string, avail: boolean) {
    if (avail) {
      this.router.navigateByUrl(url);
    }
  }

  friendlyName(url: string) {
    return this.breadcrumbService.getFriendName(url);
  }
}
