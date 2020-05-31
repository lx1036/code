

import {Component, Injector, Input, OnInit, ViewChild, ViewContainerRef} from '@angular/core';

import {PluginLoaderService} from '../common/services/pluginloader/pluginloader.service';

@Component({
  selector: 'kd-plugin-holder',
  template: `
    <div>
      <div class="plugin">
        <mat-card *ngIf="entryError">This plugin has no entry component</mat-card>
        <ng-template #pluginViewRef #elseBlock></ng-template>
      </div>
    </div>
  `,
})
export class PluginHolderComponent implements OnInit {
  @ViewChild('pluginViewRef', {read: ViewContainerRef, static: true}) vcRef: ViewContainerRef;
  @Input('pluginName') private pluginName: string;
  entryError = false;

  constructor(private injector: Injector, private pluginLoader: PluginLoaderService) {}

  ngOnInit() {
    try {
      this.loadPlugin(this.pluginName);
    } catch (e) {
      console.log(e);
    }
  }

  loadPlugin(pluginName: string) {
    this.pluginLoader.load(pluginName).then(moduleFactory => {
      const moduleRef = moduleFactory.create(this.injector);
      // tslint:disable-next-line:no-any
      const entryComponent = (moduleFactory.moduleType as any).entry;
      try {
        const compFactory = moduleRef.componentFactoryResolver.resolveComponentFactory(
          entryComponent,
        );
        this.vcRef.createComponent(compFactory);
      } catch (e) {
        this.entryError = true;
      }
    });
  }
}
