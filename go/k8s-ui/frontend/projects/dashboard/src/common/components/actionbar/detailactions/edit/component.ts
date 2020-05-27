

import {Component, Input} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {ObjectMeta, TypeMeta} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {VerberService} from '../../../../services/global/verber';

@Component({
  selector: 'kd-actionbar-detail-edit',
  templateUrl: './template.html',
})
export class ActionbarDetailEditComponent {
  @Input() objectMeta: ObjectMeta;
  @Input() typeMeta: TypeMeta;
  @Input() displayName: string;

  constructor(private readonly verber_: VerberService) {}

  onClick(): void {
    this.verber_.showEditDialog(this.displayName, this.typeMeta, this.objectMeta);
  }
}
