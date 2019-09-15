import {Component} from '@angular/core';
import {ModalService} from './modal';


@Component({
  selector: 'ng-modal-basic',
  template: `
    <button (click)="showModal()"><span>Show Modal</span></button>
    <ng-modal [isVisible]="isVisible" title="modal title" (cancel)="handleCancel()" (ok)="handleOK()">
      <p>Content 1</p>
      <p>Content 2</p>
      <p>Content 3</p>
    </ng-modal>
  `
})
export class BasicModalComponent {
  isVisible = false;

  showModal() {
    this.isVisible = true;
  }

  handleCancel() {
    this.isVisible = false;
  }

  handleOK() {
    this.isVisible = false;
  }
}

@Component({
  selector: 'ng-modal-async',
  template: `
    <button (click)="showModal()"><span>Show Modal</span></button>
    <ng-modal 
      [isVisible]="isVisible" 
      title="modal title" 
      (cancel)="handleCancel()" 
      (ok)="handleOK()"
      [okLoading]="isOkLoading">
      <p>Content 1</p>
      <p>Content 2</p>
      <p>Content 3</p>
    </ng-modal>
  `
})
export class AsyncModalComponent {
  isVisible = false;
  isOkLoading = false;

  showModal() {
    this.isVisible = true;
  }

  handleCancel() {
    this.isVisible = false;
  }

  handleOK() {
    this.isOkLoading = true;

    setTimeout(() => {
      this.isVisible = false;
      this.isOkLoading = false;
    }, 3000);
  }
}

@Component({
  selector: 'ng-modal-confirm-promise',
  template: `
    <button (click)="showConfirm()"><span>Confirm</span></button>
  `
})
export class ConfirmPromiseModalComponent {
  constructor(private modal: ModalService) {}

  showConfirm() {
    this.modal.confirm({
      title: 'Delete the items?',
      content: 'When clicked the OK button, this dialog will be closed after 1 second',
      onOk: () => {
        return new Promise((resolve, reject) => {
          setTimeout(Math.random() > 0.5 ? resolve : reject, 1000);
        }).catch(() => console.log('Oops errors!'));
      }
    });
  }
}
