

import {Component, Input} from '@angular/core';

@Component({selector: 'kd-settings-entry', templateUrl: './template.html'})
export class SettingsEntryComponent {
  @Input() key: string;
  @Input() desc: string;
}
