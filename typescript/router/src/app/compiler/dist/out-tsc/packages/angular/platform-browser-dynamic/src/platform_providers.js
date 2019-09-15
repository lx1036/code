"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var common_1 = require("@angular/common");
var compiler_1 = require("@angular/compiler");
var core_1 = require("@angular/core");
var platform_browser_1 = require("@angular/platform-browser");
var resource_loader_impl_1 = require("./resource_loader/resource_loader_impl");
exports.INTERNAL_BROWSER_DYNAMIC_PLATFORM_PROVIDERS = [
    platform_browser_1.ɵINTERNAL_BROWSER_PLATFORM_PROVIDERS,
    {
        provide: core_1.COMPILER_OPTIONS,
        useValue: { providers: [{ provide: compiler_1.ResourceLoader, useClass: resource_loader_impl_1.ResourceLoaderImpl, deps: [] }] },
        multi: true
    },
    { provide: core_1.PLATFORM_ID, useValue: common_1.ɵPLATFORM_BROWSER_ID },
];
//# sourceMappingURL=platform_providers.js.map