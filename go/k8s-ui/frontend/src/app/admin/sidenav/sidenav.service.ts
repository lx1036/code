import {Injectable} from '@angular/core';
import {Observable, Subject} from "rxjs";
import {Event, NavigationEnd, Router} from "@angular/router";

@Injectable()
export class SideNavService {

  private _routerChange: Subject<string> = new Subject<string>();
  get routerChange(): Observable<string> {
    return this._routerChange.asObservable();
  }
  constructor(private router: Router) {
    this.router.events.subscribe(
      (event:Event) => {
        if (event instanceof NavigationEnd) {
          this.routerChangeTrigger(event.url);
        }
      }
    );
  }
  routerChangeTrigger(url: string): void {
    this._routerChange.next(url);
  }
}
