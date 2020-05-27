

import {HttpClient, HttpParams} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs/Observable';

@Injectable()
export class LogService {
  previous_ = false;
  inverted_ = true;
  compact_ = false;
  showTimestamp_ = false;
  following_ = true;
  autoRefresh_ = false;

  constructor(private readonly http_: HttpClient) {}

  getResource<T>(uri: string, params?: HttpParams): Observable<T> {
    return this.http_.get<T>(`api/v1/log/${uri}`, {params});
  }

  setFollowing(status: boolean): void {
    this.following_ = status;
  }

  toggleFollowing(): void {
    this.following_ = !this.following_;
  }

  getFollowing(): boolean {
    return this.following_;
  }

  setAutoRefresh(): void {
    this.autoRefresh_ = !this.autoRefresh_;
  }

  getAutoRefresh(): boolean {
    return this.autoRefresh_;
  }

  setPrevious(): void {
    this.previous_ = !this.previous_;
  }

  getPrevious(): boolean {
    return this.previous_;
  }

  setInverted(): void {
    this.inverted_ = !this.inverted_;
  }

  getInverted(): boolean {
    return this.inverted_;
  }

  setCompact(): void {
    this.compact_ = !this.compact_;
  }

  getCompact(): boolean {
    return this.compact_;
  }

  setShowTimestamp(): void {
    this.showTimestamp_ = !this.showTimestamp_;
  }

  getShowTimestamp(): boolean {
    return this.showTimestamp_;
  }

  getLogFileName(pod: string, container: string): string {
    return `logs-from-${container}-in-${pod}.txt`;
  }
}
