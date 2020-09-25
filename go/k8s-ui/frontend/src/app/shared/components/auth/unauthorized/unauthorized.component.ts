import {Component, Inject, OnDestroy, OnInit} from '@angular/core';
import {DOCUMENT} from '@angular/common';

const defaultInterval = 1000;
const defaultLeftTime = 1;

@Component({
  selector: 'app-unauthorized',
  template: `
    <div class="wrapper-back">
      <div>
        <clr-icon shape="warning" class="is-warning" size="96"></clr-icon>
        <span class="status-code">401</span>
        <span class="status-text">未登录或登录状态失效</span>
      </div>
      <div class="status-subtitle">
        正在重定向到登录页： <span class="second-number">{{leftSeconds}}</span> 秒...
      </div>
    </div>
  `,
})
export class UnauthorizedComponent implements OnInit, OnDestroy {

  leftSeconds: number = defaultLeftTime;
  timeInterval: any = null;
  constructor(@Inject(DOCUMENT) private document: any) { }

  ngOnInit() {
    if (!this.timeInterval) {
      this.timeInterval = setInterval(interval => {
        this.leftSeconds--;
        if (this.leftSeconds <= 0) {
          // 未授权重定向到登录页面
          // document.location.href
          const currentUrl = this.document.location.origin;
          setTimeout(() => {
            this.document.location.href = `${currentUrl}/sign-in`;
          }, defaultLeftTime);
          clearInterval(this.timeInterval);
        }
      }, defaultInterval);
    }
  }

  ngOnDestroy(): void {
  }
}
