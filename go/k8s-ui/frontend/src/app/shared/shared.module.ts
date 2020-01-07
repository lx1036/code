import {NgModule} from '@angular/core';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {RouterModule} from '@angular/router';
import {BrowserModule} from '@angular/platform-browser';
import {FormsModule} from '@angular/forms';
import {HttpClientModule} from '@angular/common/http';
import {TranslateModule} from '@ngx-translate/core';
import {MessageService} from "./message.service";
import {CacheService} from "./cache.service";


@NgModule({
  imports: [
    BrowserAnimationsModule,
    RouterModule,
    TranslateModule,
    BrowserModule,
    FormsModule,
    // ResourceLimitModule,
    HttpClientModule,
    // EchartsModule,
    // ClarityModule,
    // CollapseModule
  ],
  exports: [],
  declarations: [],
  providers: [
    MessageService,
    CacheService,
  ],
})
export class SharedModule {
}
