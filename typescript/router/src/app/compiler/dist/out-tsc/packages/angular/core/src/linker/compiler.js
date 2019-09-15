"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var di_1 = require("../di");
/**
 * Combination of NgModuleFactory and ComponentFactorys.
 *
 * @experimental
 */
var /**
 * Combination of NgModuleFactory and ComponentFactorys.
 *
 * @experimental
 */
ModuleWithComponentFactories = /** @class */ (function () {
    function ModuleWithComponentFactories(ngModuleFactory, componentFactories) {
        this.ngModuleFactory = ngModuleFactory;
        this.componentFactories = componentFactories;
    }
    return ModuleWithComponentFactories;
}());
exports.ModuleWithComponentFactories = ModuleWithComponentFactories;
function _throwError() {
    throw new Error("Runtime compiler is not loaded");
}
/**
 * Low-level service for running the angular compiler during runtime
 * to create {@link ComponentFactory}s, which
 * can later be used to create and render a Component instance.
 *
 * Each `@NgModule` provides an own `Compiler` to its injector,
 * that will use the directives/pipes of the ng module for compilation
 * of components.
 *
 */
var Compiler = /** @class */ (function () {
    function Compiler() {
    }
    /**
     * Compiles the given NgModule and all of its components. All templates of the components listed
     * in `entryComponents` have to be inlined.
     */
    /**
       * Compiles the given NgModule and all of its components. All templates of the components listed
       * in `entryComponents` have to be inlined.
       */
    Compiler.prototype.compileModuleSync = /**
       * Compiles the given NgModule and all of its components. All templates of the components listed
       * in `entryComponents` have to be inlined.
       */
    function (moduleType) { throw _throwError(); };
    /**
     * Compiles the given NgModule and all of its components
     */
    /**
       * Compiles the given NgModule and all of its components
       */
    Compiler.prototype.compileModuleAsync = /**
       * Compiles the given NgModule and all of its components
       */
    function (moduleType) { throw _throwError(); };
    /**
     * Same as {@link #compileModuleSync} but also creates ComponentFactories for all components.
     */
    /**
       * Same as {@link #compileModuleSync} but also creates ComponentFactories for all components.
       */
    Compiler.prototype.compileModuleAndAllComponentsSync = /**
       * Same as {@link #compileModuleSync} but also creates ComponentFactories for all components.
       */
    function (moduleType) {
        throw _throwError();
    };
    /**
     * Same as {@link #compileModuleAsync} but also creates ComponentFactories for all components.
     */
    /**
       * Same as {@link #compileModuleAsync} but also creates ComponentFactories for all components.
       */
    Compiler.prototype.compileModuleAndAllComponentsAsync = /**
       * Same as {@link #compileModuleAsync} but also creates ComponentFactories for all components.
       */
    function (moduleType) {
        throw _throwError();
    };
    /**
     * Clears all caches.
     */
    /**
       * Clears all caches.
       */
    Compiler.prototype.clearCache = /**
       * Clears all caches.
       */
    function () { };
    /**
     * Clears the cache for the given component/ngModule.
     */
    /**
       * Clears the cache for the given component/ngModule.
       */
    Compiler.prototype.clearCacheFor = /**
       * Clears the cache for the given component/ngModule.
       */
    function (type) { };
    Compiler.decorators = [
        { type: di_1.Injectable },
    ];
    return Compiler;
}());
exports.Compiler = Compiler;
/**
 * Token to provide CompilerOptions in the platform injector.
 *
 * @experimental
 */
exports.COMPILER_OPTIONS = new di_1.InjectionToken('compilerOptions');
/**
 * A factory for creating a Compiler
 *
 * @experimental
 */
var /**
 * A factory for creating a Compiler
 *
 * @experimental
 */
CompilerFactory = /** @class */ (function () {
    function CompilerFactory() {
    }
    return CompilerFactory;
}());
exports.CompilerFactory = CompilerFactory;
//# sourceMappingURL=compiler.js.map