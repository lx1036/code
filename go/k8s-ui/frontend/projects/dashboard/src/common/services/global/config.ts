

import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {AppConfig} from '@api/backendapi';
import {VersionInfo} from '@api/frontendapi';
import {Observable} from 'rxjs/Observable';
import {version} from '../../../environments/version';

@Injectable()
export class ConfigService {
  private readonly configPath_ = 'config';
  private config_: AppConfig;
  private initTime_: number;

  constructor(private readonly http: HttpClient) {}

  init(): void {
    this.getAppConfig().subscribe(config => {
      // Set init time when response from the backend will arrive.
      this.config_ = config;
      this.initTime_ = new Date().getTime();
    });
  }

  getAppConfig(): Observable<AppConfig> {
    return this.http.get<AppConfig>(this.configPath_);
  }

  getServerTime(): Date {
    if (this.config_.serverTime) {
      const elapsed = new Date().getTime() - this.initTime_;
      return new Date(this.config_.serverTime + elapsed);
    } else {
      return new Date();
    }
  }

  getVersionInfo(): VersionInfo {
    return version;
  }
}
