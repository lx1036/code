

import 'rxjs/add/operator/catch';

import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {CanIResponse} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';

import {ERRORS} from '../../errors/errors';

@Injectable()
export class AuthorizerService {
  authorizationSubUrl_ = '/cani';

  constructor(private readonly http_: HttpClient) {}

  proxyGET<T>(url: string): Observable<T> {
    return this.http_
      .get<CanIResponse>(`${url}${this.authorizationSubUrl_}`)
      .switchMap<CanIResponse, T>(response => {
        if (!response.allowed) {
          return Observable.throwError(ERRORS.forbidden);
        }

        return this.http_.get<T>(url);
      })
      .catch(e => {
        return Observable.throwError(e);
      });
  }
}
