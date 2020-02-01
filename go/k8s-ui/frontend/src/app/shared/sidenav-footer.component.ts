import {Component} from '@angular/core';

@Component({
  selector: 'app-sidenav-footer',
  template: `
      <a class="nav-link ng-star-inserted footer" href="javascript:void(0)">
        <span class="nav-text"> © {{year}} K8S-UI · {{version}}</span>
      </a>
  `
})
export class SideNavFooterComponent {
  version = require('../../../package.json').version;
  year = new Date().getFullYear().toString();
}
