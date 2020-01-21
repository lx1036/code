import {Component, OnInit} from '@angular/core';

@Component({
  selector: 'app-admin',
  template: `
    <clr-main-container class="main-container">
<!--      <global-message></global-message>-->
      <app-admin-nav></app-admin-nav>
      <div class="content-container">
        <div class="content-area">
<!--          <wayne-breadcrumb></wayne-breadcrumb>-->
          <router-outlet></router-outlet>
        </div>
<!--        <wayne-sidenav style="display: flex; order: -1"></wayne-sidenav>-->
      </div>
    </clr-main-container>

<!--    <confiramtion-dialog style="display: flex"></confiramtion-dialog>-->
<!--    <wayne-ace-editor></wayne-ace-editor>-->
<!--    <tpl-detail></tpl-detail>-->
  `
})
export class AdminComponent implements OnInit {
  constructor() {
  }

  ngOnInit() {
  }
}
