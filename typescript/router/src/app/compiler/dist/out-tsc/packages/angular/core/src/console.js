"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var di_1 = require("./di");
var Console = /** @class */ (function () {
    function Console() {
    }
    Console.prototype.log = function (message) {
        // tslint:disable-next-line:no-console
        console.log(message);
    };
    // Note: for reporting errors use `DOM.logError()` as it is platform specific
    // Note: for reporting errors use `DOM.logError()` as it is platform specific
    Console.prototype.warn = 
    // Note: for reporting errors use `DOM.logError()` as it is platform specific
    function (message) {
        // tslint:disable-next-line:no-console
        console.warn(message);
    };
    Console.decorators = [
        { type: di_1.Injectable },
    ];
    return Console;
}());
exports.Console = Console;
//# sourceMappingURL=console.js.map