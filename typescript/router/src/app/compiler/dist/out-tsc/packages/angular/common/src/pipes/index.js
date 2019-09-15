"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var async_pipe_1 = require("./async_pipe");
exports.AsyncPipe = async_pipe_1.AsyncPipe;
var case_conversion_pipes_1 = require("./case_conversion_pipes");
exports.LowerCasePipe = case_conversion_pipes_1.LowerCasePipe;
exports.TitleCasePipe = case_conversion_pipes_1.TitleCasePipe;
exports.UpperCasePipe = case_conversion_pipes_1.UpperCasePipe;
var date_pipe_1 = require("./date_pipe");
exports.DatePipe = date_pipe_1.DatePipe;
var i18n_plural_pipe_1 = require("./i18n_plural_pipe");
exports.I18nPluralPipe = i18n_plural_pipe_1.I18nPluralPipe;
var i18n_select_pipe_1 = require("./i18n_select_pipe");
exports.I18nSelectPipe = i18n_select_pipe_1.I18nSelectPipe;
var json_pipe_1 = require("./json_pipe");
exports.JsonPipe = json_pipe_1.JsonPipe;
var number_pipe_1 = require("./number_pipe");
exports.CurrencyPipe = number_pipe_1.CurrencyPipe;
exports.DecimalPipe = number_pipe_1.DecimalPipe;
exports.PercentPipe = number_pipe_1.PercentPipe;
var slice_pipe_1 = require("./slice_pipe");
exports.SlicePipe = slice_pipe_1.SlicePipe;
/**
 * A collection of Angular pipes that are likely to be used in each and every application.
 */
exports.COMMON_PIPES = [
    async_pipe_1.AsyncPipe,
    case_conversion_pipes_1.UpperCasePipe,
    case_conversion_pipes_1.LowerCasePipe,
    json_pipe_1.JsonPipe,
    slice_pipe_1.SlicePipe,
    number_pipe_1.DecimalPipe,
    number_pipe_1.PercentPipe,
    case_conversion_pipes_1.TitleCasePipe,
    number_pipe_1.CurrencyPipe,
    date_pipe_1.DatePipe,
    i18n_plural_pipe_1.I18nPluralPipe,
    i18n_select_pipe_1.I18nSelectPipe,
];
//# sourceMappingURL=index.js.map