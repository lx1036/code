/**
 * @link https://zhuanlan.zhihu.com/p/45121775
 * @link https://github.com/SebastianM/tinystate
 */
import {Component, Inject, Injectable, InjectionToken, ModuleWithProviders, NgModule, Optional} from "@angular/core";
import {BehaviorSubject, Observable, Observer} from "rxjs";
import {distinctUntilChanged, map, observeOn} from "rxjs/operators";


export const STATE_PLUGINS = new InjectionToken<Plugin>('STATE_PLUGINS');

interface StatePlugin {
  sendState: (state) => void;
}

export class RootStore {
  private _plugins: StatePlugin[];
  
  constructor(@Optional()@Inject(STATE_PLUGINS) plugins: null | StatePlugin[], ) {
    this._plugins = Array.isArray(plugins) ? plugins : [];
  }
  
}


@Injectable({providedIn: NgxModule})
export abstract class Store<T extends object> implements Observer{
  closed: boolean;
  complete: () => void;
  error: (err: any) => void;
  next: (value: T) => void;
  
  
  private _state$ = new BehaviorSubject<T>(<T>Object.assign({}, this.getInitialState()));
  
  select<S>(selectFn: (state) => S): Observable<S> {
    return this._state$.pipe(
      map(state => selectFn(state)),
      distinctUntilChanged(),
      // observeOn(async),
    );
  }
  
  protected setState(stateFn: (state) => T): void {
    this._state$.next(stateFn(this._state$.value));
  }
  
  protected abstract getInitialState(): T;
}


@NgModule()
export class NgxModule {
  static forRoot(): ModuleWithProviders {
    return {
      ngModule: NgxModule,
      providers: [
        RootStore,
      ]
    };
  }
}

export class ReduxDevtoolsPlugin implements StatePlugin{
  private _devTools;
  
  constructor(private _window: Window) {
    const devTools = this._window['__REDUX_DEVTOOLS_EXTENSION__'] || this._window['devToolsExtension'];
  
    if (!devTools) return;
    
    this._devTools = devTools.connect({name: 'CustomState v1.0.0'})
  }
  
  sendState(state: object): void {
    this._devTools.send('NO_NAME', state);
  }
}

@NgModule()
export class ReduxDevtoolsPluginModule {
  static forRoot(): ModuleWithProviders {
    return {
      ngModule: ReduxDevtoolsPluginModule,
      providers: [
        ReduxDevtoolsPlugin,
        {
          provide: STATE_PLUGINS,
          useClass: ReduxDevtoolsPlugin,
          multi: true,
        }
      ]
    }
  }
}

/**
 * ****************************** Demo **********************************
 */
export interface CounterState {
  count: number;
}

/**
 * A Container is a very simple class that holds your state and some logic for updating it.
 * The shape of the state is described via an interface (in this example: CounterState).
 */
export class CounterStore extends Store<CounterState> {
  getInitialState(): CounterState {
    return {count: 0};
  }
  
  increment(increment: number = 1) {
    this.setState(state => ({ count: state.count + increment }));
  }
  
  decrement(decrement: number = 1) {
    this.setState(state => ({ count: state.count - decrement }));
  }
}

@Component({
  selector: 'my-component',
  template: `
    <h1>
      Counter: {{ counter$ | async }}
    </h1>
    <button (click)="increment()">Increment</button>
    <button (click)="decrement()">Decrement</button>
  `,
  providers: [
    CounterStore
  ]
})
export class MyComponent {
  counter$: Observable<number> = this.counterStore.select<number>(state => state.count);
  
  constructor(private counterStore: CounterStore) {}
  
  increment() {
    this.counterStore.increment();
  }
  
  decrement() {
    this.counterStore.decrement();
  }
}