

// Override default material tooltip values.
import {MatTooltipDefaultOptions} from "@angular/material/tooltip";

export const KD_TOOLTIP_DEFAULT_OPTIONS: MatTooltipDefaultOptions = {
  showDelay: 500,
  hideDelay: 0,
  touchendHideDelay: 0,
};

export interface Config {
  authTokenCookieName: string;
  skipLoginPageCookieName: string;
  csrfHeaderName: string;
  authTokenHeaderName: string;
  defaultNamespace: string;
}

export const CONFIG: Config = {
  authTokenCookieName: 'jweToken',
  authTokenHeaderName: 'jweToken',
  csrfHeaderName: 'X-CSRF-TOKEN',
  skipLoginPageCookieName: 'skipLoginPage',
  defaultNamespace: 'default',
};
