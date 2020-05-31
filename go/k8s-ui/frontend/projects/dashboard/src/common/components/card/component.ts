

import {Component, Input} from '@angular/core';
import {Animations} from '../../animations/animations';

@Component({
  selector: 'kd-card',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
  animations: [Animations.expandInOut],
})
export class CardComponent {
  @Input() initialized = true;
  @Input() role: string;
  @Input() withFooter = false;
  @Input() withTitle = true;
  @Input() expandable = true;
  @Input()
  set titleClasses(val: string) {
    this.classes_ = val.split(/\s+/);
  }
  @Input() expanded = true;
  @Input() graphMode = false;

  private classes_: string[] = [];

  expand(): void {
    if (this.expandable) {
      this.expanded = !this.expanded;
    }
  }

  getTitleClasses(): {[clsName: string]: boolean} {
    const ngCls = {} as {[clsName: string]: boolean};
    if (!this.expanded) {
      ngCls['kd-minimized-card-header'] = true;
    }

    if (this.expandable) {
      ngCls['kd-card-header'] = true;
    }

    for (const cls of this.classes_) {
      ngCls[cls] = true;
    }

    return ngCls;
  }
}
