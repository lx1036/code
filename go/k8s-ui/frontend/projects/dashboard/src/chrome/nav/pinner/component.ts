

import {Component, Input} from '@angular/core';
import {PinnedResource} from '@api/backendapi';
import {PinnerService} from '../../../common/services/global/pinner';
import {Resource} from '../../../common/services/resource/endpoint';

@Component({
  selector: 'kd-pinner-nav',
  templateUrl: './template.html',
  styleUrls: ['../style.scss'],
})
export class PinnerNavComponent {
  @Input() kind: string;
  constructor(private readonly pinner_: PinnerService) {}

  isInitialized(): boolean {
    return this.pinner_.isInitialized();
  }

  getResourceHref(resource: PinnedResource): string {
    let href = `/${resource.kind}`;
    if (resource.namespace !== undefined) {
      href += `/${resource.namespace}`;
    }
    href += `/${resource.name}`;

    return href;
  }

  getPinnedResources(): PinnedResource[] {
    return this.pinner_.getPinnedForKind(this.kind);
  }

  unpin(resource: PinnedResource): void {
    this.pinner_.unpinResource(resource);
  }
}
