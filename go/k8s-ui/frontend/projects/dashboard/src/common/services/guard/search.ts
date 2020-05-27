

import {Injectable} from '@angular/core';
import {
  ActivatedRouteSnapshot,
  CanDeactivate,
  Params,
  Router,
  RouterStateSnapshot,
  UrlTree,
} from '@angular/router';
import {SearchComponent} from '../../../search/component';
import {SEARCH_QUERY_STATE_PARAM} from '../../params/params';

@Injectable()
export class SearchGuard implements CanDeactivate<SearchComponent> {
  private readonly queryParamSeparator_ = '&';
  private readonly queryParamStart_ = '?';

  constructor(private readonly router_: Router) {}

  canDeactivate(
    _cmp: SearchComponent,
    _route: ActivatedRouteSnapshot,
    _routeSnapshot: RouterStateSnapshot,
    nextState?: RouterStateSnapshot,
  ): boolean | UrlTree {
    let url = nextState.url;
    const queryParams = this.getQueryParams_(url);

    if (queryParams[SEARCH_QUERY_STATE_PARAM]) {
      url = this.removeQueryParamFromUrl_(url);
      return this.router_.parseUrl(url);
    }

    return true;
  }

  private getQueryParams_(url: string): Params {
    const paramMap: {[key: string]: string} = {};
    const queryStartIdx = url.indexOf(this.queryParamStart_) + 1;
    const partials = url.substring(queryStartIdx).split(this.queryParamSeparator_);

    for (const partial of partials) {
      const params = partial.split('=');
      if (params.length === 2) {
        paramMap[params[0]] = params[1];
      }
    }

    return paramMap;
  }

  private removeQueryParamFromUrl_(url: string): string {
    const queryStartIdx = url.indexOf(this.queryParamStart_) + 1;
    const rawUrl = url.substring(0, queryStartIdx - 1);

    const paramMap = this.getQueryParams_(url);
    if (paramMap[SEARCH_QUERY_STATE_PARAM]) {
      delete paramMap[SEARCH_QUERY_STATE_PARAM];
    }

    const queryParams = Object.keys(paramMap)
      .map(key => `${key}=${paramMap[key]}`)
      .join(this.queryParamSeparator_);
    return `${rawUrl}${this.queryParamStart_}${queryParams}`;
  }
}
