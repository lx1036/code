

import {Injectable} from '@angular/core';
import {Title} from '@angular/platform-browser';

import {GlobalSettingsService} from './globalsettings';

@Injectable()
export class TitleService {
  clusterName = '';

  constructor(private readonly title_: Title, private readonly settings_: GlobalSettingsService) {}

  update(): void {
    this.settings_.load(
      () => {
        this.clusterName = this.settings_.getClusterName();
        this.apply_();
      },
      () => {
        this.clusterName = '';
        this.apply_();
      },
    );
  }

  private apply_(): void {
    let title = 'Kubernetes Dashboard';

    if (this.clusterName && this.clusterName.length > 0) {
      title = `${this.clusterName} - ` + title;
    }

    this.title_.setTitle(title);
  }
}
