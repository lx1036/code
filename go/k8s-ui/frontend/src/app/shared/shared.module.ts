import {NgModule} from '@angular/core';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {RouterModule, Routes} from '@angular/router';
import {BrowserModule} from '@angular/platform-browser';
import {FormsModule} from '@angular/forms';
import {HttpClientModule} from '@angular/common/http';
import {TranslateModule} from '@ngx-translate/core';
import {ClarityModule} from '@clr/angular';
import {CommonModule} from "@angular/common";
import {CardComponent} from "./components/card/card.component";
import {ProgressComponent} from "./components/progress/progress.component";
import {BoxComponent} from "./components/box/box.component";
import {FilterBoxComponent} from "./components/filter-box/filter-box.component";
import {CheckboxComponent} from "./components/checkbox/checkbox.component";
import {CheckboxGroupComponent} from "./components/checkbox/checkbox-group/checkbox-group.component";
import {InputComponent} from "./components/input/input.component";


@NgModule({
  imports: [
    BrowserAnimationsModule,
    RouterModule,
    BrowserModule,
    FormsModule,
    CommonModule,
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
    CommonModule,

    ClarityModule,

    CardComponent,
    ProgressComponent,
    BoxComponent,
    FilterBoxComponent,
    CheckboxComponent,
    CheckboxGroupComponent,
    InputComponent,
  ],
  declarations: [
    CardComponent,
    ProgressComponent,
    BoxComponent,
    FilterBoxComponent,
    CheckboxComponent,
    CheckboxGroupComponent,
    InputComponent,
  ],
  providers: [
  ],
})
export class SharedModule {
}
