import {Component, OnDestroy, OnInit} from "@angular/core";


@Component({
  selector: "base",
  template: `
    <div class="content-area" style="position: relative;padding: .75rem .75rem .75rem .75rem;">
      <app-box [disabled]="!showBox" [ngStyle]="{padding: showBox ? '15px' : 0}">
        <router-outlet></router-outlet>
      </app-box>
    </div>
    <app-sidenav style="display: flex; order: -1"></app-sidenav>
<!--    <publish-history></publish-history>-->
<!--    <tpl-detail></tpl-detail>-->
  `
})
export class BaseComponent implements OnInit, OnDestroy{

}
