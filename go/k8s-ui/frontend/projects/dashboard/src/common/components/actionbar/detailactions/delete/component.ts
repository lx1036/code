

import {Component, Input, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {ObjectMeta, TypeMeta} from '@api/backendapi';
import {first} from 'rxjs/operators';

import {VerberService} from '../../../../services/global/verber';

@Component({
  selector: 'kd-actionbar-detail-delete',
  templateUrl: './template.html',
})
export class ActionbarDetailDeleteComponent implements OnInit {
  @Input() objectMeta: ObjectMeta;
  @Input() typeMeta: TypeMeta;
  @Input() displayName: string;

  constructor(
    private readonly verber_: VerberService,
    private readonly route_: ActivatedRoute,
    private readonly router_: Router,
  ) {}

  ngOnInit(): void {
    this.verber_.onDelete.pipe(first()).subscribe(() => {
      this.router_.navigate(['.'], {relativeTo: this.route_, queryParamsHandling: 'preserve'});
    });
  }

  onClick(): void {
    this.verber_.showDeleteDialog(this.displayName, this.typeMeta, this.objectMeta);
  }
}
