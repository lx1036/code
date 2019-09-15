import {BrowserModule} from '@angular/platform-browser';
import {Component, NgModule, ApplicationRef} from '@angular/core';

@Component({
  selector: 'test',
  template: `
    <p>test</p>
  `
})
export class MyComponent {}

let c = {cc: 'cc'};
let d = c;
c.cc = 'xxx';
console.log(d,c);
@NgModule({
  imports: [BrowserModule],
  declarations: [MyComponent],
  entryComponents: [MyComponent]
})
export class MainModule {
  constructor(appRef: ApplicationRef) {
    appRef.bootstrap(MyComponent);
  }
}


// Intersection Types(https://www.tslang.cn/docs/handbook/advanced-types.html)
let option = {providedIn: 'test'};
type person = {providedIn: string};

let testAdvancedTypes: person & {useValue: any} = {useValue: 1, providedIn: 'a'};
