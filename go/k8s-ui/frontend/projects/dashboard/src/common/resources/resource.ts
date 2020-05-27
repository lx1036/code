

import {HttpClient, HttpParams} from '@angular/common/http';
import {merge, timer} from 'rxjs';
import {publishReplay, refCount, switchMapTo} from 'rxjs/operators';

// @ts-ignore
export abstract class ResourceBase<T> {
  protected constructor(protected readonly http_: HttpClient) {}
}
