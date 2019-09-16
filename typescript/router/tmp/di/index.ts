

import 'reflect-metadata';

/**
 * yarn ts-node src/app/di/index.ts
 * @see https://github.com/ZixiaoWang/di.git
 */

export class Injector {
  private _set: Set<any>;

  constructor() {
    this._set = new Set();
  }

  has(paramType) {
    return this._set.has(paramType);
  }

  add(paramType) {
    this._set.add(paramType);
  }
}

const injector = new Injector();


export class InstanceStore {
  private _map: Map<any, any>;

  constructor() {
    this._map = new Map<any, any>();
  }

  add(providers: ProviderConfig[] = []): InstanceStore {
    let map = this.construct(providers);

    map.forEach((val, key) => {
      this._map.set(key, val);
    });

    return this;
  }

  construct(providers: ProviderConfig[]): Map<any, any> {
    let map = new Map();
    let useFunction = [];
    let useClass = [];
    let useVal = [];
    let useExist = [];

    providers.forEach((item: ProviderConfig) => {
      if (typeof item === 'function') {
        useFunction.push(item);
      } else if (typeof item === 'object') {
        if (item.useClass) {
          useClass.push(item);
        } else if (item.useValue) {
          useVal.push(item);
        } else if (item.useExist) {
          useExist.push(item);
        } else {
          throw new Error(`${item} is not an illegal ProviderConfig type.`);
        }
      } else {
        throw new Error(`${item} is illegal type.`);
      }
    });

    let list = useFunction.concat(useClass, useVal, useExist);
    let value;

    list.forEach((item: ProviderConfig) => {
      if (item.useClass) {
        value = instance(item.useClass);
      } else if (item.useValue) {
        value = item.useValue;
      } else if (item.useExist) {
        value = this._map.get(item.provide);
      }

      map.set(item.provide, value);
    });

    return map;
  }

  get(token) {
    return this._map.get(token);
  }
}

const instanceStore = new InstanceStore();

export interface ComponentStoreConfig {
  priority: number,
  restrict: boolean,
  instanceStore: InstanceStore
}

export class ComponentStore {
  private _map: Map<any,ComponentStoreConfig>;

  constructor() {
    this._map = new Map();
  }

  add(component: any, config?: ComponentStoreConfig) {
    this._map.set(component, config || {priority: 0, restrict: true, instanceStore: new InstanceStore()});
  }

  get(fn) {
    return this._map.get(fn);
  }

  getInstanceByType(component: any, paramType: any) {
    if (this._map.has(component)) {
      return this._map.get(component).instanceStore.get(paramType);
    }

    return null;
  }
}

const componentStore = new ComponentStore();




export function instance(fn) {
  let args = Reflect.getMetadata('design:paramtypes', fn) || [];

  args = args.map(paramType => {
    if (injector.has(paramType)) {
      return instance(paramType);
    } else {
      throw new Error(`${paramType.name} is not an injectable type, please add @Injectable to register class.`);
    }
  });

  return new fn(...args);
}

function container(fn) {
  let args = Reflect.getMetadata('design:paramtypes', fn) || [];

  console.log(args); // [ [Function: Wheel], [Function: V12Engine] ]

  let config: ComponentStoreConfig = componentStore.get(fn);

  args = args.map(paramType => {
    // @Component
    if (config.priority === 1) {
      let localInstance = componentStore.getInstanceByType(fn, paramType);
      let globalInstance = instanceStore.get(paramType);

      return localInstance || globalInstance;
    } else if (config.priority === 2) { // @Inject
      return componentStore.getInstanceByType(fn, paramType);
    } else if (config.priority < 1 || config.priority > 2) {
      throw new Error(`Incorrect config for ${fn.name}.`);
    }
  });

  return new fn(...args);
}

export interface ComponentConfig {
  providers: ProviderConfig[],
  restrict?: boolean
}

export interface ProviderConfig {
  provide: any,
  useValue?: any,
  useClass?: any,
  useExist?: any
}


/**
 * Decorator
 * @see https://www.tslang.cn/docs/handbook/decorators.html
 */
export function Component(config?: ComponentConfig): Function {
  console.log(config);

  return function (target) {
    let instanceStore = new InstanceStore();

    if (config) {
      instanceStore.add(config.providers || []);
    }

    console.log(target);

    componentStore.add(target, {priority: 1, restrict: config.restrict, instanceStore: instanceStore});

    return target;
  }
}

export function Injectable(): Function {
  return function (target) {
    console.log(target);

    injector.add(target);

    return target;
  }
}


/**
 * Test Case
 */
@Injectable()
class Wheel {
  showInfo(){ console.log('Wheel module is working OK...') };
}

@Injectable()
class Engine {
  showInfo(){ console.log('Engine module is working OK...') };
}

@Injectable()
class V12Engine {
  showInfo(){ console.log('V12 Engine starts!') }
}


@Component({
  providers: [
    { provide: Engine, useClass: V12Engine },

    {provide: Wheel, useClass: Wheel}
  ]
})
class Car {
  constructor( private wheel: Wheel, private engine: Engine ) {
    this.wheel.showInfo();
    this.engine.showInfo();
  }
}

function bootstrap(config: ComponentConfig) {
  instanceStore.add(config.providers);
}

// bootstrap({
//   providers: [
//     {provide: Wheel, useClass: Wheel},
//     {provide: Engine, useClass: Engine},
//     {provide: V12Engine, useClass: V12Engine},
//   ]
// });

// var car = new Car( new Wheel(), new Engine() );
var racingCar = container(Car);

console.log(racingCar);
