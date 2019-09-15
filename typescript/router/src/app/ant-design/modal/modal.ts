import {
  AfterViewInit,
  Component, ComponentFactoryResolver, ComponentRef, ElementRef,
  EventEmitter, Inject, Injectable, Injector,
  Input,
  OnChanges,
  OnDestroy,
  OnInit,
  Output,
  SimpleChanges,
  TemplateRef, Type, ViewContainerRef
} from '@angular/core';
import {BlockScrollStrategy, Overlay, OverlayRef} from '@angular/cdk/overlay';
import {InputBoolean} from '../core/decorator';
import {fromEvent, Subject} from 'rxjs';
import {DOCUMENT} from '@angular/common';
import {takeUntil} from 'rxjs/operators';
import {ESCAPE} from '@angular/cdk/keycodes';


@Injectable()
export class ModalService {

  confirm(param: { title: string; content: string; onOk: () => Promise<any | void> }) {
    
  }
}

@Injectable()
export class ModalControlService {
  registerModal(modal: ModalRef) {

  }
}

export abstract class ModalRef {

}

@Component({
  selector: 'ng-modal',
  template: `
    <div>
      <div class="ant-modal-mask"></div>
      <div>
        <div>
          <button></button>
        </div>
      </div>
    </div>
    
    <ng-template #tplOriginContent>
    
    </ng-template>
    
    <ng-template #tplContentDefault>
    
    </ng-template>
    
    <ng-template #tplContentConfirm>
    
    </ng-template>
  `
})
export class ModalComponent<T = any> extends ModalRef implements OnInit, OnChanges, AfterViewInit, OnDestroy {
  @Input() @InputBoolean() isVisible = false;
  @Input() @InputBoolean() keyboard = true;
  @Input() title: string | TemplateRef<{}>;
  @Input() content: string | TemplateRef<{}> | Type<T>;
  @Input() footer: string | TemplateRef<{}>;
  @Input() container: HTMLElement | OverlayRef | (() => OverlayRef | HTMLElement)  = () => this.overlay.create();

  @Output() ok = new EventEmitter<T>();
  @Output() cancel = new EventEmitter<T>();

  private scrollStrategy: BlockScrollStrategy;
  private unsubscribe$ = new Subject<void>();
  private contentComponentRef: ComponentRef<T>;

  constructor(private overlay: Overlay,
              @Inject(DOCUMENT) private _document: any,
              private componentFactoryResolver: ComponentFactoryResolver,
              private viewContainer: ViewContainerRef,
              private elementRef: ElementRef,
              private modalControl: ModalControlService,) {
    super();
    this.scrollStrategy = this.overlay.scrollStrategies.block();
  }

  ngOnInit(): void {
    this.registerKeyboardEventListener();
    this.createIfContentComponent();
    this.appendModalIntoDom();

    this.modalControl.registerModal(this);
  }

  private registerKeyboardEventListener() {
    fromEvent(this._document.body, 'keydown')
      .pipe(takeUntil(this.unsubscribe$))
      .subscribe((event: KeyboardEvent) => {
        if (this.keyboard && event.keyCode === ESCAPE) {
          this.onClickOkCancel('cancel');
        }
      });
  }

  private createIfContentComponent() {
    if (this.content instanceof Type) {
      this.createDynamicComponent(this.content);
    }
  }

  private appendModalIntoDom() {
    const container = typeof this.container === 'function' ? this.container() : this.container;
    if (container instanceof HTMLElement) {
      container.appendChild(this.elementRef.nativeElement);
    } else if (container instanceof OverlayRef) {

    }
  }

  ngOnChanges(changes: SimpleChanges): void {

  }

  ngAfterViewInit(): void {

  }



  ngOnDestroy(): void {

  }

  private onClickOkCancel(type: 'ok' | 'cancel') {
    const trigger = {ok: this.ok, cancel: this.cancel}[type];

    if (trigger instanceof EventEmitter) {
      trigger.emit(this.getContentComponent());
    }
  }

  private getContentComponent(): T {
    return this.contentComponentRef && this.contentComponentRef.instance;
  }

  @Input() componentParams: T;

  /**
   * Only create a component dynamically, not attach to any View.
   *
   * @param component
   */
  private createDynamicComponent(component: Type<T>) {
    const componentFactory = this.componentFactoryResolver.resolveComponentFactory(component);
    const childInjector = Injector.create({providers: [{provide: ModalRef, useValue: this}], parent: this.viewContainer.injector});
    this.contentComponentRef = componentFactory.create(childInjector);

    if (this.componentParams) {

    }

    this.contentComponentRef.changeDetectorRef.detectChanges();
  }
}
