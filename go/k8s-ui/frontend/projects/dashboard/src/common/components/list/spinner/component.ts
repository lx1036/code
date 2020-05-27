

import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: '[kdLoadingSpinner]',
  templateUrl: './template.html',
  host: {
    // kd-loading-share class is defined globally in index.scss file.
    '[class.kd-loading-shade]': 'isLoading',
  },
})
export class LoadingSpinner implements OnInit {
  @Input() isLoading: boolean;

  ngOnInit(): void {
    if (this.isLoading === undefined) {
      throw Error('isLoading is a required property of loading spinner.');
    }
  }
}
