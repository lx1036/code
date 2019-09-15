"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var util_1 = require("../../dom/util");
var common_tools_1 = require("./common_tools");
var PROFILER_GLOBAL_NAME = 'profiler';
/**
 * Enabled Angular debug tools that are accessible via your browser's
 * developer console.
 *
 * Usage:
 *
 * 1. Open developer console (e.g. in Chrome Ctrl + Shift + j)
 * 1. Type `ng.` (usually the console will show auto-complete suggestion)
 * 1. Try the change detection profiler `ng.profiler.timeChangeDetection()`
 *    then hit Enter.
 *
 * @experimental All debugging apis are currently experimental.
 */
function enableDebugTools(ref) {
    util_1.exportNgVar(PROFILER_GLOBAL_NAME, new common_tools_1.AngularProfiler(ref));
    return ref;
}
exports.enableDebugTools = enableDebugTools;
/**
 * Disables Angular tools.
 *
 * @experimental All debugging apis are currently experimental.
 */
function disableDebugTools() {
    util_1.exportNgVar(PROFILER_GLOBAL_NAME, null);
}
exports.disableDebugTools = disableDebugTools;
//# sourceMappingURL=tools.js.map