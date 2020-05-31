

import {Component, OnInit} from '@angular/core';
import {LocalSettings} from '@api/backendapi';
import {LocalSettingsService} from '../../common/services/global/localsettings';

@Component({selector: 'kd-local-settings', templateUrl: './template.html'})
export class LocalSettingsComponent implements OnInit {
  settings: LocalSettings = {} as LocalSettings;

  constructor(private readonly settings_: LocalSettingsService) {}

  ngOnInit(): void {
    this.settings = this.settings_.get();
  }

  onThemeChange(): void {
    this.settings_.handleThemeChange(this.settings.isThemeDark);
  }
}
