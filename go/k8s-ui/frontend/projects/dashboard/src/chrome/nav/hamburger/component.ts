

import {Component} from '@angular/core';
import {NavService} from '../../../common/services/nav/service';

@Component({
  selector: 'kd-nav-hamburger',
  templateUrl: 'template.html',
})
export class HamburgerComponent {
  constructor(private readonly navService_: NavService) {}

  toggle(): void {
    this.navService_.toggle();
  }
}
