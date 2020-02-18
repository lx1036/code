import {Injectable} from '@angular/core';

@Injectable()
export class BreadcrumbService {
  private routesFriendlyNames: Map<string, object> = new Map<string, object>();
  private routesFriendlyNamesRegex: Map<string, object> = new Map<string, object>();

  constructor() {
  }

  getFriendName(url: string) {
    let urlInfo: any = new Object();
    const routeEnd = url.substr(url.lastIndexOf('/') + 1, url.length);
    const info: any = this.routesFriendlyNames.get(url);
    if (info !== undefined) {
      urlInfo = Object.assign({}, info);
    }
    this.routesFriendlyNamesRegex.forEach((value, key, map) => {
      if (new RegExp(key).exec(url)) {
        urlInfo = Object.assign({}, value);
      }
    });
    return Object.keys(urlInfo).length ? urlInfo : {name: routeEnd, avail: false};
  }

  isRouteHidden(url: string) {
    return false;
  }
}
