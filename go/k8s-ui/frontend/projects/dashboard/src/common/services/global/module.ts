

import {HTTP_INTERCEPTORS} from '@angular/common/http';
import {APP_INITIALIZER, Injector, NgModule} from '@angular/core';

import {ActionbarService} from './actionbar';
import {AssetsService} from './assets';
import {AuthService} from './authentication';
import {AuthorizerService} from './authorizer';
import {ConfigService} from './config';
import {CsrfTokenService} from './csrftoken';
import {GlobalSettingsService} from './globalsettings';
import {HistoryService} from './history';
import {AuthInterceptor} from './interceptor';
import {LocalSettingsService} from './localsettings';
import {LogService} from './logs';
import {NamespaceService} from './namespace';
import {NotificationsService} from './notifications';
import {ParamsService} from './params';
import {KdStateService} from './state';
import {ThemeService} from './theme';
import {TitleService} from './title';
import {VerberService} from './verber';
import {PluginsConfigService} from './plugin';
import {PluginLoaderService} from '../pluginloader/pluginloader.service';
import {ClientPluginLoaderService} from '../pluginloader/clientloader.service';
import {PinnerService} from './pinner';

@NgModule({
  providers: [
    AuthorizerService,
    AssetsService,
    LocalSettingsService,
    GlobalSettingsService,
    ConfigService,
    PluginsConfigService,
    TitleService,
    AuthService,
    CsrfTokenService,
    NotificationsService,
    ThemeService,
    KdStateService,
    NamespaceService,
    ActionbarService,
    VerberService,
    PinnerService,
    HistoryService,
    LogService,
    ParamsService,
    {
      provide: APP_INITIALIZER,
      useFactory: init,
      deps: [
        GlobalSettingsService,
        LocalSettingsService,
        ConfigService,
        HistoryService,
        PluginsConfigService,
        PinnerService,
      ],
      multi: true,
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthInterceptor,
      multi: true,
    },
    {provide: PluginLoaderService, useClass: ClientPluginLoaderService},
  ],
})
export class GlobalServicesModule {
  static injector: Injector;
  constructor(injector: Injector) {
    GlobalServicesModule.injector = injector;
  }
}

export function init(
  globalSettings: GlobalSettingsService,
  localSettings: LocalSettingsService,
  pinner: PinnerService,
  config: ConfigService,
  history: HistoryService,
  pluginsConfig: PluginsConfigService,
): Function {
  return () => {
    globalSettings.init();
    localSettings.init();
    pluginsConfig.init();
    pinner.init();
    config.init();
    history.init();
  };
}
