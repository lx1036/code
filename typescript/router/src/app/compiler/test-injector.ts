import {StaticInjector} from "../packages/angular/core/src/di/injector";
import {Injectable} from "../packages/angular/core/src/di/injectable";


@Injectable()
export class Engine {

}

@Injectable()
export class Wheel {

}

@Injectable()
export class Car {
  constructor(private _engine: Engine, private _wheel: Wheel) {}
}

const injector = new StaticInjector([
  // Car, Wheel, Engine
  {provide: Engine, useClass: Engine},
  {provide: Wheel, useClass: Wheel},
  {provide: Car, useClass: Car, deps: [Engine, Wheel]},
]);

console.log(injector.get(Car));
