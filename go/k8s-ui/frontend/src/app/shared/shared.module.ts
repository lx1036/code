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
import {NotificationService} from './notification.service';
import {PaginateComponent} from './paginate.component';
import {CardComponent} from './card.component';
import {BoxComponent} from './box.component';
import {NamespaceClient} from './client/v1/kubernetes/namespace';
import {ProgressComponent} from './progress.component';
import {SideNavFooterComponent} from './sidenav-footer.component';

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
    RouterModule,
    BrowserModule,
    TranslateModule,
    FormsModule,

    MessageComponent,
    DiffComponent,
    ConfirmationDialogComponent,
    DropdownItemComponent,
    DropdownComponent,
    PaginateComponent,
    CardComponent,
    BoxComponent,
    ProgressComponent,
    SideNavFooterComponent,
  ],
  declarations: [
    SignInComponent,
    InputComponent,
    MessageComponent,
    DiffComponent,
    ConfirmationDialogComponent,
    DropdownComponent,
    DropdownItemComponent,
    PaginateComponent,
    CardComponent,
    BoxComponent,
    ProgressComponent,
    SideNavFooterComponent,
  ],
  providers: [
    MessageService,
    CacheService,
    AuthoriseService,
    StorageService,
    NotificationService,
    NamespaceClient,
  ],
})
export class SharedModule {
}
