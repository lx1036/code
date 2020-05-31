

import {Component, Input} from '@angular/core';

@Component({
  selector: 'kd-nav-item',
  templateUrl: 'template.html',
  styleUrls: ['style.scss'],
})
export class NavItemComponent {
  @Input() state: string;
  @Input() exact = false;
}
