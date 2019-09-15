"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var injector_1 = require("../packages/angular/core/src/di/injector");
var injectable_1 = require("../packages/angular/core/src/di/injectable");
var Engine = /** @class */ (function () {
    function Engine() {
    }
    return Engine;
}());
exports.Engine = Engine;
var Wheel = /** @class */ (function () {
    function Wheel() {
    }
    return Wheel;
}());
exports.Wheel = Wheel;
var Car = /** @class */ (function () {
    function Car(_engine, _wheel) {
        this._engine = _engine;
        this._wheel = _wheel;
    }
    Car.decorators = [
        { type: injectable_1.Injectable },
    ];
    /** @nocollapse */
    Car.ctorParameters = function () { return [
        { type: Engine, },
        { type: Wheel, },
    ]; };
    return Car;
}());
exports.Car = Car;
var injector = new injector_1.StaticInjector([
    Car, Wheel, Engine
    // { provide: Engine, useClass: Engine },
    // { provide: Wheel, useClass: Wheel },
    // { provide: Car, useClass: Car, deps: [Engine, Wheel] },
]);
console.log(injector.get(Car));
//# sourceMappingURL=test-injector.js.map