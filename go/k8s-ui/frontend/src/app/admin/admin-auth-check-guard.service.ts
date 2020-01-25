import {Injectable} from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, CanActivateChild, RouterStateSnapshot} from '@angular/router';
import {AuthService} from '../shared/auth.service';

@Injectable()
export class AdminAuthCheckGuard implements CanActivate, CanActivateChild {

  constructor(public authService: AuthService, ) {}

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean  {
    return new Promise((resolve, reject) => {
      if (!this.authService.currentUser) {
        this.authService.retrieveUser().then(user => {
          this.authService.currentUser = user;
        }).catch((error) => {

        });
      }

      if (this.authService.currentUser.admin) {
        resolve(true);
      } else {
        resolve(false);
      }
    });
  }

  canActivateChild(childRoute: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return this.canActivate(childRoute, state);
  }
}
