import {AfterContentInit, Component, ElementRef, Input, OnChanges, OnDestroy, OnInit, Renderer2, SimpleChanges} from '@angular/core';

export type ButtonType = 'primary' | 'dashed' | 'danger' | 'default';
export type ButtonShape = 'circle' | 'round' | null;

@Component({
  selector: '[ng-button]',
  template: `
<!--    <i ng-icon type="loading" *ngIf="ngLoading"></i>-->
    <span (cdkObserveContent)="checkContent()" #contentElement>
      <ng-content></ng-content>
    </span>
  `
})
export class ButtonComponent implements OnInit, OnChanges, AfterContentInit, OnDestroy {
  private element: HTMLElement;

  constructor(elementRef: ElementRef, private renderer: Renderer2) {
    this.element = elementRef.nativeElement;
    this.renderer.addClass(this.element, 'ant-btn');
  }

  @Input() type: ButtonType = 'default';
  @Input() shape: ButtonShape = null;
  @Input() size: string = 'default';

  ngOnInit(): void {
    const classes = {
      [`ant-btn-${this.type}`]: this.type,
    };

    for (const index in classes) {
      this.renderer.addClass(this.element, index);
    }
  }


  ngOnChanges(changes: SimpleChanges): void {
  }

  ngAfterContentInit(): void {
  }



  ngOnDestroy(): void {
  }


  /**
   * @see https://github.com/angular/angular/issues/7289
   */
  private setClassMap() {

  }
}

@Component({
  selector: 'ng-button-group',
})
export class ButtonGroupComponent {

}
