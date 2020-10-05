import {Component, Input, OnInit} from '@angular/core';
import {ClrDatagridStateInterface} from '@clr/angular';
import {App} from "../../../shared/models/app";
import {AuthService} from "../../../shared/components/auth/auth.service";

@Component({
  selector: 'app-list-apps',
  templateUrl: './list-apps.component.html',
})

export class ListAppsComponent implements OnInit {
  showState: any;
  @Input() apps: App[];

  constructor(public authService: AuthService) {
  }

  ngOnInit() {
  }

  unstarredApp(app: App) {

  }

  deleteApp(app: App) {

  }

  refresh($event: ClrDatagridStateInterface) {

  }

  enterApp(app: App) {

  }

  goToMonitor(app: any) {

  }

  getMonitorUri() {

  }

  editApp(app: App) {

  }

  starredApp(app: App) {

  }
}
