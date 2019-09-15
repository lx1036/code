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
var dom_tokens_1 = require("../dom/dom_tokens");
function escapeHtml(text) {
    var escapedText = {
        '&': '&a;',
        '"': '&q;',
        '\'': '&s;',
        '<': '&l;',
        '>': '&g;',
    };
    return text.replace(/[&"'<>]/g, function (s) { return escapedText[s]; });
}
exports.escapeHtml = escapeHtml;
function unescapeHtml(text) {
    var unescapedText = {
        '&a;': '&',
        '&q;': '"',
        '&s;': '\'',
        '&l;': '<',
        '&g;': '>',
    };
    return text.replace(/&[^;]+;/g, function (s) { return unescapedText[s]; });
}
exports.unescapeHtml = unescapeHtml;
/**
 * Create a `StateKey<T>` that can be used to store value of type T with `TransferState`.
 *
 * Example:
 *
 * ```
 * const COUNTER_KEY = makeStateKey<number>('counter');
 * let value = 10;
 *
 * transferState.set(COUNTER_KEY, value);
 * ```
 *
 * @experimental
 */
function makeStateKey(key) {
    return key;
}
exports.makeStateKey = makeStateKey;
/**
 * A key value store that is transferred from the application on the server side to the application
 * on the client side.
 *
 * `TransferState` will be available as an injectable token. To use it import
 * `ServerTransferStateModule` on the server and `BrowserTransferStateModule` on the client.
 *
 * The values in the store are serialized/deserialized using JSON.stringify/JSON.parse. So only
 * boolean, number, string, null and non-class objects will be serialized and deserialzied in a
 * non-lossy manner.
 *
 * @experimental
 */
var TransferState = /** @class */ (function () {
    function TransferState() {
        this.store = {};
        this.onSerializeCallbacks = {};
    }
    /** @internal */
    /** @internal */
    TransferState.init = /** @internal */
    function (initState) {
        var transferState = new TransferState();
        transferState.store = initState;
        return transferState;
    };
    /**
     * Get the value corresponding to a key. Return `defaultValue` if key is not found.
     */
    /**
       * Get the value corresponding to a key. Return `defaultValue` if key is not found.
       */
    TransferState.prototype.get = /**
       * Get the value corresponding to a key. Return `defaultValue` if key is not found.
       */
    function (key, defaultValue) {
        return this.store[key] !== undefined ? this.store[key] : defaultValue;
    };
    /**
     * Set the value corresponding to a key.
     */
    /**
       * Set the value corresponding to a key.
       */
    TransferState.prototype.set = /**
       * Set the value corresponding to a key.
       */
    function (key, value) { this.store[key] = value; };
    /**
     * Remove a key from the store.
     */
    /**
       * Remove a key from the store.
       */
    TransferState.prototype.remove = /**
       * Remove a key from the store.
       */
    function (key) { delete this.store[key]; };
    /**
     * Test whether a key exists in the store.
     */
    /**
       * Test whether a key exists in the store.
       */
    TransferState.prototype.hasKey = /**
       * Test whether a key exists in the store.
       */
    function (key) { return this.store.hasOwnProperty(key); };
    /**
     * Register a callback to provide the value for a key when `toJson` is called.
     */
    /**
       * Register a callback to provide the value for a key when `toJson` is called.
       */
    TransferState.prototype.onSerialize = /**
       * Register a callback to provide the value for a key when `toJson` is called.
       */
    function (key, callback) {
        this.onSerializeCallbacks[key] = callback;
    };
    /**
     * Serialize the current state of the store to JSON.
     */
    /**
       * Serialize the current state of the store to JSON.
       */
    TransferState.prototype.toJson = /**
       * Serialize the current state of the store to JSON.
       */
    function () {
        // Call the onSerialize callbacks and put those values into the store.
        for (var key in this.onSerializeCallbacks) {
            if (this.onSerializeCallbacks.hasOwnProperty(key)) {
                try {
                    this.store[key] = this.onSerializeCallbacks[key]();
                }
                catch (e) {
                    console.warn('Exception in onSerialize callback: ', e);
                }
            }
        }
        return JSON.stringify(this.store);
    };
    TransferState.decorators = [
        { type: core_1.Injectable },
    ];
    return TransferState;
}());
exports.TransferState = TransferState;
function initTransferState(doc, appId) {
    // Locate the script tag with the JSON data transferred from the server.
    // The id of the script tag is set to the Angular appId + 'state'.
    var script = doc.getElementById(appId + '-state');
    var initialState = {};
    if (script && script.textContent) {
        try {
            initialState = JSON.parse(unescapeHtml(script.textContent));
        }
        catch (e) {
            console.warn('Exception while restoring TransferState for app ' + appId, e);
        }
    }
    return TransferState.init(initialState);
}
exports.initTransferState = initTransferState;
/**
 * NgModule to install on the client side while using the `TransferState` to transfer state from
 * server to client.
 *
 * @experimental
 */
var BrowserTransferStateModule = /** @class */ (function () {
    function BrowserTransferStateModule() {
    }
    BrowserTransferStateModule.decorators = [
        { type: core_1.NgModule, args: [{
                    providers: [{ provide: TransferState, useFactory: initTransferState, deps: [dom_tokens_1.DOCUMENT, core_1.APP_ID] }],
                },] },
    ];
    return BrowserTransferStateModule;
}());
exports.BrowserTransferStateModule = BrowserTransferStateModule;
//# sourceMappingURL=transfer_state.js.map