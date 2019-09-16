import {Component, ElementRef, Injectable, Injector, NgModule, OnInit, TemplateRef, ViewEncapsulation} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {
  ActivatedRoute,
  ActivatedRouteSnapshot, CanActivate, CanActivateChild,
  Resolve,
  Router,
  RouterModule,
  RouterStateSnapshot,
  Routes,
  RoutesRecognized
} from '@angular/router';
import {filter} from 'rxjs/operators';
import {Observable, of} from 'rxjs';

/**
 * https://angular.cn/guide/router
 *
 * 链接参数数组 link parameter array：(TODO:)
 */

@Component({
  selector: 'styled-shadow-comp',
  template: '<div class="red">StyledShadowComponent</div>',
  encapsulation: ViewEncapsulation.ShadowDom,
  styles: [`:host { border: 1px solid black; } .red { border: 1px solid red; }`]
})
class StyledShadowComponent {
}

@Component({
  selector: 'advisor',
  template: `
    <a routerLink="households">Advisor</a>
    <styled-shadow-comp></styled-shadow-comp>
    <router-outlet></router-outlet>
  `,
  styles: [
    `
      a {
          background: #ff3d00;
      }
      :host {
          border: 2px solid black;
      }
    `
  ],
  // encapsulation: ViewEncapsulation.ShadowDom
})
export class AdvisorComponent implements OnInit {
  constructor(private _route: ActivatedRoute, private _injector: Injector, private element: ElementRef) {}

  ngOnInit() {
    this._route.params.subscribe(params => {
      // console.log(params);
    });

    // console.log(this.element.nativeElement);

    // console.log(this._injector.get(TemplateRef));

    // this._router.routerState.root.firstChild.params.subscribe(params => {
    //   console.log(params);
    // });
    // console.log(this._router.routerState, this._router.routerState.root.children, this._router.routerState.toString());
  }
}

@Component({
  selector: 'household',
  template: `
    <p>Household</p>
  `
})
export class HouseholdComponent implements OnInit {
  constructor(private _route: ActivatedRoute, private _router: Router) {}

  ngOnInit() {
    this._route.params.subscribe(params => {
      console.log(params);
    });

    // console.log(this._router.routerState.root.children);


    this._router.routerState.root.firstChild.params.subscribe(params => {
      // console.log(params);
    });

    // console.log(this._router.routerState.root.params);
  }
}


@Component({
  selector: 'custom-header',
  template: `
    <nav>
      <a routerLink="/advisor/1/household/1">Nav1</a>
      <a routerLink="/advisor/2/household/2">Nav2</a>
    </nav>
  `
})
export class DemoCustomHeader implements OnInit {
  constructor(private _router: Router, private _route: ActivatedRoute) {}

  ngOnInit() {
    console.log(this._router.routerState.root.children);

    this._router.events.pipe(
      filter(event => event instanceof RoutesRecognized),
    ).subscribe((event: RoutesRecognized) => {
      // console.log(event.state.root, event.state.root.firstChild.params, event.state.root.firstChild.firstChild.params, event.url, event.urlAfterRedirects);
    });
    // this._router.events.pipe(filter(events => events instanceof RoutesRecognized)).subscribe(
    //   (event: RoutesRecognized) => {
    //   console.log(event.state.root.firstChild.params, event.state.root.firstChild.children[0].params);
    //
    //   console.log(event.state.root.firstChild, event.state.root.children);
    // });
    //
    // console.log(this._router.routerState.root.children);
    //
    // this._route.params.subscribe(params => {
    //   console.log(params);
    // });
  }
}

@Component({
  selector: 'demo-router',
  template: `
    <custom-header></custom-header>
    <router-outlet></router-outlet>
  `
})
export class DemoRouter {

}



@Injectable({
  providedIn: 'root'
})
export class TestCanActivate implements CanActivate {
  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> {
    console.log(TestCanActivate.name, 'lx1036');

    return of(true);
  }
}

@Injectable({
  providedIn: 'root'
})
export class TestCanActivateChild implements CanActivateChild {
  canActivateChild(childRoute: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> {
    console.log(TestCanActivateChild.name, 'lx1037');

    return of(true);
  }
}

const routes: Routes = [
  {
    path: '',
    redirectTo: 'advisor/1',
    pathMatch: 'full'
  },
  {
    path: 'advisor/:advisorId',
    component: AdvisorComponent,
    canActivate: [TestCanActivate],
    canActivateChild: [TestCanActivateChild],
    children: [
      {
        path: 'household/:householdId',
        component: HouseholdComponent,
      },
      {
        path: 'households',
        component: HouseholdComponent,
      }
    ]
  },
];

@NgModule({
  imports: [BrowserModule, RouterModule.forRoot(routes, {paramsInheritanceStrategy: "always"})],
  declarations: [
    DemoRouter,
    DemoCustomHeader,
    AdvisorComponent,
    HouseholdComponent,
    StyledShadowComponent,
  ],
  bootstrap: [DemoRouter]
})
export class DemoRouterModule {

}