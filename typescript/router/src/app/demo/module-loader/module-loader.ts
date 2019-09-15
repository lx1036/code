

import {
  Component,
  Injector, NgModule,
  NgModuleFactoryLoader,
  SystemJsNgModuleLoader,
  ViewContainerRef
} from '@angular/core';
import {BrowserModule} from "@angular/platform-browser";


@Component({
  selector: 'demo-module-loader',
  template: `
    <p>
      Start editing to see some magic happen :)
    </p>
    <button (click)="loadComponent()">Load element</button>
  `,
  styleUrls: [],
  providers: []
})
export class AppComponent  {
  name = 'Angular';
  constructor(
    private readonly loader: NgModuleFactoryLoader,
    private readonly injector: Injector,
    private readonly vcr: ViewContainerRef,
  ) {}

  loadComponent() {
    this.loader.load('app/demo/module-loader/hero.module#HeroModule')
      .then(factory => {
        const moduleRef = factory.create(this.injector);
        const anyComponentType = moduleRef.injector.get('ProvidedBitkanHero');
        const componentFactory = moduleRef.componentFactoryResolver.resolveComponentFactory(anyComponentType);
        this.vcr.createComponent(componentFactory);
      });
  }
}

@NgModule({
  imports: [
    BrowserModule,
  ],
  declarations: [
    AppComponent
  ],
  bootstrap: [
    AppComponent
  ],
  providers: [
    {
      provide: NgModuleFactoryLoader,
      useClass: SystemJsNgModuleLoader
    }
  ]
})
export class DemoModuleLoaderModule {

}