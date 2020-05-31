

import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {StateError} from '@api/frontendapi';
import {map} from 'rxjs/operators';

import {KdError} from '../common/errors/errors';

@Component({
  selector: 'kd-error',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class ErrorComponent implements OnInit {
  private error_: KdError;

  constructor(private readonly route_: ActivatedRoute) {}

  ngOnInit(): void {
    this.route_.paramMap.pipe(map(() => window.history.state)).subscribe((state: StateError) => {
      if (state.error) {
        this.error_ = state.error;
      }
    });
  }

  getErrorStatus(): string {
    if (this.error_) {
      return `${this.error_.status} (${this.error_.code})`;
    }

    return 'Unknown Error';
  }

  getErrorData(): string {
    if (this.error_) {
      return this.error_.message;
    }

    return 'No error data available.';
  }
}
