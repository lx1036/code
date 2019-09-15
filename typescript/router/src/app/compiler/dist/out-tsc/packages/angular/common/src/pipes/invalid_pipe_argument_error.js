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
function invalidPipeArgumentError(type, value) {
    return Error("InvalidPipeArgument: '" + value + "' for pipe '" + core_1.ɵstringify(type) + "'");
}
exports.invalidPipeArgumentError = invalidPipeArgumentError;
//# sourceMappingURL=invalid_pipe_argument_error.js.map