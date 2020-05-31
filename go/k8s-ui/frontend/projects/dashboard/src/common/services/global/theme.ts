

import {EventEmitter, Injectable} from '@angular/core';
import {ThemeSwitchCallback} from '@api/frontendapi';

@Injectable()
export class ThemeService {
  private isLightThemeEnabled_ = true;
  private readonly onThemeSwitchEvent_ = new EventEmitter<boolean>();

  isLightThemeEnabled(): boolean {
    return this.isLightThemeEnabled_;
  }

  switchTheme(isLightTheme: boolean): void {
    this.onThemeSwitchEvent_.emit(isLightTheme);
    this.isLightThemeEnabled_ = isLightTheme;
  }

  subscribe(callback: ThemeSwitchCallback): void {
    this.onThemeSwitchEvent_.subscribe(callback);
  }
}
