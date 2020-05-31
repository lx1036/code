

import {Injectable} from '@angular/core';
import {CanActivate, Router, UrlTree} from '@angular/router';
import {LoginStatus} from '@api/backendapi';
import {Observable, of} from 'rxjs';
import {first, switchMap} from 'rxjs/operators';
import {AuthService} from '../global/authentication';

@Injectable()
export class LoginGuard implements CanActivate {
  constructor(private readonly authService_: AuthService, private readonly router_: Router) {}

  canActivate(): Observable<boolean | UrlTree> {
    return this.authService_
      .getLoginStatus()
      .pipe(first())
      .pipe(
        switchMap((loginStatus: LoginStatus) => {
          if (!this.authService_.isAuthenticationEnabled(loginStatus)) {
            return this.router_.navigate(['overview']);
          }

          return of(true);
        }),
      )
      .catch(() => {
        return this.router_.navigate(['overview']);
      });
  }
}
