"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var errors_1 = require("../errors");
function expressionChangedAfterItHasBeenCheckedError(context, oldValue, currValue, isFirstCheck) {
    var msg = "ExpressionChangedAfterItHasBeenCheckedError: Expression has changed after it was checked. Previous value: '" + oldValue + "'. Current value: '" + currValue + "'.";
    if (isFirstCheck) {
        msg +=
            " It seems like the view has been created after its parent and its children have been dirty checked." +
                " Has it been created in a change detection hook ?";
    }
    return viewDebugError(msg, context);
}
exports.expressionChangedAfterItHasBeenCheckedError = expressionChangedAfterItHasBeenCheckedError;
function viewWrappedDebugError(err, context) {
    if (!(err instanceof Error)) {
        // errors that are not Error instances don't have a stack,
        // so it is ok to wrap them into a new Error object...
        err = new Error(err.toString());
    }
    _addDebugContext(err, context);
    return err;
}
exports.viewWrappedDebugError = viewWrappedDebugError;
function viewDebugError(msg, context) {
    var err = new Error(msg);
    _addDebugContext(err, context);
    return err;
}
exports.viewDebugError = viewDebugError;
function _addDebugContext(err, context) {
    err[errors_1.ERROR_DEBUG_CONTEXT] = context;
    err[errors_1.ERROR_LOGGER] = context.logError.bind(context);
}
function isViewDebugError(err) {
    return !!errors_1.getDebugContext(err);
}
exports.isViewDebugError = isViewDebugError;
function viewDestroyedError(action) {
    return new Error("ViewDestroyedError: Attempt to use a destroyed view: " + action);
}
exports.viewDestroyedError = viewDestroyedError;
//# sourceMappingURL=errors.js.map