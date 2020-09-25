import {NgModule} from '@angular/core';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {RouterModule, Routes} from '@angular/router';
import {BrowserModule} from '@angular/platform-browser';
import {FormsModule} from '@angular/forms';
import {HttpClientModule} from '@angular/common/http';
import {TranslateModule} from '@ngx-translate/core';
import {MessageService} from './message.service';
import {CacheService} from './cache.service';
import {InputComponent} from './input.component';
import {ClarityModule} from '@clr/angular';
import {MessageComponent} from './message.component';
import {DiffComponent} from './diff.component';
import {ConfirmationDialogComponent} from './confirmation-dialog.component';
import {DropdownComponent, DropdownItemComponent} from './dropdown.component';
import {NotificationService} from './notification.service';
import {PaginateComponent} from './paginate.component';
import {CardComponent} from './card.component';
import {BoxComponent} from './box.component';
import {ProgressComponent} from './progress.component';
import {SideNavFooterComponent} from './sidenav-footer.component';
import {AppService} from './app.service';
import {BreadcrumbComponent} from "./breadcrumb.component";
import {BreadcrumbService} from "./breadcrumb.service";
import {UserService} from "./user.service";


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
  ],
  exports: [
    BrowserModule,
    BrowserAnimationsModule,
    TranslateModule,
    FormsModule,
    RouterModule,

    ClarityModule,

  ],
  declarations: [
  ],
  providers: [
  ],
})
export class SharedModule {
}
