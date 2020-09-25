import {SharedModule} from "../../shared/shared.module";
import {NamespaceApiKeyComponent} from "./apikey.component";
import {ListApiKeyComponent} from "./list-apikey/list-apikey.component";
import {TokenDetailComponent} from "./token-detail/token-detail";
import {CreateEditApiKeyComponent} from "./create-edit-apikey/create-edit-apikey.component";
import {ApiKeyService} from "./apikey.service";


@NgModule({
  imports: [
    SharedModule,
  ],
  providers: [
    ApiKeyService,
  ],
  exports: [],
  declarations: [
    NamespaceApiKeyComponent,
    ListApiKeyComponent,
    TokenDetailComponent,
    CreateEditApiKeyComponent,
  ]
})

export class NamespaceApiKeyModule {
}
