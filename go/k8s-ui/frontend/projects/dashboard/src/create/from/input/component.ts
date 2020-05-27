

import {Component} from '@angular/core';

import {CreateService} from '../../../common/services/create/service';
import {HistoryService} from '../../../common/services/global/history';
import {NamespaceService} from '../../../common/services/global/namespace';

@Component({
  selector: 'kd-create-from-input',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class CreateFromInputComponent {
  inputData = '';

  constructor(
    private readonly namespace_: NamespaceService,
    private readonly create_: CreateService,
    private readonly history_: HistoryService,
  ) {}

  isCreateDisabled(): boolean {
    return !this.inputData || this.inputData.length === 0 || this.create_.isDeployDisabled();
  }

  create(): void {
    this.create_.createContent(this.inputData);
  }

  cancel(): void {
    this.history_.goToPreviousState('overview');
  }

  areMultipleNamespacesSelected(): boolean {
    return this.namespace_.areMultipleNamespacesSelected();
  }
}
