

import {Component, Input} from '@angular/core';
import {ObjectMeta, TypeMeta} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {VerberService} from '../../../../services/global/verber';

@Component({
  selector: 'kd-actionbar-detail-trigger',
  templateUrl: './template.html',
})
export class ActionbarDetailTriggerComponent {
  @Input() objectMeta: ObjectMeta;
  @Input() typeMeta: TypeMeta;
  @Input() displayName: string;

  constructor(private readonly verber_: VerberService) {}

  onClick(): void {
    this.verber_.showTriggerDialog(this.displayName, this.typeMeta, this.objectMeta);
  }
}
