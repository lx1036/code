"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var date_pipe_1 = require("./date_pipe");
exports.DeprecatedDatePipe = date_pipe_1.DeprecatedDatePipe;
var number_pipe_1 = require("./number_pipe");
exports.DeprecatedCurrencyPipe = number_pipe_1.DeprecatedCurrencyPipe;
exports.DeprecatedDecimalPipe = number_pipe_1.DeprecatedDecimalPipe;
exports.DeprecatedPercentPipe = number_pipe_1.DeprecatedPercentPipe;
/**
 * A collection of deprecated i18n pipes that require intl api
 *
 * @deprecated from v5
 */
exports.COMMON_DEPRECATED_I18N_PIPES = [number_pipe_1.DeprecatedDecimalPipe, number_pipe_1.DeprecatedPercentPipe, number_pipe_1.DeprecatedCurrencyPipe, date_pipe_1.DeprecatedDatePipe];
//# sourceMappingURL=index.js.map