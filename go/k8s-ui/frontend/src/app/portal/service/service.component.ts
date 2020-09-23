import {Component, OnDestroy, OnInit} from '@angular/core';
import {AuthService} from "../../shared/auth.service";
import {syncStatusInterval} from "../../shared/shared.const";
import {ClrDatagridStateInterface} from "@clr/angular";

@Component({
  selector: 'app-service',
  templateUrl: './service.component.html',
  styleUrls: ['./service.component.scss']
})
export class ServiceComponent implements OnInit, OnDestroy {
  timer = null;
  serviceTpls: ServiceTp[];

  constructor(
    public authService: AuthService,
  ) {


    this.periodSyncStatus();
  }

  ngOnInit(): void {
  }

  ngOnDestroy() {
    clearInterval(this.timer)
  }

  periodSyncStatus() {
    this.timer = setInterval(() => {
      this.syncStatus()
    }, syncStatusInterval)
  }

  syncStatus() {
    if (this.serviceTpls && this.serviceTpls.length > 0) {

    }

  }

  onlineChange() {
    this.retrieve();
  }

  retrieve(state?: ClrDatagridStateInterface): void {

  }

}
