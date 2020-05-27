

import {Component, Input} from '@angular/core';
import {ObjectMeta} from '@api/backendapi';
import {KdStateService} from '../../../../services/global/state';

@Component({
  selector: 'kd-actionbar-detail-exec',
  templateUrl: './template.html',
})
export class ActionbarDetailExecComponent {
  @Input() objectMeta: ObjectMeta;

  constructor(private readonly kdState_: KdStateService) {}

  getHref(): string {
    return this.kdState_.href('shell', this.objectMeta.name, this.objectMeta.namespace);
  }
}
