

import {Injectable, Injector} from '@angular/core';
import {NavigationEnd, Router} from '@angular/router';
import {filter, pairwise} from 'rxjs/operators';

@Injectable()
export class HistoryService {
  private router_: Router;
  private previousStateUrl_: string;
  private currentStateUrl_: string;

  constructor(private readonly injector_: Injector) {}

  /** Initializes the service. Must be called before use. */
  init(): void {
    this.router_ = this.injector_.get(Router);

    this.router_.events
      .pipe(filter(e => e instanceof NavigationEnd))
      .pipe(pairwise())
      .subscribe((e: [NavigationEnd, NavigationEnd]) => {
        this.previousStateUrl_ = e[0].url;
        this.currentStateUrl_ = e[1].url;
      });
  }

  /**
   * Goes back to previous state or to the provided defaultState if none set.
   */
  goToPreviousState(defaultState: string): Promise<boolean> {
    if (this.previousStateUrl_ && this.previousStateUrl_ !== this.currentStateUrl_) {
      return this.router_.navigateByUrl(this.previousStateUrl_);
    }

    return this.router_.navigate([defaultState], {queryParamsHandling: 'preserve'});
  }
}
