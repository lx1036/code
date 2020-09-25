import {HttpClient} from "@angular/common/http";
import {Injectable} from "@angular/core";
import {KubeResourcesName} from "../../shared.const";
import {Observable, throwError} from "rxjs";
import {catchError} from "rxjs/operators";


@Injectable()
export class KubernetesClient {
  constructor(private http: HttpClient) {
  }

  get(cluster: string, kind: KubeResourcesName, name: string, namespace?: string, appId?: string): Observable<any> {
    if ((typeof (appId) === 'undefined') || (!appId)) {
      appId = '0';
    }

    let link = `/api/v1/apps/${appId}/_proxy/clusters/${cluster}/${kind}/${name}`;
    if (namespace) {
      link = `/api/v1/apps/${appId}/_proxy/clusters/${cluster}/namespaces/${namespace}/${kind}/${name}`;
    }

    return this.http
      .get(link).pipe(
        catchError(err => throwError(err))
      )
  }
}
