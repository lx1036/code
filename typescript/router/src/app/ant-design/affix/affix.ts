import {
  ChangeDetectionStrategy,
  Component,
  ElementRef,
  Inject,
  Injectable,
  Input,
  OnDestroy,
  OnInit
} from "@angular/core";
import {DOCUMENT} from "@angular/common";
import {Platform} from "@angular/cdk/platform";


@Injectable()
export class NgScrollService {
  
  getScroll(_target: Window | Element, b: boolean) {
    
  }
}

@Component({
  selector: 'ng-affix',
  template: `
    <div #fixedEl>
      <ng-content></ng-content>
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush,
  styles: [
    `
    
    `
  ],
})
export class AffixComponent implements OnInit, OnDestroy {
  private readonly events = ['resize', 'scroll', 'touchstart', 'touchmove', 'touchend', 'pageshow', 'load'];
  private _target: Window | Element | null = null;
  private readonly element: HTMLElement;
  
  private timeout: number;
  
  private _offsetTop: number;
  @Input()
  set offsetTop(value: number|null) {
    if (!value) {
      return;
    }
    
    this._offsetTop = value;
  }
  
  get offsetTop(): number {
    return this._offsetTop;
  }
  
  constructor(el: ElementRef, private scrollService: NgScrollService, @Inject(DOCUMENT) private document: any, private platform: Platform) {
    this.element = el.nativeElement;
    
    if (this.platform.isBrowser) {
      this._target = window;
    }
  }
  
  public ngOnInit(): void {
    this.timeout = setTimeout(() => {
      this.setTargetEventListeners();
    });
  }
  
  public ngOnDestroy(): void {
    this.clearTargetEventListeners();
    clearTimeout(this.timeout);
  }
  
  private setTargetEventListeners(): void {
    this.clearTargetEventListeners();
    
    if (this.platform.isBrowser) {
      this.events.forEach((event: string) => {
        this._target!.addEventListener(event, this.updatePosition, false);
      });
    }
  }
  
  private clearTargetEventListeners() {
    if (this.platform.isBrowser) {
      this.events.forEach((event: string) => {
        this._target!.removeEventListener(event, this.updatePosition, false);
      });
    }
  }
  
  private updatePosition(event: Event) {
    const scrollTop = this.scrollService.getScroll(this._target, true);
    const elementOffset = this.getOffset(this.element, this._target);
    const targetRect = this.getTargetRect(this._target);
    const offsetTop = this.offsetTop;
    
    if (scrollTop >= (elementOffset.top - offsetTop) /*&& offsetMode.top*/) {
      this.setAffixStyle(event, {
        position: 'fixed',
      });
    }
  }
  
  private getOffset(element: HTMLElement, _target: Window | Element): {top: number, left: number, width: number, height: number} {
    const rect = element.getBoundingClientRect();
    const width = rect.width;
    const height = rect.height;
    const top = rect.top
    
  }
  
  private getTargetRect(target: Window | Element): ClientRect {
    return target !== window ? (target as HTMLElement).getBoundingClientRect() : ({top: 0, bottom: 0, left: 0, right: 0} as ClientRect);
  }
  
  private setAffixStyle(event: Event, param2: {}) {
  
  }
}
