import {AfterContentChecked, Component, ContentChildren, Directive, ElementRef, NgModule, QueryList} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';


@Directive({
  selector: '[child-directive]'
})
export class ChildDirective {
  constructor(public elementRef: ElementRef) {}

  ngOnInit() {
    console.log(this.elementRef);
  }
}

@Component({
  selector: 'parent',
  template: `
    <ng-content></ng-content>
  `
})
export class Parent implements AfterContentChecked {
  @ContentChildren(ChildDirective, {descendants: true}) children: QueryList<ChildDirective>;

  ngAfterContentChecked() {
    console.log('lx', this.children.first);
  }
}

@Component({
  selector: 'child',
  template: `<div child-directive>child</div>`
})
export class Child {

}

@Component({
  selector: 'demo-test-content-projection',
  template: `
    <parent>
      <child></child>
      <child></child>
      <child></child>
    </parent>
    <div>
      demo test descended content projection
    </div>
  `
})
export class Demo {

}


@NgModule({
  imports: [
    BrowserModule,
  ],
  declarations: [
    Parent,
    Child,
    ChildDirective,
    Demo,
  ],
  bootstrap: [
    Demo,
  ]
})
export class DemoTestContentProjection {

}