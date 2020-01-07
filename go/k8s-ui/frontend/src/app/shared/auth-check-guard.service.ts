import {Injectable} from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, CanActivateChild, RouterStateSnapshot} from "@angular/router";

@Injectable()
export class AuthCheckGuard implements CanActivate, CanActivateChild {

  constructor() {
  }
  
  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return false;
  }
  
  canActivateChild(childRoute: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return false;
  }
}
