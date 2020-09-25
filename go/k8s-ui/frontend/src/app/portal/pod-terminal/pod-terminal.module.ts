import { NgModule } from '@angular/core';
import {PodTerminalComponent} from './pod-terminal.component';
import {RouterModule, Routes} from "@angular/router";


const routes: Routes = [
  {
    path: 'portal/namespace/:nid/app/:id/:resourceType/:resourceName/pod/:podName/terminal/:cluster/:namespace',
    component: PodTerminalComponent
  },
  {
    path: 'portal/namespace/:nid/app/:id/:resourceType/:resourceName/pod/:podName/container/:container/terminal/:cluster/:namespace',
    component: PodTerminalComponent
  }
];

@NgModule({
  declarations: [
    PodTerminalComponent
  ],
  imports: [
    RouterModule.forChild(routes),
  ]
})
export class PodTerminalModule { }
