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
var format_number_1 = require("../../i18n/format_number");
var locale_data_api_1 = require("../../i18n/locale_data_api");
var invalid_pipe_argument_error_1 = require("../invalid_pipe_argument_error");
var intl_1 = require("./intl");
function formatNumber(pipe, locale, value, style, digits, currency, currencyAsSymbol) {
    if (currency === void 0) { currency = null; }
    if (currencyAsSymbol === void 0) { currencyAsSymbol = false; }
    if (value == null)
        return null;
    // Convert strings to numbers
    value = typeof value === 'string' && !isNaN(+value - parseFloat(value)) ? +value : value;
    if (typeof value !== 'number') {
        throw invalid_pipe_argument_error_1.invalidPipeArgumentError(pipe, value);
    }
    var minInt;
    var minFraction;
    var maxFraction;
    if (style !== locale_data_api_1.NumberFormatStyle.Currency) {
        // rely on Intl default for currency
        minInt = 1;
        minFraction = 0;
        maxFraction = 3;
    }
    if (digits) {
        var parts = digits.match(format_number_1.NUMBER_FORMAT_REGEXP);
        if (parts === null) {
            throw new Error(digits + " is not a valid digit info for number pipes");
        }
        if (parts[1] != null) {
            // min integer digits
            minInt = format_number_1.parseIntAutoRadix(parts[1]);
        }
        if (parts[3] != null) {
            // min fraction digits
            minFraction = format_number_1.parseIntAutoRadix(parts[3]);
        }
        if (parts[5] != null) {
            // max fraction digits
            maxFraction = format_number_1.parseIntAutoRadix(parts[5]);
        }
    }
    return intl_1.NumberFormatter.format(value, locale, style, {
        minimumIntegerDigits: minInt,
        minimumFractionDigits: minFraction,
        maximumFractionDigits: maxFraction,
        currency: currency,
        currencyAsSymbol: currencyAsSymbol,
    });
}
/**
 * @ngModule CommonModule
 *
 * Formats a number as text. Group sizing and separator and other locale-specific
 * configurations are based on the active locale.
 *
 * where `expression` is a number:
 *  - `digitInfo` is a `string` which has a following format: <br>
 *     <code>{minIntegerDigits}.{minFractionDigits}-{maxFractionDigits}</code>
 *   - `minIntegerDigits` is the minimum number of integer digits to use. Defaults to `1`.
 *   - `minFractionDigits` is the minimum number of digits after fraction. Defaults to `0`.
 *   - `maxFractionDigits` is the maximum number of digits after fraction. Defaults to `3`.
 *
 * For more information on the acceptable range for each of these numbers and other
 * details see your native internationalization library.
 *
 * WARNING: this pipe uses the Internationalization API which is not yet available in all browsers
 * and may require a polyfill. See [Browser Support](guide/browser-support) for details.
 *
 * ### Example
 *
 * {@example common/pipes/ts/number_pipe.ts region='DeprecatedNumberPipe'}
 *
 *
 */
var DeprecatedDecimalPipe = /** @class */ (function () {
    function DeprecatedDecimalPipe(_locale) {
        this._locale = _locale;
    }
    DeprecatedDecimalPipe.prototype.transform = function (value, digits) {
        return formatNumber(DeprecatedDecimalPipe, this._locale, value, locale_data_api_1.NumberFormatStyle.Decimal, digits);
    };
    DeprecatedDecimalPipe.decorators = [
        { type: core_1.Pipe, args: [{ name: 'number' },] },
    ];
    /** @nocollapse */
    DeprecatedDecimalPipe.ctorParameters = function () { return [
        { type: undefined, decorators: [{ type: core_1.Inject, args: [core_1.LOCALE_ID,] },] },
    ]; };
    return DeprecatedDecimalPipe;
}());
exports.DeprecatedDecimalPipe = DeprecatedDecimalPipe;
/**
 * @ngModule CommonModule
 *
 * @description
 *
 * Formats a number as percentage according to locale rules.
 *
 * - `digitInfo` See {@link DecimalPipe} for detailed description.
 *
 * WARNING: this pipe uses the Internationalization API which is not yet available in all browsers
 * and may require a polyfill. See [Browser Support](guide/browser-support) for details.
 *
 * ### Example
 *
 * {@example common/pipes/ts/percent_pipe.ts region='DeprecatedPercentPipe'}
 *
 *
 */
var DeprecatedPercentPipe = /** @class */ (function () {
    function DeprecatedPercentPipe(_locale) {
        this._locale = _locale;
    }
    DeprecatedPercentPipe.prototype.transform = function (value, digits) {
        return formatNumber(DeprecatedPercentPipe, this._locale, value, locale_data_api_1.NumberFormatStyle.Percent, digits);
    };
    DeprecatedPercentPipe.decorators = [
        { type: core_1.Pipe, args: [{ name: 'percent' },] },
    ];
    /** @nocollapse */
    DeprecatedPercentPipe.ctorParameters = function () { return [
        { type: undefined, decorators: [{ type: core_1.Inject, args: [core_1.LOCALE_ID,] },] },
    ]; };
    return DeprecatedPercentPipe;
}());
exports.DeprecatedPercentPipe = DeprecatedPercentPipe;
/**
 * @ngModule CommonModule
 * @description
 *
 * Formats a number as currency using locale rules.
 *
 * Use `currency` to format a number as currency.
 *
 * - `currencyCode` is the [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217) currency code, such
 *    as `USD` for the US dollar and `EUR` for the euro.
 * - `symbolDisplay` is a boolean indicating whether to use the currency symbol or code.
 *   - `true`: use symbol (e.g. `$`).
 *   - `false`(default): use code (e.g. `USD`).
 * - `digitInfo` See {@link DecimalPipe} for detailed description.
 *
 * WARNING: this pipe uses the Internationalization API which is not yet available in all browsers
 * and may require a polyfill. See [Browser Support](guide/browser-support) for details.
 *
 * ### Example
 *
 * {@example common/pipes/ts/currency_pipe.ts region='DeprecatedCurrencyPipe'}
 *
 *
 */
var DeprecatedCurrencyPipe = /** @class */ (function () {
    function DeprecatedCurrencyPipe(_locale) {
        this._locale = _locale;
    }
    DeprecatedCurrencyPipe.prototype.transform = function (value, currencyCode, symbolDisplay, digits) {
        if (currencyCode === void 0) { currencyCode = 'USD'; }
        if (symbolDisplay === void 0) { symbolDisplay = false; }
        return formatNumber(DeprecatedCurrencyPipe, this._locale, value, locale_data_api_1.NumberFormatStyle.Currency, digits, currencyCode, symbolDisplay);
    };
    DeprecatedCurrencyPipe.decorators = [
        { type: core_1.Pipe, args: [{ name: 'currency' },] },
    ];
    /** @nocollapse */
    DeprecatedCurrencyPipe.ctorParameters = function () { return [
        { type: undefined, decorators: [{ type: core_1.Inject, args: [core_1.LOCALE_ID,] },] },
    ]; };
    return DeprecatedCurrencyPipe;
}());
exports.DeprecatedCurrencyPipe = DeprecatedCurrencyPipe;
//# sourceMappingURL=number_pipe.js.map