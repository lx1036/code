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
var compiler_1 = require("./compiler");
var _SEPARATOR = '#';
var FACTORY_CLASS_SUFFIX = 'NgFactory';
/**
 * Configuration for SystemJsNgModuleLoader.
 * token.
 *
 * @experimental
 */
var /**
 * Configuration for SystemJsNgModuleLoader.
 * token.
 *
 * @experimental
 */
SystemJsNgModuleLoaderConfig = /** @class */ (function () {
    function SystemJsNgModuleLoaderConfig() {
    }
    return SystemJsNgModuleLoaderConfig;
}());
exports.SystemJsNgModuleLoaderConfig = SystemJsNgModuleLoaderConfig;
var DEFAULT_CONFIG = {
    factoryPathPrefix: '',
    factoryPathSuffix: '.ngfactory',
};
/**
 * NgModuleFactoryLoader that uses SystemJS to load NgModuleFactory
 * @experimental
 */
var SystemJsNgModuleLoader = /** @class */ (function () {
    function SystemJsNgModuleLoader(_compiler, config) {
        this._compiler = _compiler;
        this._config = config || DEFAULT_CONFIG;
    }
    SystemJsNgModuleLoader.prototype.load = function (path) {
        var offlineMode = this._compiler instanceof compiler_1.Compiler;
        return offlineMode ? this.loadFactory(path) : this.loadAndCompile(path);
    };
    SystemJsNgModuleLoader.prototype.loadAndCompile = function (path) {
        var _this = this;
        var _a = path.split(_SEPARATOR), module = _a[0], exportName = _a[1];
        if (exportName === undefined) {
            exportName = 'default';
        }
        return System.import(module)
            .then(function (module) { return module[exportName]; })
            .then(function (type) { return checkNotEmpty(type, module, exportName); })
            .then(function (type) { return _this._compiler.compileModuleAsync(type); });
    };
    SystemJsNgModuleLoader.prototype.loadFactory = function (path) {
        var _a = path.split(_SEPARATOR), module = _a[0], exportName = _a[1];
        var factoryClassSuffix = FACTORY_CLASS_SUFFIX;
        if (exportName === undefined) {
            exportName = 'default';
            factoryClassSuffix = '';
        }
        return System.import(this._config.factoryPathPrefix + module + this._config.factoryPathSuffix)
            .then(function (module) { return module[exportName + factoryClassSuffix]; })
            .then(function (factory) { return checkNotEmpty(factory, module, exportName); });
    };
    SystemJsNgModuleLoader.decorators = [
        { type: di_1.Injectable },
    ];
    /** @nocollapse */
    SystemJsNgModuleLoader.ctorParameters = function () { return [
        { type: compiler_1.Compiler, },
        { type: SystemJsNgModuleLoaderConfig, decorators: [{ type: di_1.Optional },] },
    ]; };
    return SystemJsNgModuleLoader;
}());
exports.SystemJsNgModuleLoader = SystemJsNgModuleLoader;
function checkNotEmpty(value, modulePath, exportName) {
    if (!value) {
        throw new Error("Cannot find '" + exportName + "' in '" + modulePath + "'");
    }
    return value;
}
//# sourceMappingURL=system_js_ng_module_factory_loader.js.map