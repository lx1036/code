import {Injectable} from "@angular/core";
import {ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot, UrlTree} from "@angular/router";
import {Observable, of} from "rxjs";
import {AuthService} from "../common/services/global/authentication";
import {catchError, first, switchMap} from "rxjs/operators";
import {LoginStatus} from "../typings/backendapi";


@Injectable()
export class LoginGuard implements CanActivate{
  constructor(private readonly authService: AuthService, private readonly router: Router) {}

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean | UrlTree> {
    return this.authService
    .getLoginStatus()
    .pipe(first())
    .pipe(
      switchMap((loginStatus: LoginStatus) => {
        if (!this.authService.isAuthenticationEnabled(loginStatus)) {
          return this.router.navigate(['overview']);
        }

        return of(true);
      }),
      catchError(() => {
        return this.router.navigate(['overview']);
      })
    );
  }
}
