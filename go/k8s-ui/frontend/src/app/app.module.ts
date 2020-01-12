import {APP_INITIALIZER, Injectable, Injector, NgModule} from '@angular/core';
import { AppComponent } from './app.component';
import {PortalModule} from './portal/portal.module';
import {AdminModule} from './admin/admin.module';
import {
  HTTP_INTERCEPTORS,
  HttpClient,
  HttpClientModule,
  HttpErrorResponse, HttpEvent,
  HttpHandler, HttpHeaders,
  HttpInterceptor,
  HttpRequest
} from '@angular/common/http';
import {TranslateLoader, TranslateModule} from '@ngx-translate/core';
import {TranslateHttpLoader} from '@ngx-translate/http-loader';
import {PodTerminalModule} from './portal/pod-terminal.module';
import {AuthService} from './shared/auth.service';
import {httpStatusCode, LoginTokenKey} from './shared/shared.const';
import {Router, RouterModule, Routes} from '@angular/router';
import {Observable} from 'rxjs';
import {Location} from '@angular/common';
const packageJson = require('../../package.json');
import { environment} from '../environments/environment';
import {UnauthorizedComponent} from './shared/unauthorized.component';
import {PageNotFoundComponent} from './shared/page-not-found.component';

export function HttpLoaderFactory(httpClient: HttpClient) {
  return new TranslateHttpLoader(httpClient, './assets/i18n/', '.json?v=' + packageJson.version);
}

function initUser(authService: AuthService, injector: Injector) {
  return () => authService.retrieveUser().catch((error: HttpErrorResponse) => {
    console.log('error status: ' + error.status);
    if (error.status ===  httpStatusCode.Unauthorized) {
      injector.get(Router).navigateByUrl('sign-in')
      .then(() => console.log('navigate to sign-in page'))
      .catch((err: HttpErrorResponse) => console.log('can\'t navigate to sign-in page because of:'  + err.message ));
    }
  });
}

function initConfig(authService: AuthService) {
  return () => authService.initConfig();
}

@Injectable()
class AuthInterceptor implements HttpInterceptor {
  constructor(private location: Location) {}

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const token = localStorage.getItem(LoginTokenKey);
    const headers: {[name: string]: string|string[]} = {};
    for (const key of request.headers.keys()) {
      headers[key] = request.headers.getAll(key);
    }
    headers['Content-Type'] = 'application/json';
    if (token) { // if logged in
      headers.Authorization = 'Bearer ' + token;
    }
    const url = this.location.normalize(environment.api) + this.location.prepareExternalUrl(request.url);
    const req = request.clone({url, headers: new HttpHeaders(headers)});

    return next.handle(req).pipe();
  }
}

const routes: Routes = [
  {path: '', redirectTo: 'portal/namespace/0/app', pathMatch: 'full'},
  {path: 'unauthorized', component: UnauthorizedComponent},
  {path: '**', component: PageNotFoundComponent},
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule],
  declarations: [
    UnauthorizedComponent,
    PageNotFoundComponent,
  ]
})
export class RoutingModule { }

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
