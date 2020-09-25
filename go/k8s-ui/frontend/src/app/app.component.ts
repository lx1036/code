import {AfterViewInit, Component} from '@angular/core';
import {ScrollBarService} from './shared/scroll-bar.service';
import {TranslateService} from '@ngx-translate/core';
import {StorageService} from "./shared/common/storage.service";

@Component({
  selector: 'app-root',
  template: `
    <router-outlet></router-outlet>
  `,
})
export class AppComponent implements AfterViewInit {
  constructor(private scrollBar: ScrollBarService, public translate: TranslateService, private storage: StorageService) {
    this.translate.addLangs(['en', 'zh-Hans']);
    this.translate.setDefaultLang('en');
    const lang = this.storage.get('lang');
    if (lang) {
      this.translate.use(lang);
    } else {
      this.translate.use('en');
    }
  }

  ngAfterViewInit(): void {
    // this.scrollBar.init(); // calculate scroll-bar width
  }
}
