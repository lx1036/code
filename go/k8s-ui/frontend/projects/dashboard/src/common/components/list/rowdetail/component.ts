

import {Component} from '@angular/core';
import {Event} from 'typings/backendapi';

@Component({
  selector: 'kd-row-detail',
  templateUrl: 'template.html',
  styleUrls: ['style.scss'],
})
export class RowDetailComponent {
  events: Event[] = [];
}
