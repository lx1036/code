

import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';

import {NAMESPACE_STATE_PARAM} from '../../../../../common/params/params';
import {ActionbarService, ResourceMeta} from '../../../../../common/services/global/actionbar';

@Component({
  selector: '',
  templateUrl: './template.html',
})
export class ActionbarComponent implements OnInit {
  isInitialized = false;
  isVisible = false;
  resourceMeta: ResourceMeta;

  private _unsubscribe = new Subject<void>();

  constructor(private readonly actionbar_: ActionbarService, private readonly router_: Router) {}

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

  onClick(): void {
    this.router_.navigate(['overview'], {
      queryParamsHandling: 'merge',
      queryParams: {[NAMESPACE_STATE_PARAM]: this.resourceMeta.objectMeta.name},
    });
  }
}
