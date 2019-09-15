import {
  AfterViewInit,
  Component,
  Directive,
  ElementRef,
  HostListener, Input,
  NgModule,
  OnDestroy,
  Renderer2
} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {NgxDatatableModule} from './src';
import {fromEvent} from "rxjs";
import {takeUntil} from "rxjs/operators";


@Directive({
  selector: '[resize-header-cell]'
})
export class ResizeHeaderCell implements AfterViewInit, OnDestroy {
  @Input() resizeEnabled: boolean = true;
  
  @HostListener('mousedown', ['$event'])
  onMouseDown(event: MouseEvent): void {
    const isHandle = (<HTMLElement>(event.target)).classList.contains('resize-handle');
    const initialWidth = this.element.nativeElement.clientWidth;
  
    if (isHandle) {
      event.stopPropagation(); // ???
      
      fromEvent(document, 'mousemove').pipe(
        takeUntil(fromEvent(document, 'mouseup'))
      ).subscribe((event: MouseEvent) => {
        const newWidth = event.screenX + initialWidth;
        
        console.log(newWidth, event.screenX, initialWidth);
      });
    }
  }
  
  
  constructor(private element: ElementRef, private renderer: Renderer2) {
    const div: HTMLElement = this.element.nativeElement;
    
    console.log(div.clientWidth, div.clientHeight, div.clientLeft, div.clientTop);
  }
  
  
  ngAfterViewInit(): void {
    const renderer2 = this.renderer;
    const node = renderer2.createElement('span');
    
    if (this.resizeEnabled) {
      renderer2.addClass(node, 'resize-handle');
    } else {
      renderer2.addClass(node, 'resize-handle--not-resizable');
    }
    
    renderer2.appendChild(this.element.nativeElement, node);
  }
  
  ngOnDestroy(): void {
  }
}







/**
 * Datatable:
 * datatable-header
 * datatable-body
 * datatable-footer
 */
@Component({
  selector: 'demo-datatable',
  template: `
    <div>
      <ngx-datatable
        [rows]="rows"
        [columns]="columns"
        [loadingIndicator]="true"
      >
      </ngx-datatable>
    </div>
    <div resize-header-cell style="width: 100px">
      <p>resize div</p>
    </div>
  `,
  styles: [
    `
        .resize-handle {
            cursor: ew-resize;
            display: inline-block;
            position: absolute;
            right: 0;
            top: 0;
            bottom: 0;
            width: 5px;
            padding: 0 4px;
            visibility: hidden;
        }
    `
  ]
})
export class DemoDataTableComponent {
  rows = [
    { name: 'Austin', gender: 'Male', company: 'Swimlane' },
    { name: 'Dany', gender: 'Male', company: 'KFC' },
    { name: 'Molly', gender: 'Female', company: 'Burger King' },
  ];
  columns = [
    { prop: 'name' },
    { name: 'Gender' },
    { name: 'Company' }
  ];
}



@NgModule({
  imports:[
    BrowserModule,
    NgxDatatableModule,
  ],
  declarations: [
    DemoDataTableComponent,
    ResizeHeaderCell,
  ],
  bootstrap: [
    DemoDataTableComponent,
  ]
})
export class DemoDataTableModule {
  
}