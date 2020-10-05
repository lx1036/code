import {RouterModule, Routes} from "@angular/router";
import {NgModule} from "@angular/core";
import {ServiceComponent} from "./portal/service/service.component";
import {UnauthorizedComponent} from "./shared/components/auth/unauthorized/unauthorized.component";
import {PageNotFoundComponent} from "./shared/components/auth/not-found/page-not-found.component";


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
