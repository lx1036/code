"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var event_emitter_1 = require("../event_emitter");
var util_1 = require("../util");
/**
 * An unmodifiable list of items that Angular keeps up to date when the state
 * of the application changes.
 *
 * The type of object that {@link ViewChildren}, {@link ContentChildren}, and {@link QueryList}
 * provide.
 *
 * Implements an iterable interface, therefore it can be used in both ES6
 * javascript `for (var i of items)` loops as well as in Angular templates with
 * `*ngFor="let i of myList"`.
 *
 * Changes can be observed by subscribing to the changes `Observable`.
 *
 * NOTE: In the future this class will implement an `Observable` interface.
 *
 * ### Example
 * ```typescript
 * @Component({...})
 * class Container {
 *   @ViewChildren(Item) items:QueryList<Item>;
 * }
 * ```
 *
 */
var /**
 * An unmodifiable list of items that Angular keeps up to date when the state
 * of the application changes.
 *
 * The type of object that {@link ViewChildren}, {@link ContentChildren}, and {@link QueryList}
 * provide.
 *
 * Implements an iterable interface, therefore it can be used in both ES6
 * javascript `for (var i of items)` loops as well as in Angular templates with
 * `*ngFor="let i of myList"`.
 *
 * Changes can be observed by subscribing to the changes `Observable`.
 *
 * NOTE: In the future this class will implement an `Observable` interface.
 *
 * ### Example
 * ```typescript
 * @Component({...})
 * class Container {
 *   @ViewChildren(Item) items:QueryList<Item>;
 * }
 * ```
 *
 */
QueryList = /** @class */ (function () {
    function QueryList() {
        this.dirty = true;
        this._results = [];
        this.changes = new event_emitter_1.EventEmitter();
        this.length = 0;
    }
    /**
     * See
     * [Array.map](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/map)
     */
    /**
       * See
       * [Array.map](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/map)
       */
    QueryList.prototype.map = /**
       * See
       * [Array.map](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/map)
       */
    function (fn) { return this._results.map(fn); };
    /**
     * See
     * [Array.filter](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/filter)
     */
    /**
       * See
       * [Array.filter](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/filter)
       */
    QueryList.prototype.filter = /**
       * See
       * [Array.filter](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/filter)
       */
    function (fn) {
        return this._results.filter(fn);
    };
    /**
     * See
     * [Array.find](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/find)
     */
    /**
       * See
       * [Array.find](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/find)
       */
    QueryList.prototype.find = /**
       * See
       * [Array.find](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/find)
       */
    function (fn) {
        return this._results.find(fn);
    };
    /**
     * See
     * [Array.reduce](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/reduce)
     */
    /**
       * See
       * [Array.reduce](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/reduce)
       */
    QueryList.prototype.reduce = /**
       * See
       * [Array.reduce](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/reduce)
       */
    function (fn, init) {
        return this._results.reduce(fn, init);
    };
    /**
     * See
     * [Array.forEach](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/forEach)
     */
    /**
       * See
       * [Array.forEach](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/forEach)
       */
    QueryList.prototype.forEach = /**
       * See
       * [Array.forEach](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/forEach)
       */
    function (fn) { this._results.forEach(fn); };
    /**
     * See
     * [Array.some](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/some)
     */
    /**
       * See
       * [Array.some](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/some)
       */
    QueryList.prototype.some = /**
       * See
       * [Array.some](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/some)
       */
    function (fn) {
        return this._results.some(fn);
    };
    QueryList.prototype.toArray = function () { return this._results.slice(); };
    QueryList.prototype[util_1.getSymbolIterator()] = function () { return this._results[util_1.getSymbolIterator()](); };
    QueryList.prototype.toString = function () { return this._results.toString(); };
    QueryList.prototype.reset = function (res) {
        this._results = flatten(res);
        this.dirty = false;
        this.length = this._results.length;
        this.last = this._results[this.length - 1];
        this.first = this._results[0];
    };
    QueryList.prototype.notifyOnChanges = function () { this.changes.emit(this); };
    /** internal */
    /** internal */
    QueryList.prototype.setDirty = /** internal */
    function () { this.dirty = true; };
    /** internal */
    /** internal */
    QueryList.prototype.destroy = /** internal */
    function () {
        this.changes.complete();
        this.changes.unsubscribe();
    };
    return QueryList;
}());
exports.QueryList = QueryList;
function flatten(list) {
    return list.reduce(function (flat, item) {
        var flatItem = Array.isArray(item) ? flatten(item) : item;
        return flat.concat(flatItem);
    }, []);
}
//# sourceMappingURL=query_list.js.map