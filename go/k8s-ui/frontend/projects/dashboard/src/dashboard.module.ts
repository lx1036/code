import {BrowserModule, HAMMER_GESTURE_CONFIG} from "@angular/platform-browser";
import {BrowserAnimationsModule} from "@angular/platform-browser/animations";
import {HttpClientModule} from "@angular/common/http";
import {RouterModule, Routes} from "@angular/router";
import {ErrorHandler, NgModule} from "@angular/core";
import {RootComponent} from "./dashboard.component";
import {LoginComponent} from "./login/login.component";
import {LoginGuard} from "./login/login.guard";
import {SharedModule} from "./shared.module";
import {CardComponent} from "./common/components/card/card.component";
import {UploadFileComponent} from "./common/components/uploadfile/uploadfile.component";
import {AlertDialog} from "./common/components/dialog/dialog";



export const routes: Routes = [
  {path: 'login', component: LoginComponent, canActivate: [LoginGuard]},
  {path: '', redirectTo: '/overview', pathMatch: 'full'},
  {path: '**', redirectTo: '/overview'},
];


@NgModule({
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    // CoreModule,
    // ChromeModule,
    // LoginModule,
    RouterModule.forRoot(routes, {
      useHash: false,
      onSameUrlNavigation: 'reload',
    }),

    SharedModule,
  ],
  providers: [
    // {provide: ErrorHandler, useClass: GlobalErrorHandler},
    // {provide: HAMMER_GESTURE_CONFIG, useClass: GestureConfig},
  ],
  declarations: [
    RootComponent,

    LoginComponent,
    CardComponent,
    UploadFileComponent,
    AlertDialog,
  ],
  bootstrap: [RootComponent],
})
export class RootModule {}
