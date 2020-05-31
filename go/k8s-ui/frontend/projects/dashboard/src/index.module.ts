

import {HttpClientModule} from '@angular/common/http';
import {ErrorHandler, NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {RouterModule} from '@angular/router';
import {AngularPageVisibilityModule} from 'angular-page-visibility';
import {ChromeModule} from './chrome/module';
import {CoreModule} from './core.module';
import {GlobalErrorHandler} from './error/handler';
import {RootComponent} from './index.component';
import {routes} from './index.routing';
import {LoginModule} from './login/module';

@NgModule({
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    CoreModule,
    ChromeModule,
    LoginModule,
    AngularPageVisibilityModule,
    RouterModule.forRoot(routes, {
      useHash: true,
      onSameUrlNavigation: 'reload',
    }),
  ],
  providers: [{provide: ErrorHandler, useClass: GlobalErrorHandler}],
  declarations: [RootComponent],
  bootstrap: [RootComponent],
})
export class RootModule {}
