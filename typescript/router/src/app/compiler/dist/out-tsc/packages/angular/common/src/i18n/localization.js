"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = Object.setPrototypeOf ||
        ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
        function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var locale_data_api_1 = require("./locale_data_api");
/**
 * @deprecated from v5
 */
exports.DEPRECATED_PLURAL_FN = new core_1.InjectionToken('UseV4Plurals');
/**
 * @experimental
 */
var /**
 * @experimental
 */
NgLocalization = /** @class */ (function () {
    function NgLocalization() {
    }
    return NgLocalization;
}());
exports.NgLocalization = NgLocalization;
/**
 * Returns the plural category for a given value.
 * - "=value" when the case exists,
 * - the plural category otherwise
 */
function getPluralCategory(value, cases, ngLocalization, locale) {
    var key = "=" + value;
    if (cases.indexOf(key) > -1) {
        return key;
    }
    key = ngLocalization.getPluralCategory(value, locale);
    if (cases.indexOf(key) > -1) {
        return key;
    }
    if (cases.indexOf('other') > -1) {
        return 'other';
    }
    throw new Error("No plural message found for value \"" + value + "\"");
}
exports.getPluralCategory = getPluralCategory;
/**
 * Returns the plural case based on the locale
 *
 * @experimental
 */
var NgLocaleLocalization = /** @class */ (function (_super) {
    __extends(NgLocaleLocalization, _super);
    function NgLocaleLocalization(locale, /** @deprecated from v5 */
    deprecatedPluralFn) {
        var _this = _super.call(this) || this;
        _this.locale = locale;
        _this.deprecatedPluralFn = deprecatedPluralFn;
        return _this;
    }
    NgLocaleLocalization.prototype.getPluralCategory = function (value, locale) {
        var plural = this.deprecatedPluralFn ? this.deprecatedPluralFn(locale || this.locale, value) :
            locale_data_api_1.getLocalePluralCase(locale || this.locale)(value);
        switch (plural) {
            case locale_data_api_1.Plural.Zero:
                return 'zero';
            case locale_data_api_1.Plural.One:
                return 'one';
            case locale_data_api_1.Plural.Two:
                return 'two';
            case locale_data_api_1.Plural.Few:
                return 'few';
            case locale_data_api_1.Plural.Many:
                return 'many';
            default:
                return 'other';
        }
    };
    NgLocaleLocalization.decorators = [
        { type: core_1.Injectable },
    ];
    /** @nocollapse */
    NgLocaleLocalization.ctorParameters = function () { return [
        { type: undefined, decorators: [{ type: core_1.Inject, args: [core_1.LOCALE_ID,] },] },
        { type: undefined, decorators: [{ type: core_1.Optional }, { type: core_1.Inject, args: [exports.DEPRECATED_PLURAL_FN,] },] },
    ]; };
    return NgLocaleLocalization;
}(NgLocalization));
exports.NgLocaleLocalization = NgLocaleLocalization;
/**
 * Returns the plural case based on the locale
 *
 * @deprecated from v5 the plural case function is in locale data files common/locales/*.ts
 * @experimental
 */
function getPluralCase(locale, nLike) {
    // TODO(vicb): lazy compute
    if (typeof nLike === 'string') {
        nLike = parseInt(nLike, 10);
    }
    var n = nLike;
    var nDecimal = n.toString().replace(/^[^.]*\.?/, '');
    var i = Math.floor(Math.abs(n));
    var v = nDecimal.length;
    var f = parseInt(nDecimal, 10);
    var t = parseInt(n.toString().replace(/^[^.]*\.?|0+$/g, ''), 10) || 0;
    var lang = locale.split('-')[0].toLowerCase();
    switch (lang) {
        case 'af':
        case 'asa':
        case 'az':
        case 'bem':
        case 'bez':
        case 'bg':
        case 'brx':
        case 'ce':
        case 'cgg':
        case 'chr':
        case 'ckb':
        case 'ee':
        case 'el':
        case 'eo':
        case 'es':
        case 'eu':
        case 'fo':
        case 'fur':
        case 'gsw':
        case 'ha':
        case 'haw':
        case 'hu':
        case 'jgo':
        case 'jmc':
        case 'ka':
        case 'kk':
        case 'kkj':
        case 'kl':
        case 'ks':
        case 'ksb':
        case 'ky':
        case 'lb':
        case 'lg':
        case 'mas':
        case 'mgo':
        case 'ml':
        case 'mn':
        case 'nb':
        case 'nd':
        case 'ne':
        case 'nn':
        case 'nnh':
        case 'nyn':
        case 'om':
        case 'or':
        case 'os':
        case 'ps':
        case 'rm':
        case 'rof':
        case 'rwk':
        case 'saq':
        case 'seh':
        case 'sn':
        case 'so':
        case 'sq':
        case 'ta':
        case 'te':
        case 'teo':
        case 'tk':
        case 'tr':
        case 'ug':
        case 'uz':
        case 'vo':
        case 'vun':
        case 'wae':
        case 'xog':
            if (n === 1)
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'ak':
        case 'ln':
        case 'mg':
        case 'pa':
        case 'ti':
            if (n === Math.floor(n) && n >= 0 && n <= 1)
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'am':
        case 'as':
        case 'bn':
        case 'fa':
        case 'gu':
        case 'hi':
        case 'kn':
        case 'mr':
        case 'zu':
            if (i === 0 || n === 1)
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'ar':
            if (n === 0)
                return locale_data_api_1.Plural.Zero;
            if (n === 1)
                return locale_data_api_1.Plural.One;
            if (n === 2)
                return locale_data_api_1.Plural.Two;
            if (n % 100 === Math.floor(n % 100) && n % 100 >= 3 && n % 100 <= 10)
                return locale_data_api_1.Plural.Few;
            if (n % 100 === Math.floor(n % 100) && n % 100 >= 11 && n % 100 <= 99)
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'ast':
        case 'ca':
        case 'de':
        case 'en':
        case 'et':
        case 'fi':
        case 'fy':
        case 'gl':
        case 'it':
        case 'nl':
        case 'sv':
        case 'sw':
        case 'ur':
        case 'yi':
            if (i === 1 && v === 0)
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'be':
            if (n % 10 === 1 && !(n % 100 === 11))
                return locale_data_api_1.Plural.One;
            if (n % 10 === Math.floor(n % 10) && n % 10 >= 2 && n % 10 <= 4 &&
                !(n % 100 >= 12 && n % 100 <= 14))
                return locale_data_api_1.Plural.Few;
            if (n % 10 === 0 || n % 10 === Math.floor(n % 10) && n % 10 >= 5 && n % 10 <= 9 ||
                n % 100 === Math.floor(n % 100) && n % 100 >= 11 && n % 100 <= 14)
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'br':
            if (n % 10 === 1 && !(n % 100 === 11 || n % 100 === 71 || n % 100 === 91))
                return locale_data_api_1.Plural.One;
            if (n % 10 === 2 && !(n % 100 === 12 || n % 100 === 72 || n % 100 === 92))
                return locale_data_api_1.Plural.Two;
            if (n % 10 === Math.floor(n % 10) && (n % 10 >= 3 && n % 10 <= 4 || n % 10 === 9) &&
                !(n % 100 >= 10 && n % 100 <= 19 || n % 100 >= 70 && n % 100 <= 79 ||
                    n % 100 >= 90 && n % 100 <= 99))
                return locale_data_api_1.Plural.Few;
            if (!(n === 0) && n % 1e6 === 0)
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'bs':
        case 'hr':
        case 'sr':
            if (v === 0 && i % 10 === 1 && !(i % 100 === 11) || f % 10 === 1 && !(f % 100 === 11))
                return locale_data_api_1.Plural.One;
            if (v === 0 && i % 10 === Math.floor(i % 10) && i % 10 >= 2 && i % 10 <= 4 &&
                !(i % 100 >= 12 && i % 100 <= 14) ||
                f % 10 === Math.floor(f % 10) && f % 10 >= 2 && f % 10 <= 4 &&
                    !(f % 100 >= 12 && f % 100 <= 14))
                return locale_data_api_1.Plural.Few;
            return locale_data_api_1.Plural.Other;
        case 'cs':
        case 'sk':
            if (i === 1 && v === 0)
                return locale_data_api_1.Plural.One;
            if (i === Math.floor(i) && i >= 2 && i <= 4 && v === 0)
                return locale_data_api_1.Plural.Few;
            if (!(v === 0))
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'cy':
            if (n === 0)
                return locale_data_api_1.Plural.Zero;
            if (n === 1)
                return locale_data_api_1.Plural.One;
            if (n === 2)
                return locale_data_api_1.Plural.Two;
            if (n === 3)
                return locale_data_api_1.Plural.Few;
            if (n === 6)
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'da':
            if (n === 1 || !(t === 0) && (i === 0 || i === 1))
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'dsb':
        case 'hsb':
            if (v === 0 && i % 100 === 1 || f % 100 === 1)
                return locale_data_api_1.Plural.One;
            if (v === 0 && i % 100 === 2 || f % 100 === 2)
                return locale_data_api_1.Plural.Two;
            if (v === 0 && i % 100 === Math.floor(i % 100) && i % 100 >= 3 && i % 100 <= 4 ||
                f % 100 === Math.floor(f % 100) && f % 100 >= 3 && f % 100 <= 4)
                return locale_data_api_1.Plural.Few;
            return locale_data_api_1.Plural.Other;
        case 'ff':
        case 'fr':
        case 'hy':
        case 'kab':
            if (i === 0 || i === 1)
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'fil':
            if (v === 0 && (i === 1 || i === 2 || i === 3) ||
                v === 0 && !(i % 10 === 4 || i % 10 === 6 || i % 10 === 9) ||
                !(v === 0) && !(f % 10 === 4 || f % 10 === 6 || f % 10 === 9))
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'ga':
            if (n === 1)
                return locale_data_api_1.Plural.One;
            if (n === 2)
                return locale_data_api_1.Plural.Two;
            if (n === Math.floor(n) && n >= 3 && n <= 6)
                return locale_data_api_1.Plural.Few;
            if (n === Math.floor(n) && n >= 7 && n <= 10)
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'gd':
            if (n === 1 || n === 11)
                return locale_data_api_1.Plural.One;
            if (n === 2 || n === 12)
                return locale_data_api_1.Plural.Two;
            if (n === Math.floor(n) && (n >= 3 && n <= 10 || n >= 13 && n <= 19))
                return locale_data_api_1.Plural.Few;
            return locale_data_api_1.Plural.Other;
        case 'gv':
            if (v === 0 && i % 10 === 1)
                return locale_data_api_1.Plural.One;
            if (v === 0 && i % 10 === 2)
                return locale_data_api_1.Plural.Two;
            if (v === 0 &&
                (i % 100 === 0 || i % 100 === 20 || i % 100 === 40 || i % 100 === 60 || i % 100 === 80))
                return locale_data_api_1.Plural.Few;
            if (!(v === 0))
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'he':
            if (i === 1 && v === 0)
                return locale_data_api_1.Plural.One;
            if (i === 2 && v === 0)
                return locale_data_api_1.Plural.Two;
            if (v === 0 && !(n >= 0 && n <= 10) && n % 10 === 0)
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'is':
            if (t === 0 && i % 10 === 1 && !(i % 100 === 11) || !(t === 0))
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'ksh':
            if (n === 0)
                return locale_data_api_1.Plural.Zero;
            if (n === 1)
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'kw':
        case 'naq':
        case 'se':
        case 'smn':
            if (n === 1)
                return locale_data_api_1.Plural.One;
            if (n === 2)
                return locale_data_api_1.Plural.Two;
            return locale_data_api_1.Plural.Other;
        case 'lag':
            if (n === 0)
                return locale_data_api_1.Plural.Zero;
            if ((i === 0 || i === 1) && !(n === 0))
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'lt':
            if (n % 10 === 1 && !(n % 100 >= 11 && n % 100 <= 19))
                return locale_data_api_1.Plural.One;
            if (n % 10 === Math.floor(n % 10) && n % 10 >= 2 && n % 10 <= 9 &&
                !(n % 100 >= 11 && n % 100 <= 19))
                return locale_data_api_1.Plural.Few;
            if (!(f === 0))
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'lv':
        case 'prg':
            if (n % 10 === 0 || n % 100 === Math.floor(n % 100) && n % 100 >= 11 && n % 100 <= 19 ||
                v === 2 && f % 100 === Math.floor(f % 100) && f % 100 >= 11 && f % 100 <= 19)
                return locale_data_api_1.Plural.Zero;
            if (n % 10 === 1 && !(n % 100 === 11) || v === 2 && f % 10 === 1 && !(f % 100 === 11) ||
                !(v === 2) && f % 10 === 1)
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'mk':
            if (v === 0 && i % 10 === 1 || f % 10 === 1)
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'mt':
            if (n === 1)
                return locale_data_api_1.Plural.One;
            if (n === 0 || n % 100 === Math.floor(n % 100) && n % 100 >= 2 && n % 100 <= 10)
                return locale_data_api_1.Plural.Few;
            if (n % 100 === Math.floor(n % 100) && n % 100 >= 11 && n % 100 <= 19)
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'pl':
            if (i === 1 && v === 0)
                return locale_data_api_1.Plural.One;
            if (v === 0 && i % 10 === Math.floor(i % 10) && i % 10 >= 2 && i % 10 <= 4 &&
                !(i % 100 >= 12 && i % 100 <= 14))
                return locale_data_api_1.Plural.Few;
            if (v === 0 && !(i === 1) && i % 10 === Math.floor(i % 10) && i % 10 >= 0 && i % 10 <= 1 ||
                v === 0 && i % 10 === Math.floor(i % 10) && i % 10 >= 5 && i % 10 <= 9 ||
                v === 0 && i % 100 === Math.floor(i % 100) && i % 100 >= 12 && i % 100 <= 14)
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'pt':
            if (n === Math.floor(n) && n >= 0 && n <= 2 && !(n === 2))
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'ro':
            if (i === 1 && v === 0)
                return locale_data_api_1.Plural.One;
            if (!(v === 0) || n === 0 ||
                !(n === 1) && n % 100 === Math.floor(n % 100) && n % 100 >= 1 && n % 100 <= 19)
                return locale_data_api_1.Plural.Few;
            return locale_data_api_1.Plural.Other;
        case 'ru':
        case 'uk':
            if (v === 0 && i % 10 === 1 && !(i % 100 === 11))
                return locale_data_api_1.Plural.One;
            if (v === 0 && i % 10 === Math.floor(i % 10) && i % 10 >= 2 && i % 10 <= 4 &&
                !(i % 100 >= 12 && i % 100 <= 14))
                return locale_data_api_1.Plural.Few;
            if (v === 0 && i % 10 === 0 ||
                v === 0 && i % 10 === Math.floor(i % 10) && i % 10 >= 5 && i % 10 <= 9 ||
                v === 0 && i % 100 === Math.floor(i % 100) && i % 100 >= 11 && i % 100 <= 14)
                return locale_data_api_1.Plural.Many;
            return locale_data_api_1.Plural.Other;
        case 'shi':
            if (i === 0 || n === 1)
                return locale_data_api_1.Plural.One;
            if (n === Math.floor(n) && n >= 2 && n <= 10)
                return locale_data_api_1.Plural.Few;
            return locale_data_api_1.Plural.Other;
        case 'si':
            if (n === 0 || n === 1 || i === 0 && f === 1)
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        case 'sl':
            if (v === 0 && i % 100 === 1)
                return locale_data_api_1.Plural.One;
            if (v === 0 && i % 100 === 2)
                return locale_data_api_1.Plural.Two;
            if (v === 0 && i % 100 === Math.floor(i % 100) && i % 100 >= 3 && i % 100 <= 4 || !(v === 0))
                return locale_data_api_1.Plural.Few;
            return locale_data_api_1.Plural.Other;
        case 'tzm':
            if (n === Math.floor(n) && n >= 0 && n <= 1 || n === Math.floor(n) && n >= 11 && n <= 99)
                return locale_data_api_1.Plural.One;
            return locale_data_api_1.Plural.Other;
        // When there is no specification, the default is always "other"
        // Spec: http://cldr.unicode.org/index/cldr-spec/plural-rules
        // > other (required—general plural form — also used if the language only has a single form)
        default:
            return locale_data_api_1.Plural.Other;
    }
}
exports.getPluralCase = getPluralCase;
//# sourceMappingURL=localization.js.map