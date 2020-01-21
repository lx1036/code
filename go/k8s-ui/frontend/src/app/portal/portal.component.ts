import {Component, OnInit} from '@angular/core';

@Component({
  selector: 'app-portal',
  template: `
    <clr-main-container class="main-container">
      <app-global-message></app-global-message>
      <app-diff></app-diff>
      <app-nav></app-nav>
      <router-outlet></router-outlet>
    </clr-main-container>
    <app-confirmation-dialog style="display: flex"></app-confirmation-dialog>
  `
})

export class PortalComponent implements OnInit {
  constructor() {
  }

  ngOnInit() {
  }
}
