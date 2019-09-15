"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var application_init_1 = require("./application_init");
var application_ref_1 = require("./application_ref");
var application_tokens_1 = require("./application_tokens");
var change_detection_1 = require("./change_detection/change_detection");
var metadata_1 = require("./di/metadata");
var tokens_1 = require("./i18n/tokens");
var compiler_1 = require("./linker/compiler");
var metadata_2 = require("./metadata");
function _iterableDiffersFactory() {
    return change_detection_1.defaultIterableDiffers;
}
exports._iterableDiffersFactory = _iterableDiffersFactory;
function _keyValueDiffersFactory() {
    return change_detection_1.defaultKeyValueDiffers;
}
exports._keyValueDiffersFactory = _keyValueDiffersFactory;
function _localeFactory(locale) {
    return locale || 'en-US';
}
exports._localeFactory = _localeFactory;
/**
 * This module includes the providers of @angular/core that are needed
 * to bootstrap components via `ApplicationRef`.
 *
 * @experimental
 */
var ApplicationModule = /** @class */ (function () {
    // Inject ApplicationRef to make it eager...
    function ApplicationModule(appRef) {
    }
    ApplicationModule.decorators = [
        { type: metadata_2.NgModule, args: [{
                    providers: [
                        application_ref_1.ApplicationRef,
                        application_init_1.ApplicationInitStatus,
                        compiler_1.Compiler,
                        application_tokens_1.APP_ID_RANDOM_PROVIDER,
                        { provide: change_detection_1.IterableDiffers, useFactory: _iterableDiffersFactory },
                        { provide: change_detection_1.KeyValueDiffers, useFactory: _keyValueDiffersFactory },
                        {
                            provide: tokens_1.LOCALE_ID,
                            useFactory: _localeFactory,
                            deps: [[new metadata_1.Inject(tokens_1.LOCALE_ID), new metadata_1.Optional(), new metadata_1.SkipSelf()]]
                        },
                    ]
                },] },
    ];
    /** @nocollapse */
    ApplicationModule.ctorParameters = function () { return [
        { type: application_ref_1.ApplicationRef, },
    ]; };
    return ApplicationModule;
}());
exports.ApplicationModule = ApplicationModule;
//# sourceMappingURL=application_module.js.map