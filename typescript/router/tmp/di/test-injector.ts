import {Injector, Injectable} from "@angular/core";


export class Engine {

}

export class Wheel {

}

@Injectable()
export class Car {
  constructor(private _engine: Engine, private _wheel: Wheel) {}
}

// The right mental model is to think that every DOM element has an Injector. (In practice, only interesting elements containing directives will have an injector, but this is a performance optimization)
const injector = Injector.create([
  // Car, Wheel, Engine
  {provide: Engine, useClass: Engine},
  {provide: Wheel, useClass: Wheel},
  {provide: Car, useClass: Car, deps: [Engine, Wheel]},
]);

// console.log(injector.get(Car));
