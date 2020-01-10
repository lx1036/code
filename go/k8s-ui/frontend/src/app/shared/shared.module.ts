import {NgModule} from '@angular/core';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {RouterModule, Routes} from '@angular/router';
import {BrowserModule} from '@angular/platform-browser';
import {FormsModule} from '@angular/forms';
import {HttpClientModule} from '@angular/common/http';
import {TranslateModule} from '@ngx-translate/core';
import {MessageService} from './message.service';
import {CacheService} from './cache.service';
import {AuthoriseService} from './client/v1/auth.service';
import {SignInComponent} from './sign-in.component';
import {InputComponent} from './input.component';

const routes: Routes = [
  {
    path: 'sign-in', component: SignInComponent
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class AuthRoutingModule {
}


@NgModule({
  imports: [
    BrowserAnimationsModule,
    RouterModule,
    TranslateModule,
    BrowserModule,
    FormsModule,
    // ResourceLimitModule,
    HttpClientModule,
    // EchartsModule,
    // ClarityModule,
    // CollapseModule

    AuthRoutingModule,
  ],
  exports: [],
  declarations: [
    SignInComponent,
    InputComponent,

  ],
  providers: [
    MessageService,
    CacheService,
    AuthoriseService
  ],
})
export class SharedModule {
}