

import {Component, Input} from '@angular/core';

import {ResourceMeta} from '../../../../services/global/actionbar';
import {KdStateService} from '../../../../services/global/state';

@Component({
  selector: 'kd-actionbar-detail-logs',
  templateUrl: './template.html',
})
export class ActionbarDetailLogsComponent {
  @Input() resourceMeta: ResourceMeta;

  constructor(private readonly kdState_: KdStateService) {}

  getHref(): string {
    return this.kdState_.href(
      'log',
      this.resourceMeta.objectMeta.name,
      this.resourceMeta.objectMeta.namespace,
      this.resourceMeta.typeMeta.kind,
    );
  }
}
