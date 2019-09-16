import { Component, OnInit } from '@angular/core';
import {User} from '../../models/user';
import {createFeatureSelector, Store} from '@ngrx/store';
import {AppState, AuthState, SignUp} from '../../store';
import {filter} from 'rxjs/operators';
import {select} from '@ngrx/store/src/store';
import {Observable} from 'rxjs';

@Component({
  selector: 'app-sign-up',
  template: `
    <div class="row">
      <div class="col-md-4">
        <h1>Sign up</h1>
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
          <span>Already have an account?&nbsp;</span>
          <a [routerLink]="['/log-in']">Log in!</a>
        </p>
      </div>
    </div>
  `,
  styles: []
})
export class SignUpComponent implements OnInit {
  user: User = new User();
  authState: Observable<any>;
  errorMessage: string;


  constructor(private _store: Store<AppState>) {
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

    this._store.dispatch(new SignUp(payload));
  }
}
