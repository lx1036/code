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
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
/**
 * @module
 * @description
 * Entry point from which you should import all public core APIs.
 */
__export(require("./metadata"));
__export(require("./version"));
__export(require("./di"));
var application_ref_1 = require("./application_ref");
exports.createPlatform = application_ref_1.createPlatform;
exports.assertPlatform = application_ref_1.assertPlatform;
exports.destroyPlatform = application_ref_1.destroyPlatform;
exports.getPlatform = application_ref_1.getPlatform;
exports.PlatformRef = application_ref_1.PlatformRef;
exports.ApplicationRef = application_ref_1.ApplicationRef;
exports.enableProdMode = application_ref_1.enableProdMode;
exports.isDevMode = application_ref_1.isDevMode;
exports.createPlatformFactory = application_ref_1.createPlatformFactory;
exports.NgProbeToken = application_ref_1.NgProbeToken;
var application_tokens_1 = require("./application_tokens");
exports.APP_ID = application_tokens_1.APP_ID;
exports.PACKAGE_ROOT_URL = application_tokens_1.PACKAGE_ROOT_URL;
exports.PLATFORM_INITIALIZER = application_tokens_1.PLATFORM_INITIALIZER;
exports.PLATFORM_ID = application_tokens_1.PLATFORM_ID;
exports.APP_BOOTSTRAP_LISTENER = application_tokens_1.APP_BOOTSTRAP_LISTENER;
var application_init_1 = require("./application_init");
exports.APP_INITIALIZER = application_init_1.APP_INITIALIZER;
exports.ApplicationInitStatus = application_init_1.ApplicationInitStatus;
__export(require("./zone"));
__export(require("./render"));
__export(require("./linker"));
var debug_node_1 = require("./debug/debug_node");
exports.DebugElement = debug_node_1.DebugElement;
exports.DebugNode = debug_node_1.DebugNode;
exports.asNativeElements = debug_node_1.asNativeElements;
exports.getDebugNode = debug_node_1.getDebugNode;
var testability_1 = require("./testability/testability");
exports.Testability = testability_1.Testability;
exports.TestabilityRegistry = testability_1.TestabilityRegistry;
exports.setTestabilityGetter = testability_1.setTestabilityGetter;
__export(require("./change_detection"));
__export(require("./platform_core_providers"));
var tokens_1 = require("./i18n/tokens");
exports.TRANSLATIONS = tokens_1.TRANSLATIONS;
exports.TRANSLATIONS_FORMAT = tokens_1.TRANSLATIONS_FORMAT;
exports.LOCALE_ID = tokens_1.LOCALE_ID;
exports.MissingTranslationStrategy = tokens_1.MissingTranslationStrategy;
var application_module_1 = require("./application_module");
exports.ApplicationModule = application_module_1.ApplicationModule;
var profile_1 = require("./profile/profile");
exports.wtfCreateScope = profile_1.wtfCreateScope;
exports.wtfLeave = profile_1.wtfLeave;
exports.wtfStartTimeRange = profile_1.wtfStartTimeRange;
exports.wtfEndTimeRange = profile_1.wtfEndTimeRange;
var type_1 = require("./type");
exports.Type = type_1.Type;
var event_emitter_1 = require("./event_emitter");
exports.EventEmitter = event_emitter_1.EventEmitter;
var error_handler_1 = require("./error_handler");
exports.ErrorHandler = error_handler_1.ErrorHandler;
__export(require("./core_private_export"));
__export(require("./core_render3_private_export"));
var security_1 = require("./sanitization/security");
exports.Sanitizer = security_1.Sanitizer;
exports.SecurityContext = security_1.SecurityContext;
__export(require("./codegen_private_exports"));
//# sourceMappingURL=core.js.map