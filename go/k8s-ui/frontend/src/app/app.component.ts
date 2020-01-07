import {AfterViewInit, Component} from '@angular/core';
import {ScrollBarService} from "./shared/client/v1/scroll-bar.service";

@Component({
  selector: 'app-root',
  template: `
    <router-outlet></router-outlet>
  `,
})
export class AppComponent implements AfterViewInit {
  constructor(private scrollBar: ScrollBarService, ) {

  }

  ngAfterViewInit(): void {
    // this.scrollBar.init(); // calculate scroll-bar width
  }
}
