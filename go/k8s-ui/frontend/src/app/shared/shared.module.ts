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
import {AppService} from './app.service';
import {BreadcrumbComponent} from "./breadcrumb.component";
import {BreadcrumbService} from "./breadcrumb.service";
import {UserService} from "./user.service";
import {PodClientService} from "./client/v1/kubernetes/pod.service";
import {NodeClientService} from "./client/v1/kubernetes/node.service";

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

    BoxComponent,
    BreadcrumbComponent,
    CardComponent,
    ConfirmationDialogComponent,
    DiffComponent,
    DropdownItemComponent,
    DropdownComponent,
    MessageComponent,
    PaginateComponent,
    ProgressComponent,
    SideNavFooterComponent,
  ],
  declarations: [
    BoxComponent,
    BreadcrumbComponent,
    CardComponent,
    ConfirmationDialogComponent,
    DiffComponent,
    DropdownComponent,
    DropdownItemComponent,
    InputComponent,
    MessageComponent,
    PaginateComponent,
    ProgressComponent,
    SideNavFooterComponent,
    SignInComponent,
  ],
  providers: [
    AppService,
    AuthoriseService,
    BreadcrumbService,
    CacheService,
    MessageService,
    NotificationService,
    NamespaceClient,
    NodeClientService,
    PodClientService,
    StorageService,
    UserService,
  ],
})
export class SharedModule {
}
