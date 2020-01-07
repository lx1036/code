import { NgModule } from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {SignInComponent} from "./sign-in.component";
import {SharedModule} from "./shared.module";

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
  declarations: [
    SignInComponent,
  ],
  imports: [
    SharedModule,
    AuthRoutingModule
  ]
})
export class AuthModule { }
