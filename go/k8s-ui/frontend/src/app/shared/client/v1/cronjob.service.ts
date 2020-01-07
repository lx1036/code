import {Injectable} from '@angular/core';
import {Observable, throwError} from 'rxjs';
import {HttpClient, HttpParams} from '@angular/common/http';
import {catchError} from "rxjs/operators";

@Injectable()
export class CronjobService {

  constructor(private http: HttpClient) {
  }


  deleteById(id: number, appId: number, logical?: boolean): Observable<any> {
    const options: any = {};
    if (logical != null) {
      let params = new HttpParams();
      params = params.set('logical', logical + '');
      options.params = params;
    }

    return this.http.delete(`/api/v1/apps/${appId}/cronjobs/${id}`, options).pipe(
        catchError(error => throwError(error)),
      );
  }

  getById(id: number, appId: number): Observable<any> {
    return this.http.get(`/api/v1/apps/${appId}/cronjobs/${id}`).pipe(
        catchError(error => throwError(error)),
      );
  }
}
