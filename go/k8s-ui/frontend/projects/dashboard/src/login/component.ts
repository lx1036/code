

import {HttpClient, HttpErrorResponse} from '@angular/common/http';
import {Component, NgZone, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {
  AuthenticationMode,
  EnabledAuthenticationModes,
  LoginSkippableResponse,
  LoginSpec,
} from '@api/backendapi';
import {KdError, KdFile, StateError} from '@api/frontendapi';
import {map} from 'rxjs/operators';

import {AsKdError, K8SError} from '../common/errors/errors';
import {AuthService} from '../common/services/global/authentication';
import {PluginsConfigService} from '../common/services/global/plugin';

enum LoginModes {
  Kubeconfig = 'kubeconfig',
  Basic = 'basic',
  Token = 'token',
}

@Component({
  selector: 'kd-login',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class LoginComponent implements OnInit {
  loginModes = LoginModes;
  selectedAuthenticationMode = LoginModes.Kubeconfig;
  errors: KdError[] = [];

  private enabledAuthenticationModes_: AuthenticationMode[] = [];
  private isLoginSkippable_ = false;
  private kubeconfig_: string;
  private token_: string;
  private username_: string;
  private password_: string;

  constructor(
    private readonly authService_: AuthService,
    private readonly state_: Router,
    private readonly http_: HttpClient,
    private readonly ngZone_: NgZone,
    private readonly route_: ActivatedRoute,
    private readonly pluginConfigService_: PluginsConfigService,
  ) {}

  ngOnInit(): void {
    this.http_
      .get<EnabledAuthenticationModes>('api/v1/login/modes')
      .subscribe((enabledModes: EnabledAuthenticationModes) => {
        this.enabledAuthenticationModes_ = enabledModes.modes;
        this.enabledAuthenticationModes_.push(LoginModes.Kubeconfig);
        this.selectedAuthenticationMode = this.enabledAuthenticationModes_[0] as LoginModes;
      });

    this.http_
      .get<LoginSkippableResponse>('api/v1/login/skippable')
      .subscribe((loginSkippableResponse: LoginSkippableResponse) => {
        this.isLoginSkippable_ = loginSkippableResponse.skippable;
      });

    this.route_.paramMap.pipe(map(() => window.history.state)).subscribe((state: StateError) => {
      if (state.error) {
        this.errors = [state.error];
      }
    });
  }

  getEnabledAuthenticationModes(): AuthenticationMode[] {
    return this.enabledAuthenticationModes_;
  }

  login(): void {
    this.authService_.login(this.getLoginSpec_()).subscribe(
      (errors: K8SError[]) => {
        if (errors.length > 0) {
          this.errors = errors.map((error: K8SError) =>
            new K8SError(error.ErrStatus).toKdError().localize(),
          );
          return;
        }

        this.pluginConfigService_.refreshConfig();
        this.ngZone_.run(() => {
          this.state_.navigate(['overview']);
        });
      },
      (err: HttpErrorResponse) => {
        this.errors = [AsKdError(err)];
      },
    );
  }

  skip(): void {
    this.authService_.skipLoginPage(true);
    this.state_.navigate(['overview']);
  }

  isSkipButtonEnabled(): boolean {
    return this.isLoginSkippable_;
  }

  isLoginEnabled(): boolean {
    return this.authService_.domainWhitelist.indexOf(location.hostname) > -1
      ? true
      : location.protocol === this.authService_.allowedProtocol;
  }

  onChange(event: Event & KdFile): void {
    switch (this.selectedAuthenticationMode) {
      case LoginModes.Kubeconfig:
        this.onFileLoad_(event as KdFile);
        break;
      case LoginModes.Token:
        this.token_ = (event.target as HTMLInputElement).value;
        break;
      case LoginModes.Basic:
        if ((event.target as HTMLInputElement).id === 'username') {
          this.username_ = (event.target as HTMLInputElement).value;
        } else {
          this.password_ = (event.target as HTMLInputElement).value;
        }
        break;
      default:
    }
  }

  private onFileLoad_(file: KdFile): void {
    this.kubeconfig_ = file.content;
  }

  private getLoginSpec_(): LoginSpec {
    switch (this.selectedAuthenticationMode) {
      case LoginModes.Kubeconfig:
        return {kubeConfig: this.kubeconfig_} as LoginSpec;
      case LoginModes.Token:
        return {token: this.token_} as LoginSpec;
      case LoginModes.Basic:
        return {
          username: this.username_,
          password: this.password_,
        } as LoginSpec;
      default:
        return {} as LoginSpec;
    }
  }
}
