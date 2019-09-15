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
 * Entry point for all public APIs of the common package.
 */
__export(require("./location/index"));
var format_date_1 = require("./i18n/format_date");
exports.formatDate = format_date_1.formatDate;
var format_number_1 = require("./i18n/format_number");
exports.formatCurrency = format_number_1.formatCurrency;
exports.formatNumber = format_number_1.formatNumber;
exports.formatPercent = format_number_1.formatPercent;
var localization_1 = require("./i18n/localization");
exports.NgLocaleLocalization = localization_1.NgLocaleLocalization;
exports.NgLocalization = localization_1.NgLocalization;
var locale_data_1 = require("./i18n/locale_data");
exports.registerLocaleData = locale_data_1.registerLocaleData;
var locale_data_api_1 = require("./i18n/locale_data_api");
exports.Plural = locale_data_api_1.Plural;
exports.NumberFormatStyle = locale_data_api_1.NumberFormatStyle;
exports.FormStyle = locale_data_api_1.FormStyle;
exports.TranslationWidth = locale_data_api_1.TranslationWidth;
exports.FormatWidth = locale_data_api_1.FormatWidth;
exports.NumberSymbol = locale_data_api_1.NumberSymbol;
exports.WeekDay = locale_data_api_1.WeekDay;
exports.getNumberOfCurrencyDigits = locale_data_api_1.getNumberOfCurrencyDigits;
exports.getCurrencySymbol = locale_data_api_1.getCurrencySymbol;
exports.getLocaleDayPeriods = locale_data_api_1.getLocaleDayPeriods;
exports.getLocaleDayNames = locale_data_api_1.getLocaleDayNames;
exports.getLocaleMonthNames = locale_data_api_1.getLocaleMonthNames;
exports.getLocaleId = locale_data_api_1.getLocaleId;
exports.getLocaleEraNames = locale_data_api_1.getLocaleEraNames;
exports.getLocaleWeekEndRange = locale_data_api_1.getLocaleWeekEndRange;
exports.getLocaleFirstDayOfWeek = locale_data_api_1.getLocaleFirstDayOfWeek;
exports.getLocaleDateFormat = locale_data_api_1.getLocaleDateFormat;
exports.getLocaleDateTimeFormat = locale_data_api_1.getLocaleDateTimeFormat;
exports.getLocaleExtraDayPeriodRules = locale_data_api_1.getLocaleExtraDayPeriodRules;
exports.getLocaleExtraDayPeriods = locale_data_api_1.getLocaleExtraDayPeriods;
exports.getLocalePluralCase = locale_data_api_1.getLocalePluralCase;
exports.getLocaleTimeFormat = locale_data_api_1.getLocaleTimeFormat;
exports.getLocaleNumberSymbol = locale_data_api_1.getLocaleNumberSymbol;
exports.getLocaleNumberFormat = locale_data_api_1.getLocaleNumberFormat;
exports.getLocaleCurrencyName = locale_data_api_1.getLocaleCurrencyName;
exports.getLocaleCurrencySymbol = locale_data_api_1.getLocaleCurrencySymbol;
var cookie_1 = require("./cookie");
exports.ɵparseCookieValue = cookie_1.parseCookieValue;
var common_module_1 = require("./common_module");
exports.CommonModule = common_module_1.CommonModule;
exports.DeprecatedI18NPipesModule = common_module_1.DeprecatedI18NPipesModule;
var index_1 = require("./directives/index");
exports.NgClass = index_1.NgClass;
exports.NgForOf = index_1.NgForOf;
exports.NgForOfContext = index_1.NgForOfContext;
exports.NgIf = index_1.NgIf;
exports.NgIfContext = index_1.NgIfContext;
exports.NgPlural = index_1.NgPlural;
exports.NgPluralCase = index_1.NgPluralCase;
exports.NgStyle = index_1.NgStyle;
exports.NgSwitch = index_1.NgSwitch;
exports.NgSwitchCase = index_1.NgSwitchCase;
exports.NgSwitchDefault = index_1.NgSwitchDefault;
exports.NgTemplateOutlet = index_1.NgTemplateOutlet;
exports.NgComponentOutlet = index_1.NgComponentOutlet;
var dom_tokens_1 = require("./dom_tokens");
exports.DOCUMENT = dom_tokens_1.DOCUMENT;
var index_2 = require("./pipes/index");
exports.AsyncPipe = index_2.AsyncPipe;
exports.DatePipe = index_2.DatePipe;
exports.I18nPluralPipe = index_2.I18nPluralPipe;
exports.I18nSelectPipe = index_2.I18nSelectPipe;
exports.JsonPipe = index_2.JsonPipe;
exports.LowerCasePipe = index_2.LowerCasePipe;
exports.CurrencyPipe = index_2.CurrencyPipe;
exports.DecimalPipe = index_2.DecimalPipe;
exports.PercentPipe = index_2.PercentPipe;
exports.SlicePipe = index_2.SlicePipe;
exports.UpperCasePipe = index_2.UpperCasePipe;
exports.TitleCasePipe = index_2.TitleCasePipe;
var index_3 = require("./pipes/deprecated/index");
exports.DeprecatedDatePipe = index_3.DeprecatedDatePipe;
exports.DeprecatedCurrencyPipe = index_3.DeprecatedCurrencyPipe;
exports.DeprecatedDecimalPipe = index_3.DeprecatedDecimalPipe;
exports.DeprecatedPercentPipe = index_3.DeprecatedPercentPipe;
var platform_id_1 = require("./platform_id");
exports.ɵPLATFORM_BROWSER_ID = platform_id_1.PLATFORM_BROWSER_ID;
exports.ɵPLATFORM_SERVER_ID = platform_id_1.PLATFORM_SERVER_ID;
exports.ɵPLATFORM_WORKER_APP_ID = platform_id_1.PLATFORM_WORKER_APP_ID;
exports.ɵPLATFORM_WORKER_UI_ID = platform_id_1.PLATFORM_WORKER_UI_ID;
exports.isPlatformBrowser = platform_id_1.isPlatformBrowser;
exports.isPlatformServer = platform_id_1.isPlatformServer;
exports.isPlatformWorkerApp = platform_id_1.isPlatformWorkerApp;
exports.isPlatformWorkerUi = platform_id_1.isPlatformWorkerUi;
var version_1 = require("./version");
exports.VERSION = version_1.VERSION;
//# sourceMappingURL=common.js.map