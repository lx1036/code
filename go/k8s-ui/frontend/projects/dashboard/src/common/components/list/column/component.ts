

import {
  Component,
  ComponentFactoryResolver,
  ComponentRef,
  Input,
  OnChanges,
  Type,
  ViewChild,
  ViewContainerRef,
} from '@angular/core';
import {ActionColumn} from '@api/frontendapi';
import {CRD, CRDDetail, CRDObject, Resource} from 'typings/backendapi';

@Component({
  selector: 'kd-dynamic-cell',
  templateUrl: './template.html',
})
export class ColumnComponent<T extends ActionColumn> implements OnChanges {
  @Input() component: Type<T>;
  @Input() resource: Resource;
  @ViewChild('target', {read: ViewContainerRef, static: true}) target: ViewContainerRef;
  private componentRef_: ComponentRef<T> = undefined;

  constructor(private readonly resolver_: ComponentFactoryResolver) {}

  ngOnChanges(): void {
    if (this.componentRef_) {
      this.target.remove();
      this.componentRef_ = undefined;
    }

    const factory = this.resolver_.resolveComponentFactory(this.component);
    this.componentRef_ = this.target.createComponent(factory);
    this.componentRef_.instance.setObjectMeta(this.resource.objectMeta);
    this.componentRef_.instance.setTypeMeta(this.resource.typeMeta);

    if ((this.resource as CRDDetail).names !== undefined) {
      this.componentRef_.instance.setDisplayName((this.resource as CRDDetail).names.kind);
    }
  }
}
