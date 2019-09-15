"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var compiler_factory_1 = require("./compiler_factory");
/**
 * A platform that included corePlatform and the compiler.
 *
 * @experimental
 */
exports.platformCoreDynamic = core_1.createPlatformFactory(core_1.platformCore, 'coreDynamic', [
    { provide: core_1.COMPILER_OPTIONS, useValue: {}, multi: true },
    { provide: core_1.CompilerFactory, useClass: compiler_factory_1.JitCompilerFactory, deps: [core_1.COMPILER_OPTIONS] },
]);
//# sourceMappingURL=platform_core_dynamic.js.map