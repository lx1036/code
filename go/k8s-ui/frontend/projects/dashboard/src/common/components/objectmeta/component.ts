

import {Component, Input} from '@angular/core';
import {ObjectMeta} from '@api/backendapi';

@Component({
  selector: 'kd-object-meta',
  templateUrl: './template.html',
})
export class ObjectMetaComponent {
  @Input() initialized = false;

  private objectMeta_: ObjectMeta;
  get objectMeta(): ObjectMeta {
    return this.objectMeta_;
  }

  @Input()
  set objectMeta(val: ObjectMeta) {
    if (val === undefined) {
      this.objectMeta_ = {} as ObjectMeta;
    } else {
      this.objectMeta_ = val;
    }
  }
}
