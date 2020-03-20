import {Injectable} from "@angular/core";
import {Observable} from "rxjs";
import {PluginsConfig} from "../../../typings/frontend-api";
import {HttpClient} from "@angular/common/http";


@Injectable()
export class PluginConfigService {
  private readonly pluginConfigPath_ = 'api/v1/plugin/config';
  private config_: PluginsConfig = {status: 204, plugins: [], errors: []};

  constructor(private readonly http: HttpClient) {}

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
}
