import {AfterViewInit, Component, Directive, ElementRef, Inject, NgModule, NgZone, OnDestroy} from "@angular/core";
import {BrowserModule} from "@angular/platform-browser";
import {DragDropModule} from "@angular/cdk/drag-drop";
import {DOCUMENT} from "@angular/common";
import {take} from "rxjs/operators";



@Directive({
  selector: '[customCdkDrag]',
  host: {
    'class': 'cdk-drag',
    '[class.cdk-drag-dragging]': '_isDragging()'
  }
})
export class CustomCdkDrag implements AfterViewInit, OnDestroy {
  constructor(
    @Inject(DOCUMENT) public document: Document,
    public zone: NgZone,
    public element: ElementRef<HTMLElement>
    ) {
    
  }
  
  ngAfterViewInit(): void {
    this.zone.onStable.asObservable().pipe(take(1)).subscribe(() => {
      this.zone.run(() => {
        this.element.nativeElement.addEventListener('mousedown', (event: MouseEvent) => {
          console.log(event);
        });
      });
    });
  }
  
  ngOnDestroy(): void {
  }
  
  
  _isDragging(): boolean {
    return false;
  }
}


@Component({
  selector: 'demo-drag-drop',
  template: `
    <div class="box" cdkDrag>
      I can be dragged.
    </div>
    
    <div class="box" customCdkDrag>
      I can be dragged too.
    </div>
  `,
  styles: [
    `
        .box {
            width: 500px;
            height: 500px;
            border: 2px solid red;
            border-radius: 20px;
            display: inline-flex;
            justify-content: center;
            text-align: center;
            align-items: center;
        }
    `
  ]
})
export class Demo {

}


@NgModule({
  imports: [
    BrowserModule,
    DragDropModule,
  ],
  declarations: [
    Demo,
    CustomCdkDrag,
  ],
  bootstrap: [
    Demo,
  ]
})
export class DemoDragDrop {

}


