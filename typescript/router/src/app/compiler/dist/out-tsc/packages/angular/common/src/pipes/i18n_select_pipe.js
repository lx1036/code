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
var invalid_pipe_argument_error_1 = require("./invalid_pipe_argument_error");
/**
 * @ngModule CommonModule
 * @description
 *
 * Generic selector that displays the string that matches the current value.
 *
 * If none of the keys of the `mapping` match the `value`, then the content
 * of the `other` key is returned when present, otherwise an empty string is returned.
 *
 * ## Example
 *
 * {@example common/pipes/ts/i18n_pipe.ts region='I18nSelectPipeComponent'}
 *
 * @experimental
 */
var I18nSelectPipe = /** @class */ (function () {
    function I18nSelectPipe() {
    }
    /**
     * @param value a string to be internationalized.
     * @param mapping an object that indicates the text that should be displayed
     * for different values of the provided `value`.
     */
    /**
       * @param value a string to be internationalized.
       * @param mapping an object that indicates the text that should be displayed
       * for different values of the provided `value`.
       */
    I18nSelectPipe.prototype.transform = /**
       * @param value a string to be internationalized.
       * @param mapping an object that indicates the text that should be displayed
       * for different values of the provided `value`.
       */
    function (value, mapping) {
        if (value == null)
            return '';
        if (typeof mapping !== 'object' || typeof value !== 'string') {
            throw invalid_pipe_argument_error_1.invalidPipeArgumentError(I18nSelectPipe, mapping);
        }
        if (mapping.hasOwnProperty(value)) {
            return mapping[value];
        }
        if (mapping.hasOwnProperty('other')) {
            return mapping['other'];
        }
        return '';
    };
    I18nSelectPipe.decorators = [
        { type: core_1.Pipe, args: [{ name: 'i18nSelect', pure: true },] },
    ];
    return I18nSelectPipe;
}());
exports.I18nSelectPipe = I18nSelectPipe;
//# sourceMappingURL=i18n_select_pipe.js.map