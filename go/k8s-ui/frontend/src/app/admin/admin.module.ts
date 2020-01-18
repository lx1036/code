import { NgModule } from '@angular/core';
import {CronjobComponent} from './cronjob.component';
import {CreateNotificationComponent, ListNotificationComponent, NotificationComponent} from './notification.component';
import {SharedModule} from '../shared/shared.module';
import {RouterModule, Routes} from '@angular/router';
import {AdminComponent} from './admin.component';
import {AdminAuthCheckGuard} from './admin-auth-check-guard.service';
import {NavComponent} from './nav.component';
import {OverviewComponent} from './overview.component';
import {MarkdownModule} from 'ngx-markdown';


const routes: Routes = [
  {
    path: 'admin',
    component: AdminComponent,
    canActivate: [AdminAuthCheckGuard],
    canActivateChild: [AdminAuthCheckGuard],
    children: [
      {
        path: '',
        pathMatch: 'full',
        redirectTo: 'reportform/overview'
      },
      {path: 'notification', component: NotificationComponent},
      {path: 'reportform/overview', component: OverviewComponent},
    ]
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class AdminRoutingModule {
}

@NgModule({
  declarations: [
    CronjobComponent,
    NotificationComponent,
    CreateNotificationComponent,
    ListNotificationComponent,
    AdminComponent,
    NavComponent,
    OverviewComponent,
  ],
  imports: [
    SharedModule,
    AdminRoutingModule,
    MarkdownModule.forRoot(),
  ],
  providers: [
    AdminAuthCheckGuard,
  ]
})
export class AdminModule { }
