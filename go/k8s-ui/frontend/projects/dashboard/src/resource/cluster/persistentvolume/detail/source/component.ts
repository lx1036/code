

import {Component, Input} from '@angular/core';
import {PersistentVolumeSource} from '@api/backendapi';

@Component({
  selector: 'kd-persistent-volume-source',
  templateUrl: './template.html',
})
export class PersistentVolumeSourceComponent {
  @Input() source: PersistentVolumeSource;
  @Input() initialized: boolean;
}
