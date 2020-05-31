

import {Component, Input} from '@angular/core';
import {ObjectMeta, TypeMeta} from '@api/backendapi';
import {ResourceMeta} from '../../../services/global/actionbar';

@Component({
  selector: 'kd-actionbar-detail-actions',
  templateUrl: './template.html',
})
export class ActionbarDetailActionsComponent {
  @Input() objectMeta: ObjectMeta;
  @Input() typeMeta: TypeMeta;
  @Input() displayName: string;
}
