

import {Component, Input} from '@angular/core';
import {ResourceOwner} from '@api/backendapi';
import {KdStateService} from '../../services/global/state';

@Component({
  selector: 'kd-creator-card',
  templateUrl: './template.html',
})
export class CreatorCardComponent {
  @Input() creator: ResourceOwner;
  @Input() initialized: boolean;

  constructor(private readonly kdState_: KdStateService) {}

  getHref(): string {
    return this.kdState_.href(
      this.creator.typeMeta.kind,
      this.creator.objectMeta.name,
      this.creator.objectMeta.namespace,
    );
  }
}
