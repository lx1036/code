import { NgModule } from '@angular/core';
import {SharedModule} from '../shared.module';
import {RouterModule, Routes} from '@angular/router';
import {SignInComponent} from './sign-in/sign-in.component';

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
  declarations: [],
  imports: [
    SharedModule,
    AuthRoutingModule
  ]
})
export class AuthModule { }
