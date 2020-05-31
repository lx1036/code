

import {Component, Input} from '@angular/core';
import {ObjectMeta, TypeMeta} from '@api/backendapi';
import {PinnerService} from '../../../../services/global/pinner';

@Component({
  selector: 'kd-actionbar-detail-pin',
  templateUrl: './template.html',
})
export class ActionbarDetailPinComponent {
  @Input() objectMeta: ObjectMeta;
  @Input() typeMeta: TypeMeta;
  @Input() displayName: string;

  constructor(private readonly pinner_: PinnerService) {}

  onClick(): void {
    if (this.isPinned()) {
      this.pinner_.unpin(this.typeMeta.kind, this.objectMeta.name, this.objectMeta.namespace);
    } else {
      this.pinner_.pin(
        this.typeMeta.kind,
        this.objectMeta.name,
        this.objectMeta.namespace,
        this.displayName,
      );
    }
  }

  isPinned(): boolean {
    return this.pinner_.isPinned(
      this.typeMeta.kind,
      this.objectMeta.name,
      this.objectMeta.namespace,
    );
  }
}
