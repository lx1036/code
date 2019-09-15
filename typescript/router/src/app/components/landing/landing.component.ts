import { Component, OnInit } from '@angular/core';
import {createFeatureSelector, Store} from '@ngrx/store';
import {AppState, AuthState, LogOut} from '../../store';
import {filter} from 'rxjs/operators';
import {select} from '@ngrx/store/src/store';
import {Observable} from 'rxjs';

@Component({
  selector: 'app-landing',
  template: `
    <div class="row">
      <div class="col-md-4">
        <h1>Angular + NGRX</h1>
        <hr><br>
        <div *ngIf="isAuthenticated; then doSomething else doSomethingElse;"></div>
        <ng-template #doSomething>
          <p>You logged in <em>{{user.email}}!</em></p>
          <button class="btn btn-primary" (click)="logOut()">Log out</button>
        </ng-template>
        <ng-template #doSomethingElse>
          <a [routerLink]="['/log-in']" class="btn btn-primary">Log in</a>
          <a [routerLink]="['/sign-up']" class="btn btn-primary">Sign up</a>
        </ng-template>
        
        <a [routerLink]="['/status']" class="btn btn-primary">Status</a>
        
        <br><br><br>

        <div class="card" style="width: 18rem;">
          <div class="card-body">
            <h5 class="card-title">Current State</h5>
            <ul>
              <li><strong>isAuthenticated</strong> - {{isAuthenticated}}</li>
              <li><strong>user.email</strong> - {{ user?.email || 'null'}}</li>
              <li><strong>user.token</strong> - {{ user?.token || 'null'}}</li>
              <li><strong>errorMessage</strong> - {{ errorMessage || 'null'}}</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: []
})
export class LandingComponent implements OnInit {
  authState: Observable<any>;
  errorMessage: string;
  isAuthenticated: boolean;
  user = null;

  constructor(private _store: Store<AppState>) {
    this.authState = _store.pipe(
      select(createFeatureSelector<AuthState>('authState')),
      // filter((state: AuthState) => !state.errorMessage)
    );
  }

  ngOnInit() {
    this.authState.subscribe((state: AuthState) => {
      this.isAuthenticated = state.isAuthenticated;
      this.user = state.user;
      this.errorMessage = state.errorMessage;
    });
  }

  logOut(): void {
    this._store.dispatch(new LogOut());
  }
}
