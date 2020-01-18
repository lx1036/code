import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs';

@Injectable()
export class AppService {

  constructor(private http: HttpClient) {}

  getStatistics(): Observable<any> {
    return this.http.get(`/api/v1/apps/statistics`);
  }
}
