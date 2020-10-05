import {APP_INITIALIZER, Injectable, Injector, NgModule} from '@angular/core';
import { AppComponent } from './app.component';
import {PortalModule} from './portal/portal.module';
import {AdminModule} from './admin/admin.module';
import {
  HTTP_INTERCEPTORS,
  HttpClient,
  HttpClientModule,
} from '@angular/common/http';
import {TranslateLoader, TranslateModule} from '@ngx-translate/core';
import {RoutingModule} from "./app-routing.module";
import {PodTerminalModule} from "./portal/pod-terminal/pod-terminal.module";
import {BrowserModule} from "@angular/platform-browser";
import {BrowserAnimationsModule} from "@angular/platform-browser/animations";
import {AuthModule} from "./shared/components/auth/auth.module";
import {AuthService} from "./shared/components/auth/auth.service";
import {AuthInterceptor, HttpLoaderFactory, initConfig, initUser} from "./app.service";

@NgModule({
  declarations: [
    AppComponent,
  ],
  imports: [
    // angular module
    BrowserModule,
    BrowserAnimationsModule,
    RoutingModule,
    HttpClientModule,
    TranslateModule.forRoot({
      loader: {
        provide: TranslateLoader,
        useFactory: HttpLoaderFactory,
        deps: [HttpClient]
      }
    }),

    // application module
    PodTerminalModule,
    PortalModule,
    AdminModule,
    AuthModule,
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
  exports: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
