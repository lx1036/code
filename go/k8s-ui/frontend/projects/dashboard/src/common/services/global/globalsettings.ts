

import {HttpClient, HttpHeaders} from '@angular/common/http';
import {EventEmitter, Injectable} from '@angular/core';
import {GlobalSettings} from '@api/backendapi';
import {onSettingsFailCallback, onSettingsLoadCallback} from '@api/frontendapi';
import {of, ReplaySubject, Subject} from 'rxjs';
import {Observable} from 'rxjs/Observable';
import {catchError, switchMap, takeUntil} from 'rxjs/operators';

import {AuthorizerService} from './authorizer';

@Injectable()
export class GlobalSettingsService {
  onSettingsUpdate = new ReplaySubject<void>();
  onPageVisibilityChange = new EventEmitter<boolean>();

  private readonly endpoint_ = 'api/v1/settings/global';
  private settings_: GlobalSettings = {
    itemsPerPage: 10,
    clusterName: '',
    logsAutoRefreshTimeInterval: 5,
    resourceAutoRefreshTimeInterval: 5,
    disableAccessDeniedNotifications: false,
  };
  private unsubscribe_ = new Subject<void>();
  private isInitialized_ = false;
  private isPageVisible_ = true;

  constructor(
    private readonly http_: HttpClient,
    private readonly authorizer_: AuthorizerService,
  ) {}

  init(): void {
    this.load();

    this.onPageVisibilityChange.pipe(takeUntil(this.unsubscribe_)).subscribe(visible => {
      this.isPageVisible_ = visible;
      this.onSettingsUpdate.next();
    });
  }

  isInitialized(): boolean {
    return this.isInitialized_;
  }

  load(onLoad?: onSettingsLoadCallback, onFail?: onSettingsFailCallback): void {
    this.http_
      .get<GlobalSettings>(this.endpoint_)
      .toPromise()
      .then(
        settings => {
          this.settings_ = settings;
          this.isInitialized_ = true;
          this.onSettingsUpdate.next();
          if (onLoad) onLoad(settings);
        },
        err => {
          this.isInitialized_ = false;
          this.onSettingsUpdate.next();
          if (onFail) onFail(err);
        },
      );
  }

  canI(): Observable<boolean> {
    return this.authorizer_
      .proxyGET<GlobalSettings>(this.endpoint_)
      .pipe(switchMap(_ => of(true)))
      .pipe(catchError(_ => of(false)));
  }

  save(settings: GlobalSettings): Observable<GlobalSettings> {
    const httpOptions = {
      method: 'PUT',
      headers: new HttpHeaders({
        'Content-Type': 'application/json',
      }),
    };
    return this.http_.put<GlobalSettings>(this.endpoint_, settings, httpOptions);
  }

  getClusterName(): string {
    return this.settings_.clusterName;
  }

  getItemsPerPage(): number {
    return this.settings_.itemsPerPage;
  }

  getLogsAutoRefreshTimeInterval(): number {
    return this.isPageVisible_ ? this.settings_.logsAutoRefreshTimeInterval : 0;
  }

  getResourceAutoRefreshTimeInterval(): number {
    return this.isPageVisible_ ? this.settings_.resourceAutoRefreshTimeInterval : 0;
  }

  getDisableAccessDeniedNotifications(): boolean {
    return this.settings_.disableAccessDeniedNotifications;
  }
}
