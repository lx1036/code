"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var lang_1 = require("../src/util/lang");
var di_1 = require("./di");
/**
 * A function that will be executed when an application is initialized.
 * @experimental
 */
exports.APP_INITIALIZER = new di_1.InjectionToken('Application Initializer');
/**
 * A class that reflects the state of running {@link APP_INITIALIZER}s.
 *
 * @experimental
 */
var ApplicationInitStatus = /** @class */ (function () {
    function ApplicationInitStatus(appInits) {
        var _this = this;
        this.appInits = appInits;
        this.initialized = false;
        this.done = false;
        this.donePromise = new Promise(function (res, rej) {
            _this.resolve = res;
            _this.reject = rej;
        });
    }
    /** @internal */
    /** @internal */
    ApplicationInitStatus.prototype.runInitializers = /** @internal */
    function () {
        var _this = this;
        if (this.initialized) {
            return;
        }
        var asyncInitPromises = [];
        var complete = function () {
            _this.done = true;
            _this.resolve();
        };
        if (this.appInits) {
            for (var i = 0; i < this.appInits.length; i++) {
                var initResult = this.appInits[i]();
                if (lang_1.isPromise(initResult)) {
                    asyncInitPromises.push(initResult);
                }
            }
        }
        Promise.all(asyncInitPromises).then(function () { complete(); }).catch(function (e) { _this.reject(e); });
        if (asyncInitPromises.length === 0) {
            complete();
        }
        this.initialized = true;
    };
    ApplicationInitStatus.decorators = [
        { type: di_1.Injectable },
    ];
    /** @nocollapse */
    ApplicationInitStatus.ctorParameters = function () { return [
        { type: Array, decorators: [{ type: di_1.Inject, args: [exports.APP_INITIALIZER,] }, { type: di_1.Optional },] },
    ]; };
    return ApplicationInitStatus;
}());
exports.ApplicationInitStatus = ApplicationInitStatus;
//# sourceMappingURL=application_init.js.map