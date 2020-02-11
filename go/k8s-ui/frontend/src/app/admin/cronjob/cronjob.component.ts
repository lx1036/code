import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {Subscription} from 'rxjs';
import {ClrDatagridStateInterface} from '@clr/angular';
import {ConfirmationDialogService} from '../../shared/confirmation-dialog.service';
import {CronjobService} from '../../shared/client/v1/cronjob.service';
import {MessageHandlerService} from '../../shared/message-handler.service';
import {ConfirmationState, ConfirmationTargets} from '../../shared/shared.const';

@Component({
  selector: 'app-cronjob',
  template: ``,
})
export class CronjobComponent implements OnInit, OnDestroy {
  appId: string;
  subscription: Subscription;
  componentName = '计划任务';

  constructor(private route: ActivatedRoute,
              private deletionDialogService: ConfirmationDialogService,
              private cronjobService: CronjobService,
              private messageHandlerService: MessageHandlerService, ) {
    this.subscription = deletionDialogService.confirmationConfirm$.subscribe(message => {
      if (message && message.state === ConfirmationState.CONFIRMED && message.source === ConfirmationTargets.CRONJOB) {
        const id = message.data;
        this.cronjobService.deleteById(id, 0).subscribe(
          response => {
            this.messageHandlerService.showSuccess(this.componentName + '删除成功！');
            this.retrieve();
          },
          error => {
            this.messageHandlerService.handleError(error);
          }
        );
      }
    });

  }

  retrieve(state?: ClrDatagridStateInterface): void {

  }


    ngOnInit() {
    this.route.params.subscribe(params => {
      this.appId = params.aid;
      if (typeof (this.appId) === 'undefined') {
        this.appId = '';
      }
    });
  }

  ngOnDestroy(): void {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
  }
}
