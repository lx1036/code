import {Component, OnInit} from '@angular/core';

@Component({
  selector: 'app-portal',
  template: `
<!--      <clr-main-container class="main-container">-->
<!--          <global-message></global-message>-->
<!--          <diff></diff>-->
<!--          <wayne-nav></wayne-nav>-->
<!--          <router-outlet></router-outlet>-->
<!--      </clr-main-container>-->
<!--      <confiramtion-dialog style="display: flex"></confiramtion-dialog>-->
    <router-outlet></router-outlet>
  `
})

export class PortalComponent implements OnInit {
  constructor() {
  }

  ngOnInit() {
  }
}
