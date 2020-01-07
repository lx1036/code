import {Injectable} from '@angular/core';
import {Subject} from 'rxjs';
import {ConfirmationAcknowledgement} from './confirmation-state-message';

@Injectable()
export class ConfirmationDialogService {

  confirmationConfirmSource = new Subject<ConfirmationAcknowledgement>();

  confirmationConfirm$ = this.confirmationConfirmSource.asObservable();

  constructor() {

  }
}
