import {Injectable} from '@angular/core';
import {Subject} from 'rxjs';
import {ConfirmationAcknowledgement} from './confirmation-state-message';
import {ConfirmationMessage} from "./confirmation-message";

@Injectable()
export class ConfirmationDialogService {

  confirmationAnnoucedSource = new Subject<ConfirmationMessage>();
  confirmationConfirmSource = new Subject<ConfirmationAcknowledgement>();

  confirmationAnnouced$ = this.confirmationAnnoucedSource.asObservable();
  confirmationConfirm$ = this.confirmationConfirmSource.asObservable();

  // User confirm the action
  public confirm(ack: ConfirmationAcknowledgement): void {
    this.confirmationConfirmSource.next(ack);
  }

  // User cancel the action
  public cancel(ack: ConfirmationAcknowledgement): void {
    this.confirm(ack);
  }

  // Open the confirmation dialog
  public openComfirmDialog(message: ConfirmationMessage): void {
    this.confirmationAnnoucedSource.next(message);
  }
}
