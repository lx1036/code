"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var util_1 = require("../util");
var forward_ref_1 = require("./forward_ref");
/**
 * A unique object used for retrieving items from the {@link ReflectiveInjector}.
 *
 * Keys have:
 * - a system-wide unique `id`.
 * - a `token`.
 *
 * `Key` is used internally by {@link ReflectiveInjector} because its system-wide unique `id` allows
 * the
 * injector to store created objects in a more efficient way.
 *
 * `Key` should not be created directly. {@link ReflectiveInjector} creates keys automatically when
 * resolving
 * providers.
 * @deprecated No replacement
 */
var /**
 * A unique object used for retrieving items from the {@link ReflectiveInjector}.
 *
 * Keys have:
 * - a system-wide unique `id`.
 * - a `token`.
 *
 * `Key` is used internally by {@link ReflectiveInjector} because its system-wide unique `id` allows
 * the
 * injector to store created objects in a more efficient way.
 *
 * `Key` should not be created directly. {@link ReflectiveInjector} creates keys automatically when
 * resolving
 * providers.
 * @deprecated No replacement
 */
ReflectiveKey = /** @class */ (function () {
    /**
     * Private
     */
    function ReflectiveKey(token, id) {
        this.token = token;
        this.id = id;
        if (!token) {
            throw new Error('Token must be defined!');
        }
        this.displayName = util_1.stringify(this.token);
    }
    /**
     * Retrieves a `Key` for a token.
     */
    /**
       * Retrieves a `Key` for a token.
       */
    ReflectiveKey.get = /**
       * Retrieves a `Key` for a token.
       */
    function (token) {
        return _globalKeyRegistry.get(forward_ref_1.resolveForwardRef(token));
    };
    Object.defineProperty(ReflectiveKey, "numberOfKeys", {
        /**
         * @returns the number of keys registered in the system.
         */
        get: /**
           * @returns the number of keys registered in the system.
           */
        function () { return _globalKeyRegistry.numberOfKeys; },
        enumerable: true,
        configurable: true
    });
    return ReflectiveKey;
}());
exports.ReflectiveKey = ReflectiveKey;
var KeyRegistry = /** @class */ (function () {
    function KeyRegistry() {
        this._allKeys = new Map();
    }
    KeyRegistry.prototype.get = function (token) {
        if (token instanceof ReflectiveKey)
            return token;
        if (this._allKeys.has(token)) {
            return this._allKeys.get(token);
        }
        var newKey = new ReflectiveKey(token, ReflectiveKey.numberOfKeys);
        this._allKeys.set(token, newKey);
        return newKey;
    };
    Object.defineProperty(KeyRegistry.prototype, "numberOfKeys", {
        get: function () { return this._allKeys.size; },
        enumerable: true,
        configurable: true
    });
    return KeyRegistry;
}());
exports.KeyRegistry = KeyRegistry;
var _globalKeyRegistry = new KeyRegistry();
//# sourceMappingURL=reflective_key.js.map