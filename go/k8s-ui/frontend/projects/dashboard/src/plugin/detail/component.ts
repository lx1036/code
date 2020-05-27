

import {Component} from '@angular/core';
import {ActivatedRoute} from '@angular/router';

@Component({
  selector: 'kd-plugin-detail',
  template: `
    <kd-plugin-holder [pluginName]="this.pluginName()"></kd-plugin-holder>
  `,
})
export class PluginDetailComponent {
  constructor(private readonly activatedRoute_: ActivatedRoute) {}

  pluginName(): string {
    return this.activatedRoute_.snapshot.params.pluginName;
  }
}
