

import {Injectable} from '@angular/core';
import {MatDrawer} from '@angular/material/sidenav';

@Injectable({providedIn: 'root'})
export class NavService {
  private nav_: MatDrawer;

  toggle(): void {
    this.nav_.toggle();
  }

  setVisibility(isVisible: boolean): void {
    this.nav_.toggle(isVisible);
  }

  setNav(nav: MatDrawer): void {
    this.nav_ = nav;
  }
}
