import {APP_INITIALIZER, Injectable, Injector, NgModule} from '@angular/core';
import {RoutingModule} from './app-routing.module';
import { AppComponent } from './app.component';
import {PortalModule} from './portal/portal.module';
import {AdminModule} from './admin/admin.module';
import {
  HTTP_INTERCEPTORS,
  HttpClient,
  HttpClientModule,
  HttpErrorResponse, HttpEvent,
  HttpHandler,
  HttpInterceptor,
  HttpRequest
} from '@angular/common/http';
import {TranslateLoader, TranslateModule} from '@ngx-translate/core';
import {TranslateHttpLoader} from '@ngx-translate/http-loader';
import {PodTerminalModule} from './portal/pod-terminal.module';
import {AuthModule} from './shared/auth.module';
import {AuthService} from './shared/auth.service';
import {httpStatusCode} from './shared/shared.const';
import {Router} from '@angular/router';
import {Observable} from 'rxjs';
import {Location} from '@angular/common';
const packageJson = require('../../package.json');
import { environment} from '../environments/environment';

export function HttpLoaderFactory(httpClient: HttpClient) {
  return new TranslateHttpLoader(httpClient, './assets/i18n/', '.json?v=' + packageJson.version);
}


function initUser(authService: AuthService, injector: Injector) {
  return () => authService.retrieveUser().then(user => {}).catch((error: HttpErrorResponse) => {
    console.log('error status: ' + error.status);
    if (error.status ===  httpStatusCode.Unauthorized) {
      injector.get(Router).navigateByUrl('sign-in')
      .then(() => console.log('navigate to sign-in page'))
      .catch((err: HttpErrorResponse) => console.log('can\'t navigate to sign-in page because of:'  + err.message ));
    }
  });
}

function initConfig(authService: AuthService) {
  return () => {};
}


@Injectable()
class AuthInterceptor implements HttpInterceptor {
  constructor(private location: Location) {}

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const url = this.location.normalize(environment.api) + this.location.prepareExternalUrl(request.url);
    const req = request.clone({url});

    return next.handle(req).pipe();
  }
}


@NgModule({
  declarations: [
    AppComponent,
    // FilterBoxComponent,
    // CheckboxGroupComponent,
    // CheckboxComponent,
    // ConfirmationDialogComponent
  ],
  imports: [
    PodTerminalModule,
    AuthModule,
    PortalModule,
    AdminModule,
    RoutingModule,
    HttpClientModule,
    // TranslateModule.forRoot({
    //   loader: {
    //     provide: TranslateLoader,
    //     useFactory: HttpLoaderFactory,
    //     deps: [HttpClient]
    //   }
    // })
  ],
  providers: [
    AuthService,
    {
      provide: APP_INITIALIZER,
      useFactory: initUser,
      deps: [AuthService, Injector],
      multi: true
    },
    {
      provide: APP_INITIALIZER,
      useFactory: initConfig,
      deps: [AuthService],
      multi: true
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthInterceptor,
      multi: true
    },
  ],
  exports: [
    // FilterBoxComponent,
    // CheckboxComponent
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
