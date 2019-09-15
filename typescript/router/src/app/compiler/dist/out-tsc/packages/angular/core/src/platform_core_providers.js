"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var application_ref_1 = require("./application_ref");
var application_tokens_1 = require("./application_tokens");
var console_1 = require("./console");
var di_1 = require("./di");
var testability_1 = require("./testability/testability");
var _CORE_PLATFORM_PROVIDERS = [
    // Set a default platform name for platforms that don't set it explicitly.
    { provide: application_tokens_1.PLATFORM_ID, useValue: 'unknown' },
    { provide: application_ref_1.PlatformRef, deps: [di_1.Injector] },
    { provide: testability_1.TestabilityRegistry, deps: [] },
    { provide: console_1.Console, deps: [] },
];
/**
 * This platform has to be included in any other platform
 *
 * @experimental
 */
exports.platformCore = application_ref_1.createPlatformFactory(null, 'core', _CORE_PLATFORM_PROVIDERS);
//# sourceMappingURL=platform_core_providers.js.map