

import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {PluginMetadata, PluginsConfig} from '@api/frontendapi';
import {Observable} from 'rxjs/Observable';

@Injectable()
export class PluginsConfigService {
  private readonly pluginConfigPath_ = 'api/v1/plugin/config';
  private config_: PluginsConfig = {status: 204, plugins: [], errors: []};

  constructor(private readonly http: HttpClient) {}

  init(): void {
    this.fetchConfig();
  }

  refreshConfig(): void {
    this.fetchConfig();
  }

  private fetchConfig(): void {
    this.getConfig()
      .toPromise()
      .then(config => (this.config_ = config));
  }

  private getConfig(): Observable<PluginsConfig> {
    return this.http.get<PluginsConfig>(this.pluginConfigPath_);
  }

  pluginsMetadata(): PluginMetadata[] {
    return this.config_.plugins;
  }

  status(): number {
    return this.config_.status;
  }
}
