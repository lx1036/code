import {NgModule} from "@angular/core";
import {RouterModule, Routes} from "@angular/router";
import {SignInComponent} from "./sign-in/sign-in.component";
import {SharedModule} from "../../shared.module";


const routes: Routes = [
  {
    path: 'sign-in', component: SignInComponent
  }
];
@NgModule({
  declarations: [
  ],
  imports: [
    SharedModule,
    RouterModule.forChild(routes),
  ]
})
export class AuthModule { }
