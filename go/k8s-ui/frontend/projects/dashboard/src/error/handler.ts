

import {HttpErrorResponse} from '@angular/common/http';
import {ErrorHandler, Injectable, Injector, NgZone} from '@angular/core';
import {Router} from '@angular/router';
import {StateError} from '@api/frontendapi';
import {YAMLException} from 'js-yaml';

import {ApiError, AsKdError, KdError} from '../common/errors/errors';
import {AuthService} from '../common/services/global/authentication';

@Injectable()
export class GlobalErrorHandler implements ErrorHandler {
  constructor(private readonly injector_: Injector, private readonly ngZone_: NgZone) {}

  private get router_(): Router {
    return this.injector_.get(Router);
  }

  private get auth_(): AuthService {
    return this.injector_.get(AuthService);
  }

  handleError(error: HttpErrorResponse | YAMLException): void {
    if (error instanceof HttpErrorResponse) {
      this.handleHTTPError_(error);
      return;
    }

    if (error instanceof YAMLException) {
      return;
    }

    throw error;
  }

  private handleHTTPError_(error: HttpErrorResponse): void {
    this.ngZone_.run(() => {
      if (KdError.isError(error, ApiError.tokenExpired, ApiError.encryptionKeyChanged)) {
        this.router_.navigate(['login'], {
          state: {error: AsKdError(error)} as StateError,
        });
        this.auth_.removeAuthCookies();
        return;
      }

      if (!this.router_.routerState.snapshot.url.includes('error')) {
        this.router_.navigate(['error'], {
          queryParamsHandling: 'preserve',
          state: {error: AsKdError(error)} as StateError,
        });
      }
    });
  }
}
