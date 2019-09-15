"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var _DOM = (null);
function getDOM() {
    return _DOM;
}
exports.getDOM = getDOM;
function setDOM(adapter) {
    _DOM = adapter;
}
exports.setDOM = setDOM;
function setRootDomAdapter(adapter) {
    if (!_DOM) {
        _DOM = adapter;
    }
}
exports.setRootDomAdapter = setRootDomAdapter;
/* tslint:disable:requireParameterType */
/**
 * Provides DOM operations in an environment-agnostic way.
 *
 * @security Tread carefully! Interacting with the DOM directly is dangerous and
 * can introduce XSS risks.
 */
var /* tslint:disable:requireParameterType */
/**
 * Provides DOM operations in an environment-agnostic way.
 *
 * @security Tread carefully! Interacting with the DOM directly is dangerous and
 * can introduce XSS risks.
 */
DomAdapter = /** @class */ (function () {
    function DomAdapter() {
        this.resourceLoaderType = null;
    }
    Object.defineProperty(DomAdapter.prototype, "attrToPropMap", {
        /**
         * Maps attribute names to their corresponding property names for cases
         * where attribute name doesn't match property name.
         */
        get: /**
           * Maps attribute names to their corresponding property names for cases
           * where attribute name doesn't match property name.
           */
        function () { return this._attrToPropMap; },
        set: function (value) { this._attrToPropMap = value; },
        enumerable: true,
        configurable: true
    });
    return DomAdapter;
}());
exports.DomAdapter = DomAdapter;
//# sourceMappingURL=dom_adapter.js.map