

import {Component, Input} from '@angular/core';

export enum HiddenPropertyMode {
  Hidden = 'hidden',
  Visible = 'visible',
  Edit = 'edit',
}

@Component({
  selector: 'kd-hidden-property',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class HiddenPropertyComponent {
  @Input() mode = HiddenPropertyMode.Hidden;
  @Input() enableEdit = false;

  HiddenPropertyMode = HiddenPropertyMode;

  toggleVisible(): void {
    this.mode =
      this.mode !== HiddenPropertyMode.Visible
        ? HiddenPropertyMode.Visible
        : HiddenPropertyMode.Hidden;
  }

  toggleEdit(): void {
    this.mode =
      this.mode !== HiddenPropertyMode.Edit ? HiddenPropertyMode.Edit : HiddenPropertyMode.Hidden;
  }
}
