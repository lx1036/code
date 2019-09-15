"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var dom_adapter_1 = require("../dom/dom_adapter");
var BrowserGetTestability = /** @class */ (function () {
    function BrowserGetTestability() {
    }
    BrowserGetTestability.init = function () { core_1.setTestabilityGetter(new BrowserGetTestability()); };
    BrowserGetTestability.prototype.addToWindow = function (registry) {
        core_1.ɵglobal['getAngularTestability'] = function (elem, findInAncestors) {
            if (findInAncestors === void 0) { findInAncestors = true; }
            var testability = registry.findTestabilityInTree(elem, findInAncestors);
            if (testability == null) {
                throw new Error('Could not find testability for element.');
            }
            return testability;
        };
        core_1.ɵglobal['getAllAngularTestabilities'] = function () { return registry.getAllTestabilities(); };
        core_1.ɵglobal['getAllAngularRootElements'] = function () { return registry.getAllRootElements(); };
        var whenAllStable = function (callback /** TODO #9100 */) {
            var testabilities = core_1.ɵglobal['getAllAngularTestabilities']();
            var count = testabilities.length;
            var didWork = false;
            var decrement = function (didWork_ /** TODO #9100 */) {
                didWork = didWork || didWork_;
                count--;
                if (count == 0) {
                    callback(didWork);
                }
            };
            testabilities.forEach(function (testability /** TODO #9100 */) {
                testability.whenStable(decrement);
            });
        };
        if (!core_1.ɵglobal['frameworkStabilizers']) {
            core_1.ɵglobal['frameworkStabilizers'] = [];
        }
        core_1.ɵglobal['frameworkStabilizers'].push(whenAllStable);
    };
    BrowserGetTestability.prototype.findTestabilityInTree = function (registry, elem, findInAncestors) {
        if (elem == null) {
            return null;
        }
        var t = registry.getTestability(elem);
        if (t != null) {
            return t;
        }
        else if (!findInAncestors) {
            return null;
        }
        if (dom_adapter_1.getDOM().isShadowRoot(elem)) {
            return this.findTestabilityInTree(registry, dom_adapter_1.getDOM().getHost(elem), true);
        }
        return this.findTestabilityInTree(registry, dom_adapter_1.getDOM().parentElement(elem), true);
    };
    return BrowserGetTestability;
}());
exports.BrowserGetTestability = BrowserGetTestability;
//# sourceMappingURL=testability.js.map