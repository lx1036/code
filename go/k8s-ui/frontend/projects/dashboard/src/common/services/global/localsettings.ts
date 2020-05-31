

import {Injectable} from '@angular/core';
import {LocalSettings} from '@api/backendapi';
import {CookieService} from 'ngx-cookie-service';

import {ThemeService} from './theme';

@Injectable()
export class LocalSettingsService {
  private readonly cookieName_ = 'localSettings';
  private settings_: LocalSettings = {
    isThemeDark: false,
  };

  constructor(private readonly theme_: ThemeService, private readonly cookies_: CookieService) {}

  init(): void {
    const cookieValue = this.cookies_.get(this.cookieName_);
    if (cookieValue && cookieValue.length > 0) {
      this.settings_ = JSON.parse(cookieValue);
    }
  }

  get(): LocalSettings {
    return this.settings_;
  }

  handleThemeChange(isThemeDark: boolean): void {
    this.settings_.isThemeDark = isThemeDark;
    this.updateCookie_();
    this.theme_.switchTheme(!this.settings_.isThemeDark);
  }

  updateCookie_(): void {
    this.cookies_.set(
      this.cookieName_,
      JSON.stringify(this.settings_),
      null,
      null,
      null,
      false,
      'Strict',
    );
  }
}
