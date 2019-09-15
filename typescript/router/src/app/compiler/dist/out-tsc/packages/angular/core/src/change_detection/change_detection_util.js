"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var util_1 = require("../util");
function devModeEqual(a, b) {
    var isListLikeIterableA = isListLikeIterable(a);
    var isListLikeIterableB = isListLikeIterable(b);
    if (isListLikeIterableA && isListLikeIterableB) {
        return areIterablesEqual(a, b, devModeEqual);
    }
    else {
        var isAObject = a && (typeof a === 'object' || typeof a === 'function');
        var isBObject = b && (typeof b === 'object' || typeof b === 'function');
        if (!isListLikeIterableA && isAObject && !isListLikeIterableB && isBObject) {
            return true;
        }
        else {
            return util_1.looseIdentical(a, b);
        }
    }
}
exports.devModeEqual = devModeEqual;
/**
 * Indicates that the result of a {@link Pipe} transformation has changed even though the
 * reference has not changed.
 *
 * Wrapped values are unwrapped automatically during the change detection, and the unwrapped value
 * is stored.
 *
 * Example:
 *
 * ```
 * if (this._latestValue === this._latestReturnedValue) {
 *    return this._latestReturnedValue;
 *  } else {
 *    this._latestReturnedValue = this._latestValue;
 *    return WrappedValue.wrap(this._latestValue); // this will force update
 *  }
 * ```
 *
 */
var /**
 * Indicates that the result of a {@link Pipe} transformation has changed even though the
 * reference has not changed.
 *
 * Wrapped values are unwrapped automatically during the change detection, and the unwrapped value
 * is stored.
 *
 * Example:
 *
 * ```
 * if (this._latestValue === this._latestReturnedValue) {
 *    return this._latestReturnedValue;
 *  } else {
 *    this._latestReturnedValue = this._latestValue;
 *    return WrappedValue.wrap(this._latestValue); // this will force update
 *  }
 * ```
 *
 */
WrappedValue = /** @class */ (function () {
    function WrappedValue(value) {
        this.wrapped = value;
    }
    /** Creates a wrapped value. */
    /** Creates a wrapped value. */
    WrappedValue.wrap = /** Creates a wrapped value. */
    function (value) { return new WrappedValue(value); };
    /**
     * Returns the underlying value of a wrapped value.
     * Returns the given `value` when it is not wrapped.
     **/
    /**
       * Returns the underlying value of a wrapped value.
       * Returns the given `value` when it is not wrapped.
       **/
    WrappedValue.unwrap = /**
       * Returns the underlying value of a wrapped value.
       * Returns the given `value` when it is not wrapped.
       **/
    function (value) { return WrappedValue.isWrapped(value) ? value.wrapped : value; };
    /** Returns true if `value` is a wrapped value. */
    /** Returns true if `value` is a wrapped value. */
    WrappedValue.isWrapped = /** Returns true if `value` is a wrapped value. */
    function (value) { return value instanceof WrappedValue; };
    return WrappedValue;
}());
exports.WrappedValue = WrappedValue;
/**
 * Represents a basic change from a previous to a new value.
 *
 */
var /**
 * Represents a basic change from a previous to a new value.
 *
 */
SimpleChange = /** @class */ (function () {
    function SimpleChange(previousValue, currentValue, firstChange) {
        this.previousValue = previousValue;
        this.currentValue = currentValue;
        this.firstChange = firstChange;
    }
    /**
     * Check whether the new value is the first value assigned.
     */
    /**
       * Check whether the new value is the first value assigned.
       */
    SimpleChange.prototype.isFirstChange = /**
       * Check whether the new value is the first value assigned.
       */
    function () { return this.firstChange; };
    return SimpleChange;
}());
exports.SimpleChange = SimpleChange;
function isListLikeIterable(obj) {
    if (!isJsObject(obj))
        return false;
    return Array.isArray(obj) ||
        (!(obj instanceof Map) && // JS Map are iterables but return entries as [k, v]
            // JS Map are iterables but return entries as [k, v]
            util_1.getSymbolIterator() in obj); // JS Iterable have a Symbol.iterator prop
}
exports.isListLikeIterable = isListLikeIterable;
function areIterablesEqual(a, b, comparator) {
    var iterator1 = a[util_1.getSymbolIterator()]();
    var iterator2 = b[util_1.getSymbolIterator()]();
    while (true) {
        var item1 = iterator1.next();
        var item2 = iterator2.next();
        if (item1.done && item2.done)
            return true;
        if (item1.done || item2.done)
            return false;
        if (!comparator(item1.value, item2.value))
            return false;
    }
}
exports.areIterablesEqual = areIterablesEqual;
function iterateListLike(obj, fn) {
    if (Array.isArray(obj)) {
        for (var i = 0; i < obj.length; i++) {
            fn(obj[i]);
        }
    }
    else {
        var iterator = obj[util_1.getSymbolIterator()]();
        var item = void 0;
        while (!((item = iterator.next()).done)) {
            fn(item.value);
        }
    }
}
exports.iterateListLike = iterateListLike;
function isJsObject(o) {
    return o !== null && (typeof o === 'function' || typeof o === 'object');
}
exports.isJsObject = isJsObject;
//# sourceMappingURL=change_detection_util.js.map