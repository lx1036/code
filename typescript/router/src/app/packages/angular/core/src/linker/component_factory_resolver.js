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
var util_1 = require("../util");
var component_factory_1 = require("./component_factory");
function noComponentFactoryError(component) {
    var error = Error("No component factory found for " + util_1.stringify(component) + ". Did you add it to @NgModule.entryComponents?");
    error[ERROR_COMPONENT] = component;
    return error;
}
exports.noComponentFactoryError = noComponentFactoryError;
var ERROR_COMPONENT = 'ngComponent';
function getComponent(error) {
    return error[ERROR_COMPONENT];
}
exports.getComponent = getComponent;
var _NullComponentFactoryResolver = /** @class */ (function () {
    function _NullComponentFactoryResolver() {
    }
    _NullComponentFactoryResolver.prototype.resolveComponentFactory = function (component) {
        throw noComponentFactoryError(component);
    };
    return _NullComponentFactoryResolver;
}());
var ComponentFactoryResolver = /** @class */ (function () {
    function ComponentFactoryResolver() {
    }
    ComponentFactoryResolver.NULL = new _NullComponentFactoryResolver();
    return ComponentFactoryResolver;
}());
exports.ComponentFactoryResolver = ComponentFactoryResolver;
var CodegenComponentFactoryResolver = /** @class */ (function () {
    function CodegenComponentFactoryResolver(factories, _parent, _ngModule) {
        this._parent = _parent;
        this._ngModule = _ngModule;
        this._factories = new Map();
        for (var i = 0; i < factories.length; i++) {
            var factory = factories[i];
            this._factories.set(factory.componentType, factory);
        }
    }
    CodegenComponentFactoryResolver.prototype.resolveComponentFactory = function (component) {
        var factory = this._factories.get(component);
        if (!factory && this._parent) {
            factory = this._parent.resolveComponentFactory(component);
        }
        if (!factory) {
            throw noComponentFactoryError(component);
        }
        return new ComponentFactoryBoundToModule(factory, this._ngModule);
    };
    return CodegenComponentFactoryResolver;
}());
exports.CodegenComponentFactoryResolver = CodegenComponentFactoryResolver;
var ComponentFactoryBoundToModule = /** @class */ (function (_super) {
    __extends(ComponentFactoryBoundToModule, _super);
    function ComponentFactoryBoundToModule(factory, ngModule) {
        var _this = _super.call(this) || this;
        _this.factory = factory;
        _this.ngModule = ngModule;
        _this.selector = factory.selector;
        _this.componentType = factory.componentType;
        _this.ngContentSelectors = factory.ngContentSelectors;
        _this.inputs = factory.inputs;
        _this.outputs = factory.outputs;
        return _this;
    }
    ComponentFactoryBoundToModule.prototype.create = function (injector, projectableNodes, rootSelectorOrNode, ngModule) {
        return this.factory.create(injector, projectableNodes, rootSelectorOrNode, ngModule || this.ngModule);
    };
    return ComponentFactoryBoundToModule;
}(component_factory_1.ComponentFactory));
exports.ComponentFactoryBoundToModule = ComponentFactoryBoundToModule;
