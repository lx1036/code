"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
function parseCookieValue(cookieStr, name) {
    name = encodeURIComponent(name);
    for (var _i = 0, _a = cookieStr.split(';'); _i < _a.length; _i++) {
        var cookie = _a[_i];
        var eqIndex = cookie.indexOf('=');
        var _b = eqIndex == -1 ? [cookie, ''] : [cookie.slice(0, eqIndex), cookie.slice(eqIndex + 1)], cookieName = _b[0], cookieValue = _b[1];
        if (cookieName.trim() === name) {
            return decodeURIComponent(cookieValue);
        }
    }
    return null;
}
exports.parseCookieValue = parseCookieValue;
//# sourceMappingURL=cookie.js.map