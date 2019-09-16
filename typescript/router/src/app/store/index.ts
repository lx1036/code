import {
  Action,
  ActionReducer,
  ActionReducerMap,
  createFeatureSelector,
  createSelector, INIT,
  MetaReducer
} from '@ngrx/store';
import { environment } from '../../environments/environment';
import {ROUTER_NAVIGATION, RouterNavigationAction, routerReducer, RouterReducerState} from '@ngrx/router-store';
import {storeFreeze} from 'ngrx-store-freeze';
import {Actions, Effect, ofType} from '@ngrx/effects';
import {Injectable} from '@angular/core';
import {catchError, map, switchMap, tap} from 'rxjs/operators';
import {User} from '../models/user';
import {AuthService} from '../services/auth.service';
import {Router} from '@angular/router';
import {Observable, of} from 'rxjs';


export interface AppState {
  authState: AuthState;
  routerState: RouterReducerState;
}

export interface AuthState {
  // is a user authenticated?
  isAuthenticated: boolean;
  // if authenticated, there should be a user object
  user: User | null;
  // error message
  errorMessage: string | null;
}

export const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  errorMessage: null
};


export enum AuthActionTypes {
  LOGIN = '[Auth] Login',
  LOGIN_SUCCESS = '[Auth] Login Success',
  LOGIN_FAILURE = '[Auth] Login Failure',
  SIGNUP = '[Auth] Signup',
  SIGNUP_SUCCESS = '[Auth] Signup Success',
  SIGNUP_FAILURE = '[Auth] Signup Failure',
  LOGOUT = '[Auth] Logout',
  GET_STATUS = '[Auth] GetStatus',
}

export class LogIn implements Action {
  readonly type = AuthActionTypes.LOGIN;
  constructor(public payload: any) {}
}
export class LogInSuccess implements Action {
  readonly type = AuthActionTypes.LOGIN_SUCCESS;
  constructor(public payload: any) {}
}
export class LogInFailure implements Action {
  readonly type = AuthActionTypes.LOGIN_FAILURE;
  constructor(public payload: any) {}
}
export class SignUp implements Action {
  readonly type = AuthActionTypes.SIGNUP;
  constructor(public payload: any) {}
}
export class SignUpSuccess implements Action {
  readonly type = AuthActionTypes.SIGNUP_SUCCESS;
  constructor(public payload: any) {}
}
export class SignUpFailure implements Action {
  readonly type = AuthActionTypes.SIGNUP_FAILURE;
  constructor(public payload: any) {}
}
export class LogOut implements Action {
  readonly type = AuthActionTypes.LOGOUT;
}
export class GetStatus implements Action {
  readonly type = AuthActionTypes.GET_STATUS;
}
export type All =
  | LogIn
  | LogInSuccess
  | LogInFailure
  | SignUp
  | SignUpSuccess
  | SignUpFailure
  | LogOut
  | GetStatus;




export const stateReducerMap: ActionReducerMap<AppState> = {
  routerState: routerReducer,
  authState: authReducer,
};


export function authReducer(state = initialState, action: All): AuthState {
  switch (action.type) {
    case AuthActionTypes.LOGIN_SUCCESS:
      return {
        ...state,
        isAuthenticated: true,
        user: {
          token: action.payload.token,
          email: action.payload.email
        },
        errorMessage: null
      };
    case AuthActionTypes.LOGIN_FAILURE:
      return {
        ...state,
        errorMessage: 'Incorrect email and/or password.'
      };
    case AuthActionTypes.SIGNUP_SUCCESS:
      return {
        ...state,
        isAuthenticated: true,
        user: {
          token: action.payload.token,
          email: action.payload.email
        },
        errorMessage: null
      };
    case AuthActionTypes.SIGNUP_FAILURE:
      return {
        ...state,
        errorMessage: 'That email is already in use.'
      };
    case AuthActionTypes.LOGOUT:
      return initialState;
    default:
      return state;
  }
}


export function debug(reducer: ActionReducer<any>): ActionReducer<any> {
  return (state: any, action: Action): any => {
    console.log('debugOne: ', state, action);

    const newState = reducer(state, action);

    console.log('debugTwo: ', state, action);

    return newState;
  };
}

export function debugTwo(reducer: ActionReducer<any>): ActionReducer<any> {
  return (state: any, action: Action): any => {
    console.log('debugThree: ', state, action);

    const newState = reducer(state, action);

    console.log('debugFour: ', state, action);

    return newState;
  };
}

// meta-reducers are similar to middleware in Redux, 前置中间件和后置中间件
export const metaReducers: MetaReducer<AppState>[] = !environment.production ? [storeFreeze, debugTwo, debug] : [];


@Injectable()
export class UserEffects {
  // Stream of all actions dispatched in your application including actions dispatched by effect streams.
  constructor(private actions$: Actions) {

  }

  @Effect({dispatch: false})
  storeInit$ = this.actions$.pipe(
    ofType(ROUTER_NAVIGATION, '@ngrx/effects/init'),
    tap(action => console.log(action)),
  )
}

@Injectable()
export class AuthEffects {

  constructor(
    private actions: Actions,
    private authService: AuthService,
    private router: Router,
  ) {}

  // effects go here

  @Effect()
  LogIn: Observable<any> = this.actions.pipe(
    ofType(AuthActionTypes.LOGIN),
    map((action: LogIn) => action.payload),
    switchMap(payload => {
      return this.authService.logIn(payload.email, payload.password).pipe(
        map((user: User) => {
          return new LogInSuccess({token: user.token, email: payload.email});
        }),
        catchError((error) => {
          return of(new LogInFailure({ error: error }));
        })
      )}
    )
  );

  @Effect()
  SignUp: Observable<any> = this.actions.pipe(
    ofType(AuthActionTypes.SIGNUP),
    map((action: SignUp) => action.payload),
    switchMap(payload => {
      return this.authService.signUp(payload.email, payload.password).pipe(
        map((user: User) => {
          return new SignUpSuccess({token: user.token, email: payload.email});
        }),
        catchError((error) => {
          return of(new SignUpFailure({ error: error }));
        })
      )}
    )
  );

  @Effect({ dispatch: false })
  AuthSuccess: Observable<any> = this.actions.pipe(
    ofType(AuthActionTypes.LOGIN_SUCCESS, AuthActionTypes.SIGNUP_SUCCESS),
    tap((action) => {
      localStorage.setItem('token', action.payload.token);
      this.router.navigateByUrl('/');
    })
  );

  @Effect({ dispatch: false })
  AuthFailure: Observable<any> = this.actions.pipe(
    ofType(AuthActionTypes.LOGIN_FAILURE, AuthActionTypes.SIGNUP_FAILURE),
    tap((action) => {
      console.log('AuthFailure', action.payload.error);
    })
  );

  @Effect({ dispatch: false })
  public LogOut: Observable<any> = this.actions.pipe(
    ofType(AuthActionTypes.LOGOUT),
    tap((action: LogOut) => {
      localStorage.removeItem('token');
    })
  );

  // 有个 bug，未登录情况下：第一次 '/status'，跳转到 '/log-in'；返回 '/'，再进入 '/status'，进入 status component。
  // 这是错的，应该还是 '/log-in'。该 effects 未再次触发。
  @Effect({ dispatch: false })
  GetStatus: Observable<any> = this.actions.pipe(
    ofType(AuthActionTypes.GET_STATUS),
    tap(action => console.log(action)),
    switchMap(action => {
      return this.authService.getStatus();
    })
  );

  @Effect({ dispatch: false })
  routerNavigation$ = this.actions.pipe(
    ofType(ROUTER_NAVIGATION),
    tap((action: RouterNavigationAction) => console.log(action.payload))
  );

  /*@Effect({ dispatch: false })
  SignUpSuccess: Observable<any> = this.actions.pipe(
    ofType(AuthActionTypes.SIGNUP_SUCCESS),
    tap((action: SignUp) => {
      localStorage.setItem('token', action.payload.token);
      this.router.navigateByUrl('/');
    })
  );

  @Effect({ dispatch: false })
  SignUpFailure: Observable<any> = this.actions.pipe(
    ofType(AuthActionTypes.SIGNUP_FAILURE),
    tap((action: SignUpFailure) => {
      console.log('SignUpFailure', action.payload.error);
    })
  );*/
}