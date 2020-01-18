import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';

interface Summary {
  appTotal: number;
  userTotal: number;
  nodeTotal: number;
  podTotal: number;
}

@Component({
  selector: 'app-overview',
  template: `
    <div class="clr-row">
      <div class="clr-col-lg-12 clr-col-md-12 clr-col-sm-12 clr-col-xs-12">
        <div class="clr-row flex-items-xs-between flex-items-xs-top overview-offset">
          <h2 class="header-title">资源</h2>
        </div>
        <div class="clr-row group overview-offset">
          <app-card class="app-card">
            <div class="card-title">机器数量</div>
            <p class="card-text">
              <a href="javascript:void(0)" (click)="goToLink('/admin/kubernetes/node')" class="nav-link"> {{summary.nodeTotal}}</a>
            </p>
          </app-card>
          <app-card class="app-card">
            <div class="card-title">项目总数</div>
            <p class="card-text">
              <a href="javascript:void(0)" (click)="goToLink('/admin/reportform/app')" class="nav-link">{{summary.appTotal}}</a>
            </p>
          </app-card>
          <app-card class="app-card">
            <div class="card-title">用户总数</div>
            <p class="card-text">
              <a href="javascript:void(0)" (click)="goToLink('/admin/system/user')" class="nav-link"> {{summary.userTotal}}</a>
            </p>
          </app-card>
          <app-card class="app-card">
            <div class="card-title">实例数量</div>
            <p class="card-text">
              <a href="javascript:void(0)" (click)="goToLink('/admin/kubernetes/pod')" class="nav-link"> {{summary.podTotal}}</a>
            </p>
          </app-card>
        </div>
      </div>
    </div>
  `,
  styles: [
    `
      .group {
        margin-bottom: 15px;
      }

      .group, .clr-col-sm-2 {
        margin-bottom: 15px;
        font-weight: 800;
      }

      .card-text {
        font-size: 1.8rem;
        color: #377AEC;
        font-weight: 800;
        text-align: center;
      }

      .app-card {
        width: 185px;
        height: 116px;
      }

      .card-title {
        font-size: 16px;
        color: #222;
        letter-spacing: 1px;
        line-height: 24px;
        text-align: center;
      }

      .overview-offset {
        padding: 0px 15px;
      }
    `
  ],
})
export class OverviewComponent implements OnInit {
  summary: Summary = {
    appTotal: 0,
    userTotal: 0,
    nodeTotal: 0,
    podTotal: 0,
  };

  constructor(private router: Router) {}

  ngOnInit() {
  }

  goToLink(url: string) {
    this.router.navigateByUrl(url);
  }
}
