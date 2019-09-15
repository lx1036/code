"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var dom_adapter_1 = require("../../dom/dom_adapter");
/**
 * Predicates for use with {@link DebugElement}'s query functions.
 *
 * @experimental All debugging apis are currently experimental.
 */
var /**
 * Predicates for use with {@link DebugElement}'s query functions.
 *
 * @experimental All debugging apis are currently experimental.
 */
By = /** @class */ (function () {
    function By() {
    }
    /**
     * Match all elements.
     *
     * ## Example
     *
     * {@example platform-browser/dom/debug/ts/by/by.ts region='by_all'}
     */
    /**
       * Match all elements.
       *
       * ## Example
       *
       * {@example platform-browser/dom/debug/ts/by/by.ts region='by_all'}
       */
    By.all = /**
       * Match all elements.
       *
       * ## Example
       *
       * {@example platform-browser/dom/debug/ts/by/by.ts region='by_all'}
       */
    function () { return function (debugElement) { return true; }; };
    /**
     * Match elements by the given CSS selector.
     *
     * ## Example
     *
     * {@example platform-browser/dom/debug/ts/by/by.ts region='by_css'}
     */
    /**
       * Match elements by the given CSS selector.
       *
       * ## Example
       *
       * {@example platform-browser/dom/debug/ts/by/by.ts region='by_css'}
       */
    By.css = /**
       * Match elements by the given CSS selector.
       *
       * ## Example
       *
       * {@example platform-browser/dom/debug/ts/by/by.ts region='by_css'}
       */
    function (selector) {
        return function (debugElement) {
            return debugElement.nativeElement != null ?
                dom_adapter_1.getDOM().elementMatches(debugElement.nativeElement, selector) :
                false;
        };
    };
    /**
     * Match elements that have the given directive present.
     *
     * ## Example
     *
     * {@example platform-browser/dom/debug/ts/by/by.ts region='by_directive'}
     */
    /**
       * Match elements that have the given directive present.
       *
       * ## Example
       *
       * {@example platform-browser/dom/debug/ts/by/by.ts region='by_directive'}
       */
    By.directive = /**
       * Match elements that have the given directive present.
       *
       * ## Example
       *
       * {@example platform-browser/dom/debug/ts/by/by.ts region='by_directive'}
       */
    function (type) {
        return function (debugElement) { return debugElement.providerTokens.indexOf(type) !== -1; };
    };
    return By;
}());
exports.By = By;
//# sourceMappingURL=by.js.map