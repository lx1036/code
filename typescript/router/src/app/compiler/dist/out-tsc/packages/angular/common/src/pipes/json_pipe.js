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
/**
 * @ngModule CommonModule
 * @description
 *
 * Converts value into string using `JSON.stringify`. Useful for debugging.
 *
 * ### Example
 * {@example common/pipes/ts/json_pipe.ts region='JsonPipe'}
 *
 *
 */
var JsonPipe = /** @class */ (function () {
    function JsonPipe() {
    }
    JsonPipe.prototype.transform = function (value) { return JSON.stringify(value, null, 2); };
    JsonPipe.decorators = [
        { type: core_1.Pipe, args: [{ name: 'json', pure: false },] },
    ];
    return JsonPipe;
}());
exports.JsonPipe = JsonPipe;
//# sourceMappingURL=json_pipe.js.map