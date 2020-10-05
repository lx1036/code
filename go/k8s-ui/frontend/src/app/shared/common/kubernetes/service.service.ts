import {Injectable} from "@angular/core";
import {HttpClient, HttpParams} from "@angular/common/http";
import {Observable, throwError} from "rxjs";
import {catchError} from "rxjs/operators";


@Injectable()
export class ServiceService {
  constructor(private http: HttpClient) {}

  deleteById(id: number, appId: number, logical?: boolean): Observable<any> {
    const options: any = {};
    if (logical != null) {
      let params = new HttpParams();
      params = params.set('logical', logical + '');
      options.params = params;
    }

    return this.http.delete(`/api/v1/apps/${appId}/services/${id}`, options).pipe(
      catchError(err => throwError(err))
    )
  }

}

