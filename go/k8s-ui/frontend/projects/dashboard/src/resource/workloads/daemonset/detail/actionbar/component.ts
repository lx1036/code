

import {Component, OnInit} from '@angular/core';
import {Subscription} from 'rxjs/Subscription';
import {ActionbarService, ResourceMeta} from '../../../../../common/services/global/actionbar';

@Component({
  selector: '',
  templateUrl: './template.html',
})
export class ActionbarComponent implements OnInit {
  isInitialized = false;
  resourceMeta: ResourceMeta;
  resourceMetaSubscription_: Subscription;

  constructor(private readonly actionbar_: ActionbarService) {}

  ngOnInit(): void {
    this.resourceMetaSubscription_ = this.actionbar_.onInit.subscribe(
      (resourceMeta: ResourceMeta) => {
        this.resourceMeta = resourceMeta;
        this.isInitialized = true;
      },
    );
  }

  ngOnDestroy(): void {
    this.resourceMetaSubscription_.unsubscribe();
  }
}
