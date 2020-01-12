import {Injectable} from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, CanActivateChild, RouterStateSnapshot} from '@angular/router';
import {AuthService} from './auth.service';

@Injectable()
export class AuthCheckGuard implements CanActivate, CanActivateChild {

  constructor(public authService: AuthService, ) {
  }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return new Promise((resolve, reject) => {
      if (!this.authService.currentUser) {
        this.authService.retrieveUser().then(user => {
          console.log(user);

          resolve(true);
        }).catch((error) => {
          return resolve(false);
        });
      } else {
        console.log(this.authService.currentUser);
        return resolve(true);
      }
    });
  }

  canActivateChild(childRoute: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return this.canActivate(childRoute, state);
  }
}
