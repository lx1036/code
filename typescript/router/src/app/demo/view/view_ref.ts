import {
  AfterContentInit,
  AfterViewInit,
  Component,
  Directive,
  ElementRef,
  EmbeddedViewRef,
  NgModule,
  TemplateRef,
  ViewChild, ViewContainerRef
} from "@angular/core";
import {BrowserModule} from "@angular/platform-browser";

/**
 * [译] 探索 Angular 使用 ViewContainerRef 操作 DOM: https://juejin.im/post/5ab09a49518825557005d805
 */

@Directive({
  selector: '[element-ref]'
})
export class ElementRefDir {
  constructor(private elementRef: ElementRef) {
    console.log(this.elementRef);
  }
}

@Directive({
  selector: 'ng-template[template-ref]'
})
export class TemplateRefDir {
  embeddedView: EmbeddedViewRef<{name: string}>;
  
  constructor(private templateRef: TemplateRef<any>) {
    this.embeddedView = templateRef.createEmbeddedView({name: 'lx1036'});
    
    console.log(this.templateRef.elementRef.nativeElement);
  }
  
}

@Component({
  selector: 'demo-template-ref',
  template: `
    <div class="parent" element-ref>
      <p class="child">TemplateRef</p>
    </div>
    <ng-container #vc></ng-container>
    <ng-template template-ref let-name>
      <p class="child">TemplateRef2 {{name}}</p>
    </ng-template>
  `
})
export class DemoTemplateRef implements AfterContentInit {
  @ViewChild(TemplateRefDir) templateDir;
  @ViewChild('vc', {read: ViewContainerRef}) viewContainer: ViewContainerRef;
  
  ngAfterContentInit() {
    this.viewContainer.insert(this.templateDir.embeddedView);
  }
  
  ngAfterViewInit() {
    // 触发错误 ExpressionChangedAfterItHasBeenCheckedError: Expression has changed after it was checked.
    // this.viewContainer.insert(this.templateDir.embeddedView);
  }
}


@NgModule({
  imports: [BrowserModule],
  declarations: [DemoTemplateRef,TemplateRefDir,ElementRefDir],
  bootstrap: [DemoTemplateRef]
})
export class DemoView {

}
