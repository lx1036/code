import {Injectable} from '@angular/core';
import {PageState} from './page-state';
import {HttpClient, HttpParams} from '@angular/common/http';
import {Observable} from 'rxjs';
import {Notification} from './model/v1/notification';


export type NotificationType = string;
export type NotificationLevel = number;

export interface NotificationMessage {
  id?: number;
  type?: NotificationType;
  title?: string;
  message?: string;
  level?: NotificationLevel;
  is_published: boolean;
}
export interface Page {
  pageNo: number;
  pageSize: number;
  totalPage: number;
  totalCount: number;
  list: NotificationMessage[];
}

export interface NotificationLog {
  id: number;
  user_id: number;
  is_read: boolean;
  notification: NotificationMessage[];
}

@Injectable()
export class NotificationService {

  constructor(private http: HttpClient) {}

  query(pageState?: PageState): Observable<any> {
    const params = new HttpParams();
    // params.set('pageNo', pageState.page.pageNo + '');
    // params.set('pageSize', pageState.page.pageSize + '');
    // params.set('sortBy', '-id');
    return this.http.get(`/api/v1/notifications`, {params});
  }

  subscribe(pageState: PageState): Observable<any> {
    const params = new HttpParams();
    // params.set('pageNo', pageState.page.pageNo + '');
    // params.set('pageSize', pageState.page.pageSize + '');
    // params.set('sortBy', '-id');
    return this.http.get(`/api/v1/notifications/subscribe`, {params});
  }

  create(notify: Notification) {
    return this.http.post(`/api/v1/notifications`, notify);
  }
}
