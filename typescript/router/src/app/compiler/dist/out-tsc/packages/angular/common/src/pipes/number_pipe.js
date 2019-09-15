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
var format_number_1 = require("../i18n/format_number");
var locale_data_api_1 = require("../i18n/locale_data_api");
var invalid_pipe_argument_error_1 = require("./invalid_pipe_argument_error");
/**
 * @ngModule CommonModule
 * @description
 *
 * Uses the function {@link formatNumber} to format a number according to locale rules.
 *
 * Formats a number as text. Group sizing and separator and other locale-specific
 * configurations are based on the locale.
 *
 * ### Example
 *
 * {@example common/pipes/ts/number_pipe.ts region='NumberPipe'}
 *
 *
 */
var DecimalPipe = /** @class */ (function () {
    function DecimalPipe(_locale) {
        this._locale = _locale;
    }
    /**
     * @param value a number to be formatted.
     * @param digitsInfo a `string` which has a following format: <br>
     * <code>{minIntegerDigits}.{minFractionDigits}-{maxFractionDigits}</code>.
     *   - `minIntegerDigits` is the minimum number of integer digits to use. Defaults to `1`.
     *   - `minFractionDigits` is the minimum number of digits after the decimal point. Defaults to
     * `0`.
     *   - `maxFractionDigits` is the maximum number of digits after the decimal point. Defaults to
     * `3`.
     * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
     * default).
     */
    /**
       * @param value a number to be formatted.
       * @param digitsInfo a `string` which has a following format: <br>
       * <code>{minIntegerDigits}.{minFractionDigits}-{maxFractionDigits}</code>.
       *   - `minIntegerDigits` is the minimum number of integer digits to use. Defaults to `1`.
       *   - `minFractionDigits` is the minimum number of digits after the decimal point. Defaults to
       * `0`.
       *   - `maxFractionDigits` is the maximum number of digits after the decimal point. Defaults to
       * `3`.
       * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
       * default).
       */
    DecimalPipe.prototype.transform = /**
       * @param value a number to be formatted.
       * @param digitsInfo a `string` which has a following format: <br>
       * <code>{minIntegerDigits}.{minFractionDigits}-{maxFractionDigits}</code>.
       *   - `minIntegerDigits` is the minimum number of integer digits to use. Defaults to `1`.
       *   - `minFractionDigits` is the minimum number of digits after the decimal point. Defaults to
       * `0`.
       *   - `maxFractionDigits` is the maximum number of digits after the decimal point. Defaults to
       * `3`.
       * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
       * default).
       */
    function (value, digitsInfo, locale) {
        if (isEmpty(value))
            return null;
        locale = locale || this._locale;
        try {
            var num = strToNumber(value);
            return format_number_1.formatNumber(num, locale, digitsInfo);
        }
        catch (error) {
            throw invalid_pipe_argument_error_1.invalidPipeArgumentError(DecimalPipe, error.message);
        }
    };
    DecimalPipe.decorators = [
        { type: core_1.Pipe, args: [{ name: 'number' },] },
    ];
    /** @nocollapse */
    DecimalPipe.ctorParameters = function () { return [
        { type: undefined, decorators: [{ type: core_1.Inject, args: [core_1.LOCALE_ID,] },] },
    ]; };
    return DecimalPipe;
}());
exports.DecimalPipe = DecimalPipe;
/**
 * @ngModule CommonModule
 * @description
 *
 * Uses the function {@link formatPercent} to format a number as a percentage according
 * to locale rules.
 *
 * ### Example
 *
 * {@example common/pipes/ts/percent_pipe.ts region='PercentPipe'}
 *
 *
 */
var PercentPipe = /** @class */ (function () {
    function PercentPipe(_locale) {
        this._locale = _locale;
    }
    /**
     *
     * @param value a number to be formatted as a percentage.
     * @param digitsInfo see {@link DecimalPipe} for more details.
     * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
   * default).
     */
    /**
       *
       * @param value a number to be formatted as a percentage.
       * @param digitsInfo see {@link DecimalPipe} for more details.
       * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
     * default).
       */
    PercentPipe.prototype.transform = /**
       *
       * @param value a number to be formatted as a percentage.
       * @param digitsInfo see {@link DecimalPipe} for more details.
       * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
     * default).
       */
    function (value, digitsInfo, locale) {
        if (isEmpty(value))
            return null;
        locale = locale || this._locale;
        try {
            var num = strToNumber(value);
            return format_number_1.formatPercent(num, locale, digitsInfo);
        }
        catch (error) {
            throw invalid_pipe_argument_error_1.invalidPipeArgumentError(PercentPipe, error.message);
        }
    };
    PercentPipe.decorators = [
        { type: core_1.Pipe, args: [{ name: 'percent' },] },
    ];
    /** @nocollapse */
    PercentPipe.ctorParameters = function () { return [
        { type: undefined, decorators: [{ type: core_1.Inject, args: [core_1.LOCALE_ID,] },] },
    ]; };
    return PercentPipe;
}());
exports.PercentPipe = PercentPipe;
/**
 * @ngModule CommonModule
 * @description
 *
 * Uses the functions {@link getCurrencySymbol} and {@link formatCurrency} to format a
 * number as currency using locale rules.
 *
 * ### Example
 *
 * {@example common/pipes/ts/currency_pipe.ts region='CurrencyPipe'}
 *
 *
 */
var CurrencyPipe = /** @class */ (function () {
    function CurrencyPipe(_locale) {
        this._locale = _locale;
    }
    /**
     *
     * @param value a number to be formatted as currency.
     * @param currencyCodeis the [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217) currency code,
     * such as `USD` for the US dollar and `EUR` for the euro.
     * @param display indicates whether to show the currency symbol, the code or a custom value:
     *   - `code`: use code (e.g. `USD`).
     *   - `symbol`(default): use symbol (e.g. `$`).
     *   - `symbol-narrow`: some countries have two symbols for their currency, one regular and one
     *     narrow (e.g. the canadian dollar CAD has the symbol `CA$` and the symbol-narrow `$`).
     *   - `string`: use this value instead of a code or a symbol.
     *   - boolean (deprecated from v5): `true` for symbol and false for `code`.
     *   If there is no narrow symbol for the chosen currency, the regular symbol will be used.
     * @param digitsInfo see {@link DecimalPipe} for more details.
     * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
     * default).
     */
    /**
       *
       * @param value a number to be formatted as currency.
       * @param currencyCodeis the [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217) currency code,
       * such as `USD` for the US dollar and `EUR` for the euro.
       * @param display indicates whether to show the currency symbol, the code or a custom value:
       *   - `code`: use code (e.g. `USD`).
       *   - `symbol`(default): use symbol (e.g. `$`).
       *   - `symbol-narrow`: some countries have two symbols for their currency, one regular and one
       *     narrow (e.g. the canadian dollar CAD has the symbol `CA$` and the symbol-narrow `$`).
       *   - `string`: use this value instead of a code or a symbol.
       *   - boolean (deprecated from v5): `true` for symbol and false for `code`.
       *   If there is no narrow symbol for the chosen currency, the regular symbol will be used.
       * @param digitsInfo see {@link DecimalPipe} for more details.
       * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
       * default).
       */
    CurrencyPipe.prototype.transform = /**
       *
       * @param value a number to be formatted as currency.
       * @param currencyCodeis the [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217) currency code,
       * such as `USD` for the US dollar and `EUR` for the euro.
       * @param display indicates whether to show the currency symbol, the code or a custom value:
       *   - `code`: use code (e.g. `USD`).
       *   - `symbol`(default): use symbol (e.g. `$`).
       *   - `symbol-narrow`: some countries have two symbols for their currency, one regular and one
       *     narrow (e.g. the canadian dollar CAD has the symbol `CA$` and the symbol-narrow `$`).
       *   - `string`: use this value instead of a code or a symbol.
       *   - boolean (deprecated from v5): `true` for symbol and false for `code`.
       *   If there is no narrow symbol for the chosen currency, the regular symbol will be used.
       * @param digitsInfo see {@link DecimalPipe} for more details.
       * @param locale a `string` defining the locale to use (uses the current {@link LOCALE_ID} by
       * default).
       */
    function (value, currencyCode, display, digitsInfo, locale) {
        if (display === void 0) { display = 'symbol'; }
        if (isEmpty(value))
            return null;
        locale = locale || this._locale;
        if (typeof display === 'boolean') {
            if (console && console.warn) {
                console.warn("Warning: the currency pipe has been changed in Angular v5. The symbolDisplay option (third parameter) is now a string instead of a boolean. The accepted values are \"code\", \"symbol\" or \"symbol-narrow\".");
            }
            display = display ? 'symbol' : 'code';
        }
        var currency = currencyCode || 'USD';
        if (display !== 'code') {
            if (display === 'symbol' || display === 'symbol-narrow') {
                currency = locale_data_api_1.getCurrencySymbol(currency, display === 'symbol' ? 'wide' : 'narrow', locale);
            }
            else {
                currency = display;
            }
        }
        try {
            var num = strToNumber(value);
            return format_number_1.formatCurrency(num, locale, currency, currencyCode, digitsInfo);
        }
        catch (error) {
            throw invalid_pipe_argument_error_1.invalidPipeArgumentError(CurrencyPipe, error.message);
        }
    };
    CurrencyPipe.decorators = [
        { type: core_1.Pipe, args: [{ name: 'currency' },] },
    ];
    /** @nocollapse */
    CurrencyPipe.ctorParameters = function () { return [
        { type: undefined, decorators: [{ type: core_1.Inject, args: [core_1.LOCALE_ID,] },] },
    ]; };
    return CurrencyPipe;
}());
exports.CurrencyPipe = CurrencyPipe;
function isEmpty(value) {
    return value == null || value === '' || value !== value;
}
/**
 * Transforms a string into a number (if needed)
 */
function strToNumber(value) {
    // Convert strings to numbers
    if (typeof value === 'string' && !isNaN(Number(value) - parseFloat(value))) {
        return Number(value);
    }
    if (typeof value !== 'number') {
        throw new Error(value + " is not a number");
    }
    return value;
}
//# sourceMappingURL=number_pipe.js.map