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
 * Determine if the argument is shaped like a Promise
 */
function isPromise(obj) {
    // allow any Promise/A+ compliant thenable.
    // It's up to the caller to ensure that obj.then conforms to the spec
    return !!obj && typeof obj.then === 'function';
}
exports.isPromise = isPromise;
/**
 * Determine if the argument is an Observable
 */
function isObservable(obj) {
    // TODO: use Symbol.observable when https://github.com/ReactiveX/rxjs/issues/2415 will be resolved
    return !!obj && typeof obj.subscribe === 'function';
}
exports.isObservable = isObservable;
//# sourceMappingURL=lang.js.map