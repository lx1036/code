import { Component, OnInit } from '@angular/core';
import {User} from '../../models/user';
import {createFeatureSelector, Store} from '@ngrx/store';
import {AppState, AuthState, LogIn} from '../../store';
import {Observable} from 'rxjs';
import {filter, map} from 'rxjs/operators';
import {select} from '@ngrx/store/src/store';
import {state} from '@angular/animations';

@Component({
  selector: 'app-login',
  template: `
    <div class="row">
      <div class="col-md-4">
        <h1>Log in</h1>
        <hr><br>
        <div>
          <div class="alert alert-danger" role="alert">
            {{errorMessage}}
          </div>
        </div>
        <form (ngSubmit)="onSubmit()" novalidate>
          <div class="form-group">
            <label for="email">Email</label>
            <input
              [(ngModel)]="user.email"
              name="email"
              type="email"
              required
              class="form-control"
              id="email"
              placeholder="enter your email">
          </div>
          <div class="form-group">
            <label for="password">Password</label>
            <input
              [(ngModel)]="user.password"
              name="password"
              type="password"
              required
              class="form-control"
              id="password"
              placeholder="enter a password">
          </div>
          <button type="submit" class="btn btn-primary">Submit</button>
          <a [routerLink]="['/']" class="btn btn-success">Cancel</a>
        </form>
        <p>
          <span>Don't have an account?&nbsp;</span>
          <a [routerLink]="['/sign-up']">Sign up!</a>
        </p>
      </div>
    </div>
  `,
  styles: []
})
export class LoginComponent implements OnInit {
  user: User = new User();
  authState: Observable<any>;
  errorMessage: string;

  constructor(private _store: Store<AppState>) {
    /*this.authState = _store.select(createFeatureSelector<AuthState>('authState')).pipe(
      filter(state => !state.errorMessage)
    );*/

    this.authState = _store.pipe(
      select(createFeatureSelector<AuthState>('authState')),
      filter((state: AuthState) => !!state.errorMessage)
    );
  }

  ngOnInit() {
    this.authState.subscribe((state: AuthState) => {
      this.errorMessage = state.errorMessage;
    });
  }
  onSubmit(): void {
    // console.log(this.user);

    const payload = {
      email: this.user.email,
      password: this.user.password
    };
    this._store.dispatch(new LogIn(payload));

  }
}
