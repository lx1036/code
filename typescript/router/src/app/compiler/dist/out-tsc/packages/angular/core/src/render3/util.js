"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
/**
* Must use this method for CD (instead of === ) since NaN !== NaN
*/
function isDifferent(a, b) {
    // NaN is the only value that is not equal to itself so the first
    // test checks if both a and b are not NaN
    return !(a !== a && b !== b) && a !== b;
}
exports.isDifferent = isDifferent;
function stringify(value) {
    if (typeof value == 'function')
        return value.name || value;
    if (typeof value == 'string')
        return value;
    if (value == null)
        return '';
    return '' + value;
}
exports.stringify = stringify;
/**
 *  Function that throws a "not implemented" error so it's clear certain
 *  behaviors/methods aren't yet ready.
 *
 * @returns Not implemented error
 */
function notImplemented() {
    return new Error('NotImplemented');
}
exports.notImplemented = notImplemented;
/**
 * Flattens an array in non-recursive way. Input arrays are not modified.
 */
function flatten(list) {
    var result = [];
    var i = 0;
    while (i < list.length) {
        var item = list[i];
        if (Array.isArray(item)) {
            if (item.length > 0) {
                list = item.concat(list.slice(i + 1));
                i = 0;
            }
            else {
                i++;
            }
        }
        else {
            result.push(item);
            i++;
        }
    }
    return result;
}
exports.flatten = flatten;
//# sourceMappingURL=util.js.map