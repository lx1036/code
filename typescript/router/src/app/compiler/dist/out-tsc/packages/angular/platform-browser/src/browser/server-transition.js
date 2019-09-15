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
var dom_tokens_1 = require("../dom/dom_tokens");
/**
 * An id that identifies a particular application being bootstrapped, that should
 * match across the client/server boundary.
 */
exports.TRANSITION_ID = new core_1.InjectionToken('TRANSITION_ID');
function appInitializerFactory(transitionId, document, injector) {
    return function () {
        // Wait for all application initializers to be completed before removing the styles set by
        // the server.
        injector.get(core_1.ApplicationInitStatus).donePromise.then(function () {
            var dom = dom_adapter_1.getDOM();
            var styles = Array.prototype.slice.apply(dom.querySelectorAll(document, "style[ng-transition]"));
            styles.filter(function (el) { return dom.getAttribute(el, 'ng-transition') === transitionId; })
                .forEach(function (el) { return dom.remove(el); });
        });
    };
}
exports.appInitializerFactory = appInitializerFactory;
exports.SERVER_TRANSITION_PROVIDERS = [
    {
        provide: core_1.APP_INITIALIZER,
        useFactory: appInitializerFactory,
        deps: [exports.TRANSITION_ID, dom_tokens_1.DOCUMENT, core_1.Injector],
        multi: true
    },
];
//# sourceMappingURL=server-transition.js.map