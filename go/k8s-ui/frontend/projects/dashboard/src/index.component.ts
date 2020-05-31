

import {OverlayContainer} from '@angular/cdk/overlay';
import {Component, ElementRef, OnInit} from '@angular/core';

import {LocalSettingsService} from './common/services/global/localsettings';
import {ThemeService} from './common/services/global/theme';
import {TitleService} from './common/services/global/title';

enum Themes {
  Light = 'kd-light-theme',
  Dark = 'kd-dark-theme',
}

@Component({selector: 'kd-root', template: '<router-outlet></router-outlet>'})
export class RootComponent implements OnInit {
  private isLightThemeEnabled_: boolean;

  constructor(
    private readonly themeService_: ThemeService,
    private readonly settings_: LocalSettingsService,
    private readonly overlayContainer_: OverlayContainer,
    private readonly kdRootRef: ElementRef,
    private readonly titleService_: TitleService,
  ) {
    this.isLightThemeEnabled_ = this.themeService_.isLightThemeEnabled();
  }

  ngOnInit(): void {
    this.titleService_.update();
    this.themeService_.subscribe(this.onThemeChange_.bind(this));

    const localSettings = this.settings_.get();
    if (localSettings && localSettings.isThemeDark) {
      this.themeService_.switchTheme(!localSettings.isThemeDark);
      this.isLightThemeEnabled_ = !localSettings.isThemeDark;
    }

    this.applyOverlayContainerTheme_();
  }

  private applyOverlayContainerTheme_(): void {
    const classToRemove = this.getTheme(!this.isLightThemeEnabled_);
    const classToAdd = this.getTheme(this.isLightThemeEnabled_);
    this.overlayContainer_.getContainerElement().classList.remove(classToRemove);
    this.overlayContainer_.getContainerElement().classList.add(classToAdd);

    this.kdRootRef.nativeElement.classList.add(classToAdd);
    this.kdRootRef.nativeElement.classList.remove(classToRemove);
  }

  private onThemeChange_(isLightThemeEnabled: boolean): void {
    this.isLightThemeEnabled_ = isLightThemeEnabled;
    this.applyOverlayContainerTheme_();
  }

  getTheme(isLightThemeEnabled?: boolean): string {
    if (isLightThemeEnabled === undefined) {
      isLightThemeEnabled = this.isLightThemeEnabled_;
    }

    return isLightThemeEnabled ? Themes.Light : Themes.Dark;
  }
}
