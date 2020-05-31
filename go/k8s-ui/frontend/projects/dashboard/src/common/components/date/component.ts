

import {Component, Input} from '@angular/core';

/**
 * Display a date
 *
 * Examples:
 *
 * Display the date:
 * <kd-date [date]="object.timestamp"></kd-date>
 *
 * Display the age of the date, and the date in a tooltip:
 * <kd-date [date]="object.timestamp" relative></kd-date>
 *
 * Display the date in the shprt format:
 * <kd-date [date]="object.timestamp" format="short"></kd-date>
 *
 * Display the age of the date, and the date in the short format in a tooltip:
 * <kd-date [date]="object.timestamp" relative format="short"></kd-date>
 *
 */
@Component({
  selector: 'kd-date',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class DateComponent {
  @Input() date: string;
  @Input() format = 'medium';

  _relative: boolean;
  @Input('relative')
  set relative(v: boolean) {
    this._relative = v !== undefined && v !== false;
  }
}
