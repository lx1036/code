import {Component, Injector, NgModule} from '@angular/core';
import {Router, RouterModule} from '@angular/router';
import {AComponent} from './app.module';
import {Observable, interval} from 'rxjs';


@Component({
  selector: 'lazy-load',
  template: `
    <p>Lazy load component</p>
  `
})
export class LazyLoadComponent {
  constructor(private router: Router, private _injector: Injector) {
    console.log(router.url);

    console.log(_injector.get(Router), _injector);

    // const interval$ = interval(3000);
  }
}


@NgModule({
  declarations: [
    LazyLoadComponent
  ],
  imports: [
    RouterModule.forChild([
      {path: '', redirectTo: 'lazy', pathMatch: 'full'},
      {path: 'lazy', component: LazyLoadComponent}
    ]),
  ],
  exports: [
    LazyLoadComponent
  ]
})
export class LazyLoadModule {}

