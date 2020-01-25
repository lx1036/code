import {Component, OnInit} from '@angular/core';

@Component({
  selector: 'app-namespace-report',
  template: `
    <div class="content-area" style="position: relative">
      <app-box style="padding: 15px">
        <div class="clr-row">
          <div class="clr-col-lg-12 clr-col-md-12 clr-col-sm-12 clr-col-xs-12">
            <div class="clr-row flex-items-xs-between flex-items-xs-top" style="padding-left: 15px; padding-right: 15px;">
              <h2 class="header-title">{{'MENU.DEPARTMENT' | translate}}</h2>
            </div>
          </div>
        </div>

        <clr-tabs>
          <clr-tab>
            <a clrTabLink>{{'MENU.OVERVIEW' | translate}}</a>
            <clr-tab-content *clrIfActive="true">
              <app-overview></app-overview>
            </clr-tab-content>
          </clr-tab>
          <clr-tab>
            <a clrTabLink>{{'MENU.SOURCE_STAT' | translate}}</a>
            <clr-tab-content *clrIfActive>
              <app-report-resource></app-report-resource>
            </clr-tab-content>
          </clr-tab>
          <clr-tab>
            <a clrTabLink>{{'MENU.RELEASE_RECORD' | translate}}</a>
            <clr-tab-content *clrIfActive>
              <app-report-history></app-report-history>
            </clr-tab-content>
          </clr-tab>
        </clr-tabs>
      </app-box>
    </div>

    <app-sidenav-namespace></app-sidenav-namespace>
  `
})

export class NamespaceReportComponent implements OnInit {
  constructor() {
  }

  ngOnInit() {
  }
}
