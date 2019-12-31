import {Injectable} from '@angular/core';
import {Observable} from "rxjs";
import {HttpClient, HttpParams} from "@angular/common/http";
import {catchError} from "rxjs/operators";

@Injectable()
export class NamespaceClient {
  
  constructor(private http: HttpClient) {
  }
  
  getResourceUsage(namespaceId: number, appName?: string): Observable<any> {
    let params = new HttpParams();
    if (appName) {
      params = params.set('app', appName);
    }
    
    return this.http.get(`/api/v1/kubernetes/namespaces/${namespaceId}/resources`, {params: params});
  }
}
