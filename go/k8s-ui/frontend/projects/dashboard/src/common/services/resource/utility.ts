

import {HttpClient, HttpParams} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';

import {ResourceBase} from '../../resources/resource';
import {NamespaceService} from '../global/namespace';

@Injectable()
export class UtilityService<T> extends ResourceBase<T> {
  constructor(readonly http: HttpClient, private readonly namespace_: NamespaceService) {
    super(http);
  }

  shell(endpoint: string, params?: HttpParams): Observable<T> {
    return this.http.get<T>(endpoint, {params});
  }
}
