import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import {UnauthorizedComponent} from './shared/unauthorized.component';
import {PageNotFoundComponent} from './shared/page-not-found.component';


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
