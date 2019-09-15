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
exports.ISO8601_DATE_REGEX = /^(\d{4})-?(\d\d)-?(\d\d)(?:T(\d\d)(?::?(\d\d)(?::?(\d\d)(?:\.(\d+))?)?)?(Z|([+-])(\d\d):?(\d\d))?)?$/;
//    1        2       3         4          5          6          7          8  9     10      11
var NAMED_FORMATS = {};
var DATE_FORMATS_SPLIT = /((?:[^GyMLwWdEabBhHmsSzZO']+)|(?:'(?:[^']|'')*')|(?:G{1,5}|y{1,4}|M{1,5}|L{1,5}|w{1,2}|W{1}|d{1,2}|E{1,6}|a{1,5}|b{1,5}|B{1,5}|h{1,2}|H{1,2}|m{1,2}|s{1,2}|S{1,3}|z{1,4}|Z{1,5}|O{1,4}))([\s\S]*)/;
var ZoneWidth;
(function (ZoneWidth) {
    ZoneWidth[ZoneWidth["Short"] = 0] = "Short";
    ZoneWidth[ZoneWidth["ShortGMT"] = 1] = "ShortGMT";
    ZoneWidth[ZoneWidth["Long"] = 2] = "Long";
    ZoneWidth[ZoneWidth["Extended"] = 3] = "Extended";
})(ZoneWidth || (ZoneWidth = {}));
var DateType;
(function (DateType) {
    DateType[DateType["FullYear"] = 0] = "FullYear";
    DateType[DateType["Month"] = 1] = "Month";
    DateType[DateType["Date"] = 2] = "Date";
    DateType[DateType["Hours"] = 3] = "Hours";
    DateType[DateType["Minutes"] = 4] = "Minutes";
    DateType[DateType["Seconds"] = 5] = "Seconds";
    DateType[DateType["Milliseconds"] = 6] = "Milliseconds";
    DateType[DateType["Day"] = 7] = "Day";
})(DateType || (DateType = {}));
var TranslationType;
(function (TranslationType) {
    TranslationType[TranslationType["DayPeriods"] = 0] = "DayPeriods";
    TranslationType[TranslationType["Days"] = 1] = "Days";
    TranslationType[TranslationType["Months"] = 2] = "Months";
    TranslationType[TranslationType["Eras"] = 3] = "Eras";
})(TranslationType || (TranslationType = {}));
/**
 * @ngModule CommonModule
 * @description
 *
 * Formats a date according to locale rules.
 *
 * Where:
 * - `value` is a Date, a number (milliseconds since UTC epoch) or an ISO string
 *   (https://www.w3.org/TR/NOTE-datetime).
 * - `format` indicates which date/time components to include. See {@link DatePipe} for more
 *   details.
 * - `locale` is a `string` defining the locale to use.
 * - `timezone` to be used for formatting. It understands UTC/GMT and the continental US time zone
 *   abbreviations, but for general use, use a time zone offset (e.g. `'+0430'`).
 *   If not specified, host system settings are used.
 *
 * See {@link DatePipe} for more details.
 *
 *
 */
function formatDate(value, format, locale, timezone) {
    var date = toDate(value);
    var namedFormat = getNamedFormat(locale, format);
    format = namedFormat || format;
    var parts = [];
    var match;
    while (format) {
        match = DATE_FORMATS_SPLIT.exec(format);
        if (match) {
            parts = parts.concat(match.slice(1));
            var part = parts.pop();
            if (!part) {
                break;
            }
            format = part;
        }
        else {
            parts.push(format);
            break;
        }
    }
    var dateTimezoneOffset = date.getTimezoneOffset();
    if (timezone) {
        dateTimezoneOffset = timezoneToOffset(timezone, dateTimezoneOffset);
        date = convertTimezoneToLocal(date, timezone, true);
    }
    var text = '';
    parts.forEach(function (value) {
        var dateFormatter = getDateFormatter(value);
        text += dateFormatter ?
            dateFormatter(date, locale, dateTimezoneOffset) :
            value === '\'\'' ? '\'' : value.replace(/(^'|'$)/g, '').replace(/''/g, '\'');
    });
    return text;
}
exports.formatDate = formatDate;
function getNamedFormat(locale, format) {
    var localeId = locale_data_api_1.getLocaleId(locale);
    NAMED_FORMATS[localeId] = NAMED_FORMATS[localeId] || {};
    if (NAMED_FORMATS[localeId][format]) {
        return NAMED_FORMATS[localeId][format];
    }
    var formatValue = '';
    switch (format) {
        case 'shortDate':
            formatValue = locale_data_api_1.getLocaleDateFormat(locale, locale_data_api_1.FormatWidth.Short);
            break;
        case 'mediumDate':
            formatValue = locale_data_api_1.getLocaleDateFormat(locale, locale_data_api_1.FormatWidth.Medium);
            break;
        case 'longDate':
            formatValue = locale_data_api_1.getLocaleDateFormat(locale, locale_data_api_1.FormatWidth.Long);
            break;
        case 'fullDate':
            formatValue = locale_data_api_1.getLocaleDateFormat(locale, locale_data_api_1.FormatWidth.Full);
            break;
        case 'shortTime':
            formatValue = locale_data_api_1.getLocaleTimeFormat(locale, locale_data_api_1.FormatWidth.Short);
            break;
        case 'mediumTime':
            formatValue = locale_data_api_1.getLocaleTimeFormat(locale, locale_data_api_1.FormatWidth.Medium);
            break;
        case 'longTime':
            formatValue = locale_data_api_1.getLocaleTimeFormat(locale, locale_data_api_1.FormatWidth.Long);
            break;
        case 'fullTime':
            formatValue = locale_data_api_1.getLocaleTimeFormat(locale, locale_data_api_1.FormatWidth.Full);
            break;
        case 'short':
            var shortTime = getNamedFormat(locale, 'shortTime');
            var shortDate = getNamedFormat(locale, 'shortDate');
            formatValue = formatDateTime(locale_data_api_1.getLocaleDateTimeFormat(locale, locale_data_api_1.FormatWidth.Short), [shortTime, shortDate]);
            break;
        case 'medium':
            var mediumTime = getNamedFormat(locale, 'mediumTime');
            var mediumDate = getNamedFormat(locale, 'mediumDate');
            formatValue = formatDateTime(locale_data_api_1.getLocaleDateTimeFormat(locale, locale_data_api_1.FormatWidth.Medium), [mediumTime, mediumDate]);
            break;
        case 'long':
            var longTime = getNamedFormat(locale, 'longTime');
            var longDate = getNamedFormat(locale, 'longDate');
            formatValue =
                formatDateTime(locale_data_api_1.getLocaleDateTimeFormat(locale, locale_data_api_1.FormatWidth.Long), [longTime, longDate]);
            break;
        case 'full':
            var fullTime = getNamedFormat(locale, 'fullTime');
            var fullDate = getNamedFormat(locale, 'fullDate');
            formatValue =
                formatDateTime(locale_data_api_1.getLocaleDateTimeFormat(locale, locale_data_api_1.FormatWidth.Full), [fullTime, fullDate]);
            break;
    }
    if (formatValue) {
        NAMED_FORMATS[localeId][format] = formatValue;
    }
    return formatValue;
}
function formatDateTime(str, opt_values) {
    if (opt_values) {
        str = str.replace(/\{([^}]+)}/g, function (match, key) {
            return (opt_values != null && key in opt_values) ? opt_values[key] : match;
        });
    }
    return str;
}
function padNumber(num, digits, minusSign, trim, negWrap) {
    if (minusSign === void 0) { minusSign = '-'; }
    var neg = '';
    if (num < 0 || (negWrap && num <= 0)) {
        if (negWrap) {
            num = -num + 1;
        }
        else {
            num = -num;
            neg = minusSign;
        }
    }
    var strNum = String(num);
    while (strNum.length < digits) {
        strNum = '0' + strNum;
    }
    if (trim) {
        strNum = strNum.substr(strNum.length - digits);
    }
    return neg + strNum;
}
/**
 * Returns a date formatter that transforms a date into its locale digit representation
 */
function dateGetter(name, size, offset, trim, negWrap) {
    if (offset === void 0) { offset = 0; }
    if (trim === void 0) { trim = false; }
    if (negWrap === void 0) { negWrap = false; }
    return function (date, locale) {
        var part = getDatePart(name, date, size);
        if (offset > 0 || part > -offset) {
            part += offset;
        }
        if (name === DateType.Hours && part === 0 && offset === -12) {
            part = 12;
        }
        return padNumber(part, size, locale_data_api_1.getLocaleNumberSymbol(locale, locale_data_api_1.NumberSymbol.MinusSign), trim, negWrap);
    };
}
function getDatePart(name, date, size) {
    switch (name) {
        case DateType.FullYear:
            return date.getFullYear();
        case DateType.Month:
            return date.getMonth();
        case DateType.Date:
            return date.getDate();
        case DateType.Hours:
            return date.getHours();
        case DateType.Minutes:
            return date.getMinutes();
        case DateType.Seconds:
            return date.getSeconds();
        case DateType.Milliseconds:
            var div = size === 1 ? 100 : (size === 2 ? 10 : 1);
            return Math.round(date.getMilliseconds() / div);
        case DateType.Day:
            return date.getDay();
        default:
            throw new Error("Unknown DateType value \"" + name + "\".");
    }
}
/**
 * Returns a date formatter that transforms a date into its locale string representation
 */
function dateStrGetter(name, width, form, extended) {
    if (form === void 0) { form = locale_data_api_1.FormStyle.Format; }
    if (extended === void 0) { extended = false; }
    return function (date, locale) {
        return getDateTranslation(date, locale, name, width, form, extended);
    };
}
/**
 * Returns the locale translation of a date for a given form, type and width
 */
function getDateTranslation(date, locale, name, width, form, extended) {
    switch (name) {
        case TranslationType.Months:
            return locale_data_api_1.getLocaleMonthNames(locale, form, width)[date.getMonth()];
        case TranslationType.Days:
            return locale_data_api_1.getLocaleDayNames(locale, form, width)[date.getDay()];
        case TranslationType.DayPeriods:
            var currentHours_1 = date.getHours();
            var currentMinutes_1 = date.getMinutes();
            if (extended) {
                var rules = locale_data_api_1.getLocaleExtraDayPeriodRules(locale);
                var dayPeriods_1 = locale_data_api_1.getLocaleExtraDayPeriods(locale, form, width);
                var result_1;
                rules.forEach(function (rule, index) {
                    if (Array.isArray(rule)) {
                        // morning, afternoon, evening, night
                        var _a = rule[0], hoursFrom = _a.hours, minutesFrom = _a.minutes;
                        var _b = rule[1], hoursTo = _b.hours, minutesTo = _b.minutes;
                        if (currentHours_1 >= hoursFrom && currentMinutes_1 >= minutesFrom &&
                            (currentHours_1 < hoursTo ||
                                (currentHours_1 === hoursTo && currentMinutes_1 < minutesTo))) {
                            result_1 = dayPeriods_1[index];
                        }
                    }
                    else {
                        // noon or midnight
                        var hours = rule.hours, minutes = rule.minutes;
                        if (hours === currentHours_1 && minutes === currentMinutes_1) {
                            result_1 = dayPeriods_1[index];
                        }
                    }
                });
                if (result_1) {
                    return result_1;
                }
            }
            // if no rules for the day periods, we use am/pm by default
            return locale_data_api_1.getLocaleDayPeriods(locale, form, width)[currentHours_1 < 12 ? 0 : 1];
        case TranslationType.Eras:
            return locale_data_api_1.getLocaleEraNames(locale, width)[date.getFullYear() <= 0 ? 0 : 1];
        default:
            // This default case is not needed by TypeScript compiler, as the switch is exhaustive.
            // However Closure Compiler does not understand that and reports an error in typed mode.
            // The `throw new Error` below works around the problem, and the unexpected: never variable
            // makes sure tsc still checks this code is unreachable.
            var unexpected = name;
            throw new Error("unexpected translation type " + unexpected);
    }
}
/**
 * Returns a date formatter that transforms a date and an offset into a timezone with ISO8601 or
 * GMT format depending on the width (eg: short = +0430, short:GMT = GMT+4, long = GMT+04:30,
 * extended = +04:30)
 */
function timeZoneGetter(width) {
    return function (date, locale, offset) {
        var zone = -1 * offset;
        var minusSign = locale_data_api_1.getLocaleNumberSymbol(locale, locale_data_api_1.NumberSymbol.MinusSign);
        var hours = zone > 0 ? Math.floor(zone / 60) : Math.ceil(zone / 60);
        switch (width) {
            case ZoneWidth.Short:
                return ((zone >= 0) ? '+' : '') + padNumber(hours, 2, minusSign) +
                    padNumber(Math.abs(zone % 60), 2, minusSign);
            case ZoneWidth.ShortGMT:
                return 'GMT' + ((zone >= 0) ? '+' : '') + padNumber(hours, 1, minusSign);
            case ZoneWidth.Long:
                return 'GMT' + ((zone >= 0) ? '+' : '') + padNumber(hours, 2, minusSign) + ':' +
                    padNumber(Math.abs(zone % 60), 2, minusSign);
            case ZoneWidth.Extended:
                if (offset === 0) {
                    return 'Z';
                }
                else {
                    return ((zone >= 0) ? '+' : '') + padNumber(hours, 2, minusSign) + ':' +
                        padNumber(Math.abs(zone % 60), 2, minusSign);
                }
            default:
                throw new Error("Unknown zone width \"" + width + "\"");
        }
    };
}
var JANUARY = 0;
var THURSDAY = 4;
function getFirstThursdayOfYear(year) {
    var firstDayOfYear = (new Date(year, JANUARY, 1)).getDay();
    return new Date(year, 0, 1 + ((firstDayOfYear <= THURSDAY) ? THURSDAY : THURSDAY + 7) - firstDayOfYear);
}
function getThursdayThisWeek(datetime) {
    return new Date(datetime.getFullYear(), datetime.getMonth(), datetime.getDate() + (THURSDAY - datetime.getDay()));
}
function weekGetter(size, monthBased) {
    if (monthBased === void 0) { monthBased = false; }
    return function (date, locale) {
        var result;
        if (monthBased) {
            var nbDaysBefore1stDayOfMonth = new Date(date.getFullYear(), date.getMonth(), 1).getDay() - 1;
            var today = date.getDate();
            result = 1 + Math.floor((today + nbDaysBefore1stDayOfMonth) / 7);
        }
        else {
            var firstThurs = getFirstThursdayOfYear(date.getFullYear());
            var thisThurs = getThursdayThisWeek(date);
            var diff = thisThurs.getTime() - firstThurs.getTime();
            result = 1 + Math.round(diff / 6.048e8); // 6.048e8 ms per week
        }
        return padNumber(result, size, locale_data_api_1.getLocaleNumberSymbol(locale, locale_data_api_1.NumberSymbol.MinusSign));
    };
}
var DATE_FORMATS = {};
// Based on CLDR formats:
// See complete list: http://www.unicode.org/reports/tr35/tr35-dates.html#Date_Field_Symbol_Table
// See also explanations: http://cldr.unicode.org/translation/date-time
// TODO(ocombe): support all missing cldr formats: Y, U, Q, D, F, e, c, j, J, C, A, v, V, X, x
function getDateFormatter(format) {
    if (DATE_FORMATS[format]) {
        return DATE_FORMATS[format];
    }
    var formatter;
    switch (format) {
        // Era name (AD/BC)
        case 'G':
        case 'GG':
        case 'GGG':
            formatter = dateStrGetter(TranslationType.Eras, locale_data_api_1.TranslationWidth.Abbreviated);
            break;
        case 'GGGG':
            formatter = dateStrGetter(TranslationType.Eras, locale_data_api_1.TranslationWidth.Wide);
            break;
        case 'GGGGG':
            formatter = dateStrGetter(TranslationType.Eras, locale_data_api_1.TranslationWidth.Narrow);
            break;
        // 1 digit representation of the year, e.g. (AD 1 => 1, AD 199 => 199)
        case 'y':
            formatter = dateGetter(DateType.FullYear, 1, 0, false, true);
            break;
        // 2 digit representation of the year, padded (00-99). (e.g. AD 2001 => 01, AD 2010 => 10)
        case 'yy':
            formatter = dateGetter(DateType.FullYear, 2, 0, true, true);
            break;
        // 3 digit representation of the year, padded (000-999). (e.g. AD 2001 => 01, AD 2010 => 10)
        case 'yyy':
            formatter = dateGetter(DateType.FullYear, 3, 0, false, true);
            break;
        // 4 digit representation of the year (e.g. AD 1 => 0001, AD 2010 => 2010)
        case 'yyyy':
            formatter = dateGetter(DateType.FullYear, 4, 0, false, true);
            break;
        // Month of the year (1-12), numeric
        case 'M':
        case 'L':
            formatter = dateGetter(DateType.Month, 1, 1);
            break;
        case 'MM':
        case 'LL':
            formatter = dateGetter(DateType.Month, 2, 1);
            break;
        // Month of the year (January, ...), string, format
        case 'MMM':
            formatter = dateStrGetter(TranslationType.Months, locale_data_api_1.TranslationWidth.Abbreviated);
            break;
        case 'MMMM':
            formatter = dateStrGetter(TranslationType.Months, locale_data_api_1.TranslationWidth.Wide);
            break;
        case 'MMMMM':
            formatter = dateStrGetter(TranslationType.Months, locale_data_api_1.TranslationWidth.Narrow);
            break;
        // Month of the year (January, ...), string, standalone
        case 'LLL':
            formatter =
                dateStrGetter(TranslationType.Months, locale_data_api_1.TranslationWidth.Abbreviated, locale_data_api_1.FormStyle.Standalone);
            break;
        case 'LLLL':
            formatter =
                dateStrGetter(TranslationType.Months, locale_data_api_1.TranslationWidth.Wide, locale_data_api_1.FormStyle.Standalone);
            break;
        case 'LLLLL':
            formatter =
                dateStrGetter(TranslationType.Months, locale_data_api_1.TranslationWidth.Narrow, locale_data_api_1.FormStyle.Standalone);
            break;
        // Week of the year (1, ... 52)
        case 'w':
            formatter = weekGetter(1);
            break;
        case 'ww':
            formatter = weekGetter(2);
            break;
        // Week of the month (1, ...)
        case 'W':
            formatter = weekGetter(1, true);
            break;
        // Day of the month (1-31)
        case 'd':
            formatter = dateGetter(DateType.Date, 1);
            break;
        case 'dd':
            formatter = dateGetter(DateType.Date, 2);
            break;
        // Day of the Week
        case 'E':
        case 'EE':
        case 'EEE':
            formatter = dateStrGetter(TranslationType.Days, locale_data_api_1.TranslationWidth.Abbreviated);
            break;
        case 'EEEE':
            formatter = dateStrGetter(TranslationType.Days, locale_data_api_1.TranslationWidth.Wide);
            break;
        case 'EEEEE':
            formatter = dateStrGetter(TranslationType.Days, locale_data_api_1.TranslationWidth.Narrow);
            break;
        case 'EEEEEE':
            formatter = dateStrGetter(TranslationType.Days, locale_data_api_1.TranslationWidth.Short);
            break;
        // Generic period of the day (am-pm)
        case 'a':
        case 'aa':
        case 'aaa':
            formatter = dateStrGetter(TranslationType.DayPeriods, locale_data_api_1.TranslationWidth.Abbreviated);
            break;
        case 'aaaa':
            formatter = dateStrGetter(TranslationType.DayPeriods, locale_data_api_1.TranslationWidth.Wide);
            break;
        case 'aaaaa':
            formatter = dateStrGetter(TranslationType.DayPeriods, locale_data_api_1.TranslationWidth.Narrow);
            break;
        // Extended period of the day (midnight, at night, ...), standalone
        case 'b':
        case 'bb':
        case 'bbb':
            formatter = dateStrGetter(TranslationType.DayPeriods, locale_data_api_1.TranslationWidth.Abbreviated, locale_data_api_1.FormStyle.Standalone, true);
            break;
        case 'bbbb':
            formatter = dateStrGetter(TranslationType.DayPeriods, locale_data_api_1.TranslationWidth.Wide, locale_data_api_1.FormStyle.Standalone, true);
            break;
        case 'bbbbb':
            formatter = dateStrGetter(TranslationType.DayPeriods, locale_data_api_1.TranslationWidth.Narrow, locale_data_api_1.FormStyle.Standalone, true);
            break;
        // Extended period of the day (midnight, night, ...), standalone
        case 'B':
        case 'BB':
        case 'BBB':
            formatter = dateStrGetter(TranslationType.DayPeriods, locale_data_api_1.TranslationWidth.Abbreviated, locale_data_api_1.FormStyle.Format, true);
            break;
        case 'BBBB':
            formatter =
                dateStrGetter(TranslationType.DayPeriods, locale_data_api_1.TranslationWidth.Wide, locale_data_api_1.FormStyle.Format, true);
            break;
        case 'BBBBB':
            formatter = dateStrGetter(TranslationType.DayPeriods, locale_data_api_1.TranslationWidth.Narrow, locale_data_api_1.FormStyle.Format, true);
            break;
        // Hour in AM/PM, (1-12)
        case 'h':
            formatter = dateGetter(DateType.Hours, 1, -12);
            break;
        case 'hh':
            formatter = dateGetter(DateType.Hours, 2, -12);
            break;
        // Hour of the day (0-23)
        case 'H':
            formatter = dateGetter(DateType.Hours, 1);
            break;
        // Hour in day, padded (00-23)
        case 'HH':
            formatter = dateGetter(DateType.Hours, 2);
            break;
        // Minute of the hour (0-59)
        case 'm':
            formatter = dateGetter(DateType.Minutes, 1);
            break;
        case 'mm':
            formatter = dateGetter(DateType.Minutes, 2);
            break;
        // Second of the minute (0-59)
        case 's':
            formatter = dateGetter(DateType.Seconds, 1);
            break;
        case 'ss':
            formatter = dateGetter(DateType.Seconds, 2);
            break;
        // Fractional second padded (0-9)
        case 'S':
            formatter = dateGetter(DateType.Milliseconds, 1);
            break;
        case 'SS':
            formatter = dateGetter(DateType.Milliseconds, 2);
            break;
        // = millisecond
        case 'SSS':
            formatter = dateGetter(DateType.Milliseconds, 3);
            break;
        // Timezone ISO8601 short format (-0430)
        case 'Z':
        case 'ZZ':
        case 'ZZZ':
            formatter = timeZoneGetter(ZoneWidth.Short);
            break;
        // Timezone ISO8601 extended format (-04:30)
        case 'ZZZZZ':
            formatter = timeZoneGetter(ZoneWidth.Extended);
            break;
        // Timezone GMT short format (GMT+4)
        case 'O':
        case 'OO':
        case 'OOO':
        // Should be location, but fallback to format O instead because we don't have the data yet
        case 'z':
        case 'zz':
        case 'zzz':
            formatter = timeZoneGetter(ZoneWidth.ShortGMT);
            break;
        // Timezone GMT long format (GMT+0430)
        case 'OOOO':
        case 'ZZZZ':
        // Should be location, but fallback to format O instead because we don't have the data yet
        case 'zzzz':
            formatter = timeZoneGetter(ZoneWidth.Long);
            break;
        default:
            return null;
    }
    DATE_FORMATS[format] = formatter;
    return formatter;
}
function timezoneToOffset(timezone, fallback) {
    // Support: IE 9-11 only, Edge 13-15+
    // IE/Edge do not "understand" colon (`:`) in timezone
    timezone = timezone.replace(/:/g, '');
    var requestedTimezoneOffset = Date.parse('Jan 01, 1970 00:00:00 ' + timezone) / 60000;
    return isNaN(requestedTimezoneOffset) ? fallback : requestedTimezoneOffset;
}
function addDateMinutes(date, minutes) {
    date = new Date(date.getTime());
    date.setMinutes(date.getMinutes() + minutes);
    return date;
}
function convertTimezoneToLocal(date, timezone, reverse) {
    var reverseValue = reverse ? -1 : 1;
    var dateTimezoneOffset = date.getTimezoneOffset();
    var timezoneOffset = timezoneToOffset(timezone, dateTimezoneOffset);
    return addDateMinutes(date, reverseValue * (timezoneOffset - dateTimezoneOffset));
}
/**
 * Converts a value to date.
 *
 * Supported input formats:
 * - `Date`
 * - number: timestamp
 * - string: numeric (e.g. "1234"), ISO and date strings in a format supported by
 *   [Date.parse()](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date/parse).
 *   Note: ISO strings without time return a date without timeoffset.
 *
 * Throws if unable to convert to a date.
 */
function toDate(value) {
    if (isDate(value)) {
        return value;
    }
    if (typeof value === 'number' && !isNaN(value)) {
        return new Date(value);
    }
    if (typeof value === 'string') {
        value = value.trim();
        var parsedNb = parseFloat(value);
        // any string that only contains numbers, like "1234" but not like "1234hello"
        if (!isNaN(value - parsedNb)) {
            return new Date(parsedNb);
        }
        if (/^(\d{4}-\d{1,2}-\d{1,2})$/.test(value)) {
            /* For ISO Strings without time the day, month and year must be extracted from the ISO String
                  before Date creation to avoid time offset and errors in the new Date.
                  If we only replace '-' with ',' in the ISO String ("2015,01,01"), and try to create a new
                  date, some browsers (e.g. IE 9) will throw an invalid Date error.
                  If we leave the '-' ("2015-01-01") and try to create a new Date("2015-01-01") the timeoffset
                  is applied.
                  Note: ISO months are 0 for January, 1 for February, ... */
            var _a = value.split('-').map(function (val) { return +val; }), y = _a[0], m = _a[1], d = _a[2];
            return new Date(y, m - 1, d);
        }
        var match = void 0;
        if (match = value.match(exports.ISO8601_DATE_REGEX)) {
            return isoStringToDate(match);
        }
    }
    var date = new Date(value);
    if (!isDate(date)) {
        throw new Error("Unable to convert \"" + value + "\" into a date");
    }
    return date;
}
exports.toDate = toDate;
/**
 * Converts a date in ISO8601 to a Date.
 * Used instead of `Date.parse` because of browser discrepancies.
 */
function isoStringToDate(match) {
    var date = new Date(0);
    var tzHour = 0;
    var tzMin = 0;
    // match[8] means that the string contains "Z" (UTC) or a timezone like "+01:00" or "+0100"
    var dateSetter = match[8] ? date.setUTCFullYear : date.setFullYear;
    var timeSetter = match[8] ? date.setUTCHours : date.setHours;
    // if there is a timezone defined like "+01:00" or "+0100"
    if (match[9]) {
        tzHour = Number(match[9] + match[10]);
        tzMin = Number(match[9] + match[11]);
    }
    dateSetter.call(date, Number(match[1]), Number(match[2]) - 1, Number(match[3]));
    var h = Number(match[4] || 0) - tzHour;
    var m = Number(match[5] || 0) - tzMin;
    var s = Number(match[6] || 0);
    var ms = Math.round(parseFloat('0.' + (match[7] || 0)) * 1000);
    timeSetter.call(date, h, m, s, ms);
    return date;
}
exports.isoStringToDate = isoStringToDate;
function isDate(value) {
    return value instanceof Date && !isNaN(value.valueOf());
}
exports.isDate = isDate;
//# sourceMappingURL=format_date.js.map