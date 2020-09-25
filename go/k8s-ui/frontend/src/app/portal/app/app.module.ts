import {SharedModule} from "../../shared/shared.module";
import {AppComponent} from "./app.component";
import {ListAppsComponent} from "./list-app/list-apps.component";
import {NgModule} from "@angular/core";


@NgModule({
  imports: [
    SharedModule,
  ],
  providers: [
  ],
  exports: [
    AppComponent,
    ListAppsComponent,
  ],
  declarations: [
    AppComponent,
    ListAppsComponent,
  ]
})

export class AppModule {
}
