

import {Component, Inject} from '@angular/core';
import {VersionInfo} from '@api/frontendapi';
import {AssetsService} from '../common/services/global/assets';
import {ConfigService} from '../common/services/global/config';

@Component({
  selector: 'kd-about',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class AboutComponent {
  latestCopyrightYear: number;
  versionInfo: VersionInfo;

  constructor(@Inject(AssetsService) public assets: AssetsService, config: ConfigService) {
    this.versionInfo = config.getVersionInfo();
    this.latestCopyrightYear = new Date().getFullYear();
  }
}
