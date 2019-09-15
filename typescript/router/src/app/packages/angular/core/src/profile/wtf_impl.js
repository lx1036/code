"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
exports.__esModule = true;
var util_1 = require("../util");
var trace;
var events;
function detectWTF() {
    var wtf = util_1.global /** TODO #9100 */['wtf'];
    if (wtf) {
        trace = wtf['trace'];
        if (trace) {
            events = trace['events'];
            return true;
        }
    }
    return false;
}
exports.detectWTF = detectWTF;
function createScope(signature, flags) {
    if (flags === void 0) { flags = null; }
    return events.createScope(signature, flags);
}
exports.createScope = createScope;
function leave(scope, returnValue) {
    trace.leaveScope(scope, returnValue);
    return returnValue;
}
exports.leave = leave;
function startTimeRange(rangeType, action) {
    return trace.beginTimeRange(rangeType, action);
}
exports.startTimeRange = startTimeRange;
function endTimeRange(range) {
    trace.endTimeRange(range);
}
exports.endTimeRange = endTimeRange;
