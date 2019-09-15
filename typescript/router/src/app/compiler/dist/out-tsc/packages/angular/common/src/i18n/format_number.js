"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var locale_data_api_1 = require("./locale_data_api");
exports.NUMBER_FORMAT_REGEXP = /^(\d+)?\.((\d+)(-(\d+))?)?$/;
var MAX_DIGITS = 22;
var DECIMAL_SEP = '.';
var ZERO_CHAR = '0';
var PATTERN_SEP = ';';
var GROUP_SEP = ',';
var DIGIT_CHAR = '#';
var CURRENCY_CHAR = '¤';
var PERCENT_CHAR = '%';
/**
 * Transforms a number to a locale string based on a style and a format
 */
function formatNumberToLocaleString(value, pattern, locale, groupSymbol, decimalSymbol, digitsInfo, isPercent) {
    if (isPercent === void 0) { isPercent = false; }
    var formattedText = '';
    var isZero = false;
    if (!isFinite(value)) {
        formattedText = locale_data_api_1.getLocaleNumberSymbol(locale, locale_data_api_1.NumberSymbol.Infinity);
    }
    else {
        var parsedNumber = parseNumber(value);
        if (isPercent) {
            parsedNumber = toPercent(parsedNumber);
        }
        var minInt = pattern.minInt;
        var minFraction = pattern.minFrac;
        var maxFraction = pattern.maxFrac;
        if (digitsInfo) {
            var parts = digitsInfo.match(exports.NUMBER_FORMAT_REGEXP);
            if (parts === null) {
                throw new Error(digitsInfo + " is not a valid digit info");
            }
            var minIntPart = parts[1];
            var minFractionPart = parts[3];
            var maxFractionPart = parts[5];
            if (minIntPart != null) {
                minInt = parseIntAutoRadix(minIntPart);
            }
            if (minFractionPart != null) {
                minFraction = parseIntAutoRadix(minFractionPart);
            }
            if (maxFractionPart != null) {
                maxFraction = parseIntAutoRadix(maxFractionPart);
            }
            else if (minFractionPart != null && minFraction > maxFraction) {
                maxFraction = minFraction;
            }
        }
        roundNumber(parsedNumber, minFraction, maxFraction);
        var digits = parsedNumber.digits;
        var integerLen = parsedNumber.integerLen;
        var exponent = parsedNumber.exponent;
        var decimals = [];
        isZero = digits.every(function (d) { return !d; });
        // pad zeros for small numbers
        for (; integerLen < minInt; integerLen++) {
            digits.unshift(0);
        }
        // pad zeros for small numbers
        for (; integerLen < 0; integerLen++) {
            digits.unshift(0);
        }
        // extract decimals digits
        if (integerLen > 0) {
            decimals = digits.splice(integerLen, digits.length);
        }
        else {
            decimals = digits;
            digits = [0];
        }
        // format the integer digits with grouping separators
        var groups = [];
        if (digits.length >= pattern.lgSize) {
            groups.unshift(digits.splice(-pattern.lgSize, digits.length).join(''));
        }
        while (digits.length > pattern.gSize) {
            groups.unshift(digits.splice(-pattern.gSize, digits.length).join(''));
        }
        if (digits.length) {
            groups.unshift(digits.join(''));
        }
        formattedText = groups.join(locale_data_api_1.getLocaleNumberSymbol(locale, groupSymbol));
        // append the decimal digits
        if (decimals.length) {
            formattedText += locale_data_api_1.getLocaleNumberSymbol(locale, decimalSymbol) + decimals.join('');
        }
        if (exponent) {
            formattedText += locale_data_api_1.getLocaleNumberSymbol(locale, locale_data_api_1.NumberSymbol.Exponential) + '+' + exponent;
        }
    }
    if (value < 0 && !isZero) {
        formattedText = pattern.negPre + formattedText + pattern.negSuf;
    }
    else {
        formattedText = pattern.posPre + formattedText + pattern.posSuf;
    }
    return formattedText;
}
/**
 * @ngModule CommonModule
 * @description
 *
 * Formats a number as currency using locale rules.
 *
 * Use `currency` to format a number as currency.
 *
 * Where:
 * - `value` is a number.
 * - `locale` is a `string` defining the locale to use.
 * - `currency` is the string that represents the currency, it can be its symbol or its name.
 * - `currencyCode` is the [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217) currency code, such
 *    as `USD` for the US dollar and `EUR` for the euro.
 * - `digitInfo` See {@link DecimalPipe} for more details.
 *
 *
 */
function formatCurrency(value, locale, currency, currencyCode, digitsInfo) {
    var format = locale_data_api_1.getLocaleNumberFormat(locale, locale_data_api_1.NumberFormatStyle.Currency);
    var pattern = parseNumberFormat(format, locale_data_api_1.getLocaleNumberSymbol(locale, locale_data_api_1.NumberSymbol.MinusSign));
    pattern.minFrac = locale_data_api_1.getNumberOfCurrencyDigits((currencyCode));
    pattern.maxFrac = pattern.minFrac;
    var res = formatNumberToLocaleString(value, pattern, locale, locale_data_api_1.NumberSymbol.CurrencyGroup, locale_data_api_1.NumberSymbol.CurrencyDecimal, digitsInfo);
    return res
        .replace(CURRENCY_CHAR, currency)
        .replace(CURRENCY_CHAR, '');
}
exports.formatCurrency = formatCurrency;
/**
 * @ngModule CommonModule
 * @description
 *
 * Formats a number as a percentage according to locale rules.
 *
 * Where:
 * - `value` is a number.
 * - `locale` is a `string` defining the locale to use.
 * - `digitInfo` See {@link DecimalPipe} for more details.
 *
 *
 */
function formatPercent(value, locale, digitsInfo) {
    var format = locale_data_api_1.getLocaleNumberFormat(locale, locale_data_api_1.NumberFormatStyle.Percent);
    var pattern = parseNumberFormat(format, locale_data_api_1.getLocaleNumberSymbol(locale, locale_data_api_1.NumberSymbol.MinusSign));
    var res = formatNumberToLocaleString(value, pattern, locale, locale_data_api_1.NumberSymbol.Group, locale_data_api_1.NumberSymbol.Decimal, digitsInfo, true);
    return res.replace(new RegExp(PERCENT_CHAR, 'g'), locale_data_api_1.getLocaleNumberSymbol(locale, locale_data_api_1.NumberSymbol.PercentSign));
}
exports.formatPercent = formatPercent;
/**
 * @ngModule CommonModule
 * @description
 *
 * Formats a number as text. Group sizing and separator and other locale-specific
 * configurations are based on the locale.
 *
 * Where:
 * - `value` is a number.
 * - `locale` is a `string` defining the locale to use.
 * - `digitInfo` See {@link DecimalPipe} for more details.
 *
 *
 */
function formatNumber(value, locale, digitsInfo) {
    var format = locale_data_api_1.getLocaleNumberFormat(locale, locale_data_api_1.NumberFormatStyle.Decimal);
    var pattern = parseNumberFormat(format, locale_data_api_1.getLocaleNumberSymbol(locale, locale_data_api_1.NumberSymbol.MinusSign));
    return formatNumberToLocaleString(value, pattern, locale, locale_data_api_1.NumberSymbol.Group, locale_data_api_1.NumberSymbol.Decimal, digitsInfo);
}
exports.formatNumber = formatNumber;
function parseNumberFormat(format, minusSign) {
    if (minusSign === void 0) { minusSign = '-'; }
    var p = {
        minInt: 1,
        minFrac: 0,
        maxFrac: 0,
        posPre: '',
        posSuf: '',
        negPre: '',
        negSuf: '',
        gSize: 0,
        lgSize: 0
    };
    var patternParts = format.split(PATTERN_SEP);
    var positive = patternParts[0];
    var negative = patternParts[1];
    var positiveParts = positive.indexOf(DECIMAL_SEP) !== -1 ?
        positive.split(DECIMAL_SEP) :
        [
            positive.substring(0, positive.lastIndexOf(ZERO_CHAR) + 1),
            positive.substring(positive.lastIndexOf(ZERO_CHAR) + 1)
        ], integer = positiveParts[0], fraction = positiveParts[1] || '';
    p.posPre = integer.substr(0, integer.indexOf(DIGIT_CHAR));
    for (var i = 0; i < fraction.length; i++) {
        var ch = fraction.charAt(i);
        if (ch === ZERO_CHAR) {
            p.minFrac = p.maxFrac = i + 1;
        }
        else if (ch === DIGIT_CHAR) {
            p.maxFrac = i + 1;
        }
        else {
            p.posSuf += ch;
        }
    }
    var groups = integer.split(GROUP_SEP);
    p.gSize = groups[1] ? groups[1].length : 0;
    p.lgSize = (groups[2] || groups[1]) ? (groups[2] || groups[1]).length : 0;
    if (negative) {
        var trunkLen = positive.length - p.posPre.length - p.posSuf.length, pos = negative.indexOf(DIGIT_CHAR);
        p.negPre = negative.substr(0, pos).replace(/'/g, '');
        p.negSuf = negative.substr(pos + trunkLen).replace(/'/g, '');
    }
    else {
        p.negPre = minusSign + p.posPre;
        p.negSuf = p.posSuf;
    }
    return p;
}
// Transforms a parsed number into a percentage by multiplying it by 100
function toPercent(parsedNumber) {
    // if the number is 0, don't do anything
    if (parsedNumber.digits[0] === 0) {
        return parsedNumber;
    }
    // Getting the current number of decimals
    var fractionLen = parsedNumber.digits.length - parsedNumber.integerLen;
    if (parsedNumber.exponent) {
        parsedNumber.exponent += 2;
    }
    else {
        if (fractionLen === 0) {
            parsedNumber.digits.push(0, 0);
        }
        else if (fractionLen === 1) {
            parsedNumber.digits.push(0);
        }
        parsedNumber.integerLen += 2;
    }
    return parsedNumber;
}
/**
 * Parses a number.
 * Significant bits of this parse algorithm came from https://github.com/MikeMcl/big.js/
 */
function parseNumber(num) {
    var numStr = Math.abs(num) + '';
    var exponent = 0, digits, integerLen;
    var i, j, zeros;
    // Decimal point?
    if ((integerLen = numStr.indexOf(DECIMAL_SEP)) > -1) {
        numStr = numStr.replace(DECIMAL_SEP, '');
    }
    // Exponential form?
    if ((i = numStr.search(/e/i)) > 0) {
        // Work out the exponent.
        if (integerLen < 0)
            integerLen = i;
        integerLen += +numStr.slice(i + 1);
        numStr = numStr.substring(0, i);
    }
    else if (integerLen < 0) {
        // There was no decimal point or exponent so it is an integer.
        integerLen = numStr.length;
    }
    // Count the number of leading zeros.
    for (i = 0; numStr.charAt(i) === ZERO_CHAR; i++) {
        /* empty */
    }
    if (i === (zeros = numStr.length)) {
        // The digits are all zero.
        digits = [0];
        integerLen = 1;
    }
    else {
        // Count the number of trailing zeros
        zeros--;
        while (numStr.charAt(zeros) === ZERO_CHAR)
            zeros--;
        // Trailing zeros are insignificant so ignore them
        integerLen -= i;
        digits = [];
        // Convert string to array of digits without leading/trailing zeros.
        for (j = 0; i <= zeros; i++, j++) {
            digits[j] = Number(numStr.charAt(i));
        }
    }
    // If the number overflows the maximum allowed digits then use an exponent.
    if (integerLen > MAX_DIGITS) {
        digits = digits.splice(0, MAX_DIGITS - 1);
        exponent = integerLen - 1;
        integerLen = 1;
    }
    return { digits: digits, exponent: exponent, integerLen: integerLen };
}
/**
 * Round the parsed number to the specified number of decimal places
 * This function changes the parsedNumber in-place
 */
function roundNumber(parsedNumber, minFrac, maxFrac) {
    if (minFrac > maxFrac) {
        throw new Error("The minimum number of digits after fraction (" + minFrac + ") is higher than the maximum (" + maxFrac + ").");
    }
    var digits = parsedNumber.digits;
    var fractionLen = digits.length - parsedNumber.integerLen;
    var fractionSize = Math.min(Math.max(minFrac, fractionLen), maxFrac);
    // The index of the digit to where rounding is to occur
    var roundAt = fractionSize + parsedNumber.integerLen;
    var digit = digits[roundAt];
    if (roundAt > 0) {
        // Drop fractional digits beyond `roundAt`
        digits.splice(Math.max(parsedNumber.integerLen, roundAt));
        // Set non-fractional digits beyond `roundAt` to 0
        for (var j = roundAt; j < digits.length; j++) {
            digits[j] = 0;
        }
    }
    else {
        // We rounded to zero so reset the parsedNumber
        fractionLen = Math.max(0, fractionLen);
        parsedNumber.integerLen = 1;
        digits.length = Math.max(1, roundAt = fractionSize + 1);
        digits[0] = 0;
        for (var i = 1; i < roundAt; i++)
            digits[i] = 0;
    }
    if (digit >= 5) {
        if (roundAt - 1 < 0) {
            for (var k = 0; k > roundAt; k--) {
                digits.unshift(0);
                parsedNumber.integerLen++;
            }
            digits.unshift(1);
            parsedNumber.integerLen++;
        }
        else {
            digits[roundAt - 1]++;
        }
    }
    // Pad out with zeros to get the required fraction length
    for (; fractionLen < Math.max(0, fractionSize); fractionLen++)
        digits.push(0);
    var dropTrailingZeros = fractionSize !== 0;
    // Minimal length = nb of decimals required + current nb of integers
    // Any number besides that is optional and can be removed if it's a trailing 0
    var minLen = minFrac + parsedNumber.integerLen;
    // Do any carrying, e.g. a digit was rounded up to 10
    var carry = digits.reduceRight(function (carry, d, i, digits) {
        d = d + carry;
        digits[i] = d < 10 ? d : d - 10; // d % 10
        if (dropTrailingZeros) {
            // Do not keep meaningless fractional trailing zeros (e.g. 15.52000 --> 15.52)
            if (digits[i] === 0 && i >= minLen) {
                digits.pop();
            }
            else {
                dropTrailingZeros = false;
            }
        }
        return d >= 10 ? 1 : 0; // Math.floor(d / 10);
    }, 0);
    if (carry) {
        digits.unshift(carry);
        parsedNumber.integerLen++;
    }
}
function parseIntAutoRadix(text) {
    var result = parseInt(text);
    if (isNaN(result)) {
        throw new Error('Invalid integer literal when parsing ' + text);
    }
    return result;
}
exports.parseIntAutoRadix = parseIntAutoRadix;
//# sourceMappingURL=format_number.js.map