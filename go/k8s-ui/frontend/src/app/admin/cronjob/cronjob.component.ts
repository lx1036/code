import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {Subscription} from 'rxjs';
import {ConfirmationDialogService} from '../../shared/confirmation-dialog/confirmation-dialog.service';
import {ConfirmationState, ConfirmationTargets} from '../../shared/shared.const';

@Component({
  selector: 'wayne-cronjob',
  templateUrl: './cronjob.component.html',
  styleUrls: ['./cronjob.component.scss']
})
export class CronjobComponent implements OnInit, OnDestroy {

  appId: string;
  subscription: Subscription;

  constructor(private route: ActivatedRoute,
              private deletionDialogService: ConfirmationDialogService,
              private cronjobService: CronjobService, ) {
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
