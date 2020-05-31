

import {Component} from '@angular/core';
import {VersionInfo} from '@api/frontendapi';
import {ConfigService} from '../../common/services/global/config';

@Component({selector: '', templateUrl: './template.html'})
export class ActionbarComponent {
  versionInfo: VersionInfo;

  constructor(config: ConfigService) {
    this.versionInfo = config.getVersionInfo();
  }
}
