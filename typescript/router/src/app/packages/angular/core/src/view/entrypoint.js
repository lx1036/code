"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
var __extends = (this && this.__extends) || (function () {
    var extendStatics = Object.setPrototypeOf ||
        ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
        function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
exports.__esModule = true;
var injector_1 = require("../di/injector");
var ng_module_factory_1 = require("../linker/ng_module_factory");
var services_1 = require("./services");
var types_1 = require("./types");
var util_1 = require("./util");
function overrideProvider(override) {
    services_1.initServicesIfNeeded();
    return types_1.Services.overrideProvider(override);
}
exports.overrideProvider = overrideProvider;
function overrideComponentView(comp, componentFactory) {
    services_1.initServicesIfNeeded();
    return types_1.Services.overrideComponentView(comp, componentFactory);
}
exports.overrideComponentView = overrideComponentView;
function clearOverrides() {
    services_1.initServicesIfNeeded();
    return types_1.Services.clearOverrides();
}
exports.clearOverrides = clearOverrides;
// Attention: this function is called as top level function.
// Putting any logic in here will destroy closure tree shaking!
function createNgModuleFactory(ngModuleType, bootstrapComponents, defFactory) {
    return new NgModuleFactory_(ngModuleType, bootstrapComponents, defFactory);
}
exports.createNgModuleFactory = createNgModuleFactory;
var NgModuleFactory_ = /** @class */ (function (_super) {
    __extends(NgModuleFactory_, _super);
    function NgModuleFactory_(moduleType, _bootstrapComponents, _ngModuleDefFactory) {
        var _this = 
        // Attention: this ctor is called as top level function.
        // Putting any logic in here will destroy closure tree shaking!
        _super.call(this) || this;
        _this.moduleType = moduleType;
        _this._bootstrapComponents = _bootstrapComponents;
        _this._ngModuleDefFactory = _ngModuleDefFactory;
        return _this;
    }
    NgModuleFactory_.prototype.create = function (parentInjector) {
        services_1.initServicesIfNeeded();
        var def = util_1.resolveDefinition(this._ngModuleDefFactory);
        return types_1.Services.createNgModuleRef(this.moduleType, parentInjector || injector_1.Injector.NULL, this._bootstrapComponents, def);
    };
    return NgModuleFactory_;
}(ng_module_factory_1.NgModuleFactory));
