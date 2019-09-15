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
var index_1 = require("./directives/index");
var localization_1 = require("./i18n/localization");
var index_2 = require("./pipes/deprecated/index");
var index_3 = require("./pipes/index");
// Note: This does not contain the location providers,
// as they need some platform specific implementations to work.
/**
 * The module that includes all the basic Angular directives like {@link NgIf}, {@link NgForOf}, ...
 *
 *
 */
var CommonModule = /** @class */ (function () {
    function CommonModule() {
    }
    CommonModule.decorators = [
        { type: core_1.NgModule, args: [{
                    declarations: [index_1.COMMON_DIRECTIVES, index_3.COMMON_PIPES],
                    exports: [index_1.COMMON_DIRECTIVES, index_3.COMMON_PIPES],
                    providers: [
                        { provide: localization_1.NgLocalization, useClass: localization_1.NgLocaleLocalization },
                    ],
                },] },
    ];
    return CommonModule;
}());
exports.CommonModule = CommonModule;
var ɵ0 = localization_1.getPluralCase;
exports.ɵ0 = ɵ0;
/**
 * A module that contains the deprecated i18n pipes.
 *
 * @deprecated from v5
 */
var DeprecatedI18NPipesModule = /** @class */ (function () {
    function DeprecatedI18NPipesModule() {
    }
    DeprecatedI18NPipesModule.decorators = [
        { type: core_1.NgModule, args: [{
                    declarations: [index_2.COMMON_DEPRECATED_I18N_PIPES],
                    exports: [index_2.COMMON_DEPRECATED_I18N_PIPES],
                    providers: [{ provide: localization_1.DEPRECATED_PLURAL_FN, useValue: ɵ0 }],
                },] },
    ];
    return DeprecatedI18NPipesModule;
}());
exports.DeprecatedI18NPipesModule = DeprecatedI18NPipesModule;
//# sourceMappingURL=common_module.js.map