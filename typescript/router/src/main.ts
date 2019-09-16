import {COMPILER_OPTIONS, CompilerFactory, createPlatformFactory, enableProdMode, NgModuleRef, PLATFORM_INITIALIZER} from '@angular/core';

import { environment } from './environments/environment';
import {platformBrowserDynamic} from '@angular/platform-browser-dynamic';
import {DemoModuleLoaderModule} from "./app/demo/module-loader/module-loader";
import {Injector, StaticProvider} from "./app/packages/angular/core/src/di";
import {PLATFORM_ID} from "./app/packages/angular/core/src/application_tokens";
import {PlatformRef} from "./app/packages/angular/core/src/application_ref";
import {TestabilityRegistry} from "./app/packages/angular/core/src/testability/testability";
import {Console} from "./app/packages/angular/core/src/console";
import {DemoTestContentProjection} from './app/demo/content-projection/content-projection';
import {enableDebugTools} from '@angular/platform-browser';
import {DemoDragDrop} from "./app/demo/drag-drop/drag-drop";
import {DemoView} from "./app/demo/view/view_ref";
import {DemoRouterModule} from './app/demo/router/router';
import {DemoZoneModule} from "./app/demo/zone/zone";
import {DemoDataTableModule} from './app/demo/datatable/datatable';
import {AffixModuleDemo} from './app/demo/affix/affix';
import {BreadcrumbModule} from './app/ant-design/breadcrumb/breadcrumb';

if (environment.production) {
  enableProdMode();
}

// enableDebugTools();

const platform = platformBrowserDynamic();

// platform.bootstrapModule(OverlayModule)
// platform.bootstrapModule(DemoHttpModule)
// platform.bootstrapModule(TestCustomHttpClientModule)
// platform.bootstrapModule(DemoModuleLoaderModule)
// platform.bootstrapModule(DemoFormsModule)
// platform.bootstrapModule(DemoZoneModule)
// platform.bootstrapModule(DemoDataTableModule)
// platform.bootstrapModule(FormValidationModule)
// platform.bootstrapModule(DemoTestContentProjection)
// platform.bootstrapModule(DemoDragDrop)
// platform.bootstrapModule(DemoView)
// platform.bootstrapModule(DemoRouterModule)
// platform.bootstrapModule(AffixModuleDemo)
platform.bootstrapModule(BreadcrumbModule)
  .catch(err => console.log(err));


/*platform.bootstrapModule(App2Module)
.catch(err => console.log(err));*/
/**
 * https://blog.kevinyang.net/2017/09/24/angular-injector/
 * https://blog.angularindepth.com/angular-dependency-injection-and-tree-shakeable-tokens-4588a8f70d5d
 */
/*export const platformCore: ((extraProviders?: StaticProvider[]) => PlatformRef) = createPlatformFactory(null, 'core', [
  // Set a default platform name for platforms that don't set it explicitly.
  {provide: PLATFORM_ID, useValue: 'unknown'},
  {provide: PlatformRef, deps: [Injector]},
  {provide: TestabilityRegistry, deps: []},
  {provide: Console, deps: []},
]);

export const platformCoreDynamic: ((extraProviders?: StaticProvider[]) => PlatformRef) = createPlatformFactory(platformCore, 'coreDynamic', [
  {provide: COMPILER_OPTIONS, useValue: {}, multi: true},
  {provide: CompilerFactory, useClass: JitCompilerFactory, deps: [COMPILER_OPTIONS]},
]);


const platformBrowserDynamic: ((extraProviders?: StaticProvider[]) => PlatformRef) = createPlatformFactory(platformCoreDynamic, 'browserDynamic', [
  [
    {provide: PLATFORM_ID, useValue: PLATFORM_BROWSER_ID},
    {provide: PLATFORM_INITIALIZER, useValue: initDomAdapter, multi: true},
    {provide: PlatformLocation, useClass: BrowserPlatformLocation, deps: [DOCUMENT]},
    {provide: DOCUMENT, useFactory: _document, deps: []},
  ],
  {
    provide: COMPILER_OPTIONS,
    useValue: {providers: [{provide: ResourceLoader, useClass: ResourceLoaderImpl, deps: []}]},
    multi: true
  },
  {provide: PLATFORM_ID, useValue: PLATFORM_BROWSER_ID},
]);

const platformRef: PlatformRef = platformBrowserDynamic();

platformRef.bootstrapModule(AppModule).then((appModuleRef: NgModuleRef<AppModule>) => {
  console.log(appModuleRef.instance);
});*/
// const _CORE_PLATFORM_PROVIDERS: StaticProvider[] = [
//   // Set a default platform name for platforms that don't set it explicitly.
//   {provide: PLATFORM_ID, useValue: 'unknown'},
//   {provide: PlatformRef, deps: [Injector]},
//   {provide: TestabilityRegistry, deps: []},
//   {provide: Console, deps: []},
// ];
// export const platformCore = createPlatformFactory(null, 'core', _CORE_PLATFORM_PROVIDERS);
//
// platformCore();
