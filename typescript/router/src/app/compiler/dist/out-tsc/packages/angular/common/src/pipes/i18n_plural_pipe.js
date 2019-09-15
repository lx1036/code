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
var localization_1 = require("../i18n/localization");
var invalid_pipe_argument_error_1 = require("./invalid_pipe_argument_error");
var _INTERPOLATION_REGEXP = /#/g;
/**
 * @ngModule CommonModule
 * @description
 *
 * Maps a value to a string that pluralizes the value according to locale rules.
 *
 *  ## Example
 *
 * {@example common/pipes/ts/i18n_pipe.ts region='I18nPluralPipeComponent'}
 *
 * @experimental
 */
var I18nPluralPipe = /** @class */ (function () {
    function I18nPluralPipe(_localization) {
        this._localization = _localization;
    }
    /**
     * @param value the number to be formatted
     * @param pluralMap an object that mimics the ICU format, see
     * http://userguide.icu-project.org/formatparse/messages.
     * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
     * default).
     */
    /**
       * @param value the number to be formatted
       * @param pluralMap an object that mimics the ICU format, see
       * http://userguide.icu-project.org/formatparse/messages.
       * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
       * default).
       */
    I18nPluralPipe.prototype.transform = /**
       * @param value the number to be formatted
       * @param pluralMap an object that mimics the ICU format, see
       * http://userguide.icu-project.org/formatparse/messages.
       * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
       * default).
       */
    function (value, pluralMap, locale) {
        if (value == null)
            return '';
        if (typeof pluralMap !== 'object' || pluralMap === null) {
            throw invalid_pipe_argument_error_1.invalidPipeArgumentError(I18nPluralPipe, pluralMap);
        }
        var key = localization_1.getPluralCategory(value, Object.keys(pluralMap), this._localization, locale);
        return pluralMap[key].replace(_INTERPOLATION_REGEXP, value.toString());
    };
    I18nPluralPipe.decorators = [
        { type: core_1.Pipe, args: [{ name: 'i18nPlural', pure: true },] },
    ];
    /** @nocollapse */
    I18nPluralPipe.ctorParameters = function () { return [
        { type: localization_1.NgLocalization, },
    ]; };
    return I18nPluralPipe;
}());
exports.I18nPluralPipe = I18nPluralPipe;
//# sourceMappingURL=i18n_plural_pipe.js.map