import { Component, OnInit } from '@angular/core';
import {Store} from '@ngrx/store';
import {AppState, GetStatus} from '../../store';

@Component({
  selector: 'app-status',
  template: `
    <div class="row">
      <div class="col-md-4">
        <h1>Status Works!</h1>
        <hr><br>
        <a [routerLink]="['/']" class="btn btn-primary">Home</a>
      </div>
    </div>
  `,
  styles: []
})
export class StatusComponent implements OnInit {

  constructor(private _store: Store<AppState>) { }

  ngOnInit() {
    console.log('status');
    this._store.dispatch(new GetStatus());
  }
}
