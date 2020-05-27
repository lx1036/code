

import {EventEmitter, Injectable} from '@angular/core';
import {Event, NavigationEnd, NavigationStart, Router} from '@angular/router';

@Injectable()
export class KdStateService {
  onBefore = new EventEmitter();
  onSuccess = new EventEmitter();

  constructor(private readonly router_: Router) {
    this.router_.events.subscribe((event: Event) => {
      if (event instanceof NavigationStart) {
        this.onBefore.emit();
      }

      if (event instanceof NavigationEnd) {
        this.onSuccess.emit();
      }
    });
  }

  href(
    stateName: string,
    resourceName?: string,
    namespace?: string,
    resourceType?: string,
  ): string {
    resourceName = resourceName || '';
    namespace = namespace || '';
    resourceType = resourceType || '';

    if (namespace && resourceName === undefined) {
      throw new Error('Namespace can not be defined without resourceName.');
    }

    let href = `/${stateName}`;
    href = namespace ? `${href}/${namespace}` : href;
    href = resourceName ? `${href}/${resourceName}` : href;
    href = resourceType ? `${href}/${resourceType}` : href;

    return href;
  }
}
