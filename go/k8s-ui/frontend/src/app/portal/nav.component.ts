import {Component, OnInit} from '@angular/core';
import {CacheService} from '../shared/cache.service';
import {AuthService} from '../shared/auth.service';

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
    </header>
  `
})

export class NavComponent implements OnInit {
  constructor(public cacheService: CacheService, public authService: AuthService) {
  }

  ngOnInit() {
  }

  getTitle() {
    const imagePrefix = this.authService.config['system.title'];
    return imagePrefix ? imagePrefix : 'Kubernetes-UI';
  }
}
