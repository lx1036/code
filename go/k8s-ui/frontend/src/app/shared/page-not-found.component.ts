import {Component, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {defaultRoutingUrl} from './shared.const';

const defaultInterval = 1000;
const defaultLeftTime = 3;

@Component({
  selector: 'app-not-found',
  template: `
    <div>not found</div>
<!--      <div class="wrapper-back">-->
<!--          <div>-->
<!--              <clr-icon shape="warning" class="is-warning" size="96"></clr-icon>-->
<!--              <span class="status-code">404</span>-->
<!--              <span class="status-text">页面不存在</span>-->
<!--          </div>-->
<!--          <div class="status-subtitle">-->
<!--              正在重定向到首页： <span class="second-number">{{leftSeconds}}</span> 秒...-->
<!--          </div>-->
<!--      </div>-->
  `,
})
export class PageNotFoundComponent implements OnInit, OnDestroy {

  leftSeconds: number = defaultLeftTime;
  timeInterval: any = null;
  constructor(private router: Router) { }

  ngOnInit() {
    if (!this.timeInterval) {
      this.timeInterval = setInterval(interval => {
        this.leftSeconds--;
        if (this.leftSeconds <= 0) {
          this.router.navigate([defaultRoutingUrl]);
          clearInterval(this.timeInterval);
        }
      }, defaultInterval);
    }
  }

  ngOnDestroy(): void {
    if (this.timeInterval) {
      clearInterval(this.timeInterval);
    }
  }
}
