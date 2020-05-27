import {BrowserModule, HAMMER_GESTURE_CONFIG} from "@angular/platform-browser";
import {BrowserAnimationsModule} from "@angular/platform-browser/animations";
import {HttpClientModule} from "@angular/common/http";
import {RouterModule, Routes} from "@angular/router";
import {ErrorHandler, NgModule} from "@angular/core";
import {RootComponent} from "./dashboard.component";
import {LoginGuard} from "./login/login.guard";
import {SharedModule} from "./shared.module";
import {UploadFileComponent} from "./common/components/uploadfile/uploadfile.component";
import {AlertDialog} from "./common/components/dialog/dialog";
import {AuthService} from "./common/services/global/authentication";
import {CsrfTokenService} from "./common/services/global/csrftoken";
import {CookieService} from "ngx-cookie-service";
import {CoreModule} from "./core.module";
import {LoginModule} from "./login/module";
import {routes} from "./index.routing";

@NgModule({
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    CoreModule,
    // ChromeModule,
    LoginModule,
    RouterModule.forRoot(routes, {
      useHash: false,
      onSameUrlNavigation: 'reload',
    }),

    SharedModule,
  ],
  providers: [
    LoginGuard,
    AuthService,
    CsrfTokenService,
    CookieService,
    // {provide: ErrorHandler, useClass: GlobalErrorHandler},
    // {provide: HAMMER_GESTURE_CONFIG, useClass: GestureConfig},
  ],
  declarations: [
    RootComponent,
    UploadFileComponent,
    AlertDialog,
  ],
  bootstrap: [RootComponent],
})
export class RootModule {}
