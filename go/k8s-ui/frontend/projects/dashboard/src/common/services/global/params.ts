

import {Injectable} from '@angular/core';
import {ActivatedRoute, NavigationEnd, Params, Router} from '@angular/router';
import {Subject} from 'rxjs';
import {filter} from 'rxjs/operators';

@Injectable()
export class ParamsService {
  onParamChange = new Subject<void>();

  private params_: Params = {};
  private queryParamMap_: Params = {};

  constructor(private router_: Router, private route_: ActivatedRoute) {
    this.router_.events.pipe(filter(event => event instanceof NavigationEnd)).subscribe(() => {
      let active = this.route_;
      while (active.firstChild) {
        active = active.firstChild;
      }

      active.params.subscribe((params: Params) => {
        this.copyParams_(params, this.params_);
        this.onParamChange.next();
      });

      active.params.subscribe((params: Params) => {
        this.copyParams_(params, this.queryParamMap_);
        this.onParamChange.next();
      });
    });
  }

  getRouteParam(name: string) {
    return !!this.params_ ? this.params_[name] : undefined;
  }

  getQueryParam(name: string) {
    return !!this.queryParamMap_ ? this.queryParamMap_[name] : undefined;
  }

  setQueryParam(name: string, value: string) {
    if (!!this.queryParamMap_) this.queryParamMap_[name] = value;
  }

  private copyParams_(from: Params, to: Params) {
    for (const key of Object.keys(from)) {
      to[key] = from[key];
    }
  }
}
