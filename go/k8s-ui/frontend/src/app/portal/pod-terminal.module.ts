import { NgModule } from '@angular/core';
import {PodTerminalComponent} from './pod-terminal.component';
import {SharedModule} from "../shared/shared.module";



@NgModule({
  declarations: [
    PodTerminalComponent
  ],
  imports: [
    SharedModule,
  ]
})
export class PodTerminalModule { }
