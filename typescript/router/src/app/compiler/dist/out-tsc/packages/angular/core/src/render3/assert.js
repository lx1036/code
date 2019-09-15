"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
// The functions in this file verify that the assumptions we are making
// about state in an instruction are correct before implementing any logic.
// They are meant only to be called in dev mode as sanity checks.
function assertNumber(actual, msg) {
    if (typeof actual != 'number') {
        throwError(msg);
    }
}
exports.assertNumber = assertNumber;
function assertEqual(actual, expected, msg) {
    if (actual != expected) {
        throwError(msg);
    }
}
exports.assertEqual = assertEqual;
function assertNotEqual(actual, expected, msg) {
    if (actual == expected) {
        throwError(msg);
    }
}
exports.assertNotEqual = assertNotEqual;
function assertSame(actual, expected, msg) {
    if (actual !== expected) {
        throwError(msg);
    }
}
exports.assertSame = assertSame;
function assertLessThan(actual, expected, msg) {
    if (actual >= expected) {
        throwError(msg);
    }
}
exports.assertLessThan = assertLessThan;
function assertGreaterThan(actual, expected, msg) {
    if (actual <= expected) {
        throwError(msg);
    }
}
exports.assertGreaterThan = assertGreaterThan;
function assertNull(actual, msg) {
    if (actual != null) {
        throwError(msg);
    }
}
exports.assertNull = assertNull;
function assertNotNull(actual, msg) {
    if (actual == null) {
        throwError(msg);
    }
}
exports.assertNotNull = assertNotNull;
function assertComponentType(actual, msg) {
    if (msg === void 0) { msg = 'Type passed in is not ComponentType, it does not have \'ngComponentDef\' property.'; }
    if (!actual.ngComponentDef) {
        throwError(msg);
    }
}
exports.assertComponentType = assertComponentType;
function throwError(msg) {
    debugger; // Left intentionally for better debugger experience.
    throw new Error("ASSERTION ERROR: " + msg);
}
//# sourceMappingURL=assert.js.map