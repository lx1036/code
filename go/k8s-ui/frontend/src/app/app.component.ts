import {AfterViewInit, Component} from '@angular/core';
import {ScrollBarService} from './shared/scroll-bar.service';
import {TranslateService} from '@ngx-translate/core';
import {StorageService} from './shared/storage.service';

@Component({
  selector: 'app-root',
  template: `
    <router-outlet></router-outlet>
  `,
})
export class AppComponent implements AfterViewInit {
  constructor(private scrollBar: ScrollBarService, public translate: TranslateService, private storage: StorageService) {
    translate.addLangs(['en', 'zh-Hans']);
    translate.setDefaultLang('en');
    const lang = storage.get('lang');
    if (lang) {
      translate.use(lang);
    } else {
      translate.use('en');
    }
  }

  ngAfterViewInit(): void {
    // this.scrollBar.init(); // calculate scroll-bar width
  }
}
