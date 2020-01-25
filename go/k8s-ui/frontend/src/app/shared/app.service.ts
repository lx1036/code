import {Injectable} from '@angular/core';
import {HttpClient, HttpParams} from '@angular/common/http';
import {Observable} from 'rxjs';

@Injectable()
export class AppService {

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
}
