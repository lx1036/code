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
import {ClarityModule} from '@clr/angular';
import {MessageComponent} from './message.component';
import {DiffComponent} from './diff.component';
import {ConfirmationDialogComponent} from './confirmation-dialog.component';
import {DropdownComponent, DropdownItemComponent} from './dropdown.component';
import {StorageService} from './storage.service';

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
    BrowserModule,
    FormsModule,
    // ResourceLimitModule,
    HttpClientModule,
    // EchartsModule,
    ClarityModule, // https://clarity.design/documentation/get-started
    // CollapseModule
    TranslateModule,
    AuthRoutingModule,
  ],
  exports: [
    ClarityModule,

    MessageComponent,
    DiffComponent,
    ConfirmationDialogComponent,
    DropdownItemComponent,
    DropdownComponent,
  ],
  declarations: [
    SignInComponent,
    InputComponent,
    MessageComponent,
    DiffComponent,
    ConfirmationDialogComponent,
    DropdownComponent,
    DropdownItemComponent,
  ],
  providers: [
    MessageService,
    CacheService,
    AuthoriseService,
    StorageService,
  ],
})
export class SharedModule {
}
