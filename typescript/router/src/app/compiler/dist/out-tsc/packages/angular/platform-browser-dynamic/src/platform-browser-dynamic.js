"use strict";
function __export(m) {
    for (var p in m) if (!exports.hasOwnProperty(p)) exports[p] = m[p];
}
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var compiler_1 = require("@angular/compiler");
var core_1 = require("@angular/core");
var platform_core_dynamic_1 = require("./platform_core_dynamic");
var platform_providers_1 = require("./platform_providers");
var resource_loader_cache_1 = require("./resource_loader/resource_loader_cache");
__export(require("./private_export"));
var version_1 = require("./version");
exports.VERSION = version_1.VERSION;
var compiler_factory_1 = require("./compiler_factory");
exports.JitCompilerFactory = compiler_factory_1.JitCompilerFactory;
/**
 * @experimental
 */
exports.RESOURCE_CACHE_PROVIDER = [{ provide: compiler_1.ResourceLoader, useClass: resource_loader_cache_1.CachedResourceLoader, deps: [] }];
exports.platformBrowserDynamic = core_1.createPlatformFactory(platform_core_dynamic_1.platformCoreDynamic, 'browserDynamic', platform_providers_1.INTERNAL_BROWSER_DYNAMIC_PLATFORM_PROVIDERS);
//# sourceMappingURL=platform-browser-dynamic.js.map