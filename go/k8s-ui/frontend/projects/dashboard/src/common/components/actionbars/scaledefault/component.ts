

import {Component, OnDestroy, OnInit} from '@angular/core';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {Subscription} from 'rxjs/Subscription';
import {ActionbarService, ResourceMeta} from '../../../services/global/actionbar';

@Component({
  selector: '',
  templateUrl: './template.html',
})
export class ScaleDefaultActionbar implements OnInit, OnDestroy {
  isInitialized = false;
  isVisible = false;
  resourceMeta: ResourceMeta;

  private _unsubscribe = new Subject<void>();

  constructor(private readonly actionbar_: ActionbarService) {}

  ngOnInit(): void {
    this.actionbar_.onInit
      .pipe(takeUntil(this._unsubscribe))
      .subscribe((resourceMeta: ResourceMeta) => {
        this.resourceMeta = resourceMeta;
        this.isInitialized = true;
        this.isVisible = true;
      });

    this.actionbar_.onDetailsLeave
      .pipe(takeUntil(this._unsubscribe))
      .subscribe(() => (this.isVisible = false));
  }

  ngOnDestroy(): void {
    this._unsubscribe.next();
    this._unsubscribe.complete();
  }

  scalable(): boolean {
    return this.resourceMeta.typeMeta.scalable;
  }
}
