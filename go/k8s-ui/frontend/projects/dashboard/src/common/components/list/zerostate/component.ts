

import {Component, ContentChild, Input, TemplateRef} from '@angular/core';

@Component({
  selector: 'kd-list-zero-state',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class ListZeroStateComponent {
  @ContentChild('textTemplate', {read: TemplateRef}) textTemplate: TemplateRef<any>;
}
