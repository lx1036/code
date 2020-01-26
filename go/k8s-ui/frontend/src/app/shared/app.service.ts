import {Injectable} from '@angular/core';
import {HttpClient, HttpHeaders, HttpParams} from '@angular/common/http';
import {Observable} from 'rxjs';
import {App} from './model/v1/app';

@Injectable()
export class AppService {
  headers = new HttpHeaders({'Content-type': 'application/json'});
  options = {headers: this.headers};
  constructor(private http: HttpClient) {}

  getStatistics(): Observable<any> {
    return this.http.get(`/api/v1/apps/statistics`);
  }

  listResourceCount(namespaceId: number, appId?: number) {
    const params = new HttpParams();
    if (appId != null) {
      params.set('appId', appId + '');
    }

    return this.http.get(`/api/v1/namespaces/${namespaceId}/statistics`, {params});
  }

  create(app: App): Observable<any> {
    return this.http.post(`/api/v1/namespaces/${app.namespace.id}/apps`, app, this.options);
  }

  update(app: App): Observable<any> {
    return this.http.put(`/api/v1/namespaces/${app.namespace.id}/apps/${app.id}`, app, this.options);
  }
}
