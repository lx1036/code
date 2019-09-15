"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var di_1 = require("../di");
var util_1 = require("../util");
var ng_zone_1 = require("../zone/ng_zone");
/**
 * The Testability service provides testing hooks that can be accessed from
 * the browser and by services such as Protractor. Each bootstrapped Angular
 * application on the page will have an instance of Testability.
 * @experimental
 */
var Testability = /** @class */ (function () {
    function Testability(_ngZone) {
        var _this = this;
        this._ngZone = _ngZone;
        this._pendingCount = 0;
        this._isZoneStable = true;
        /**
           * Whether any work was done since the last 'whenStable' callback. This is
           * useful to detect if this could have potentially destabilized another
           * component while it is stabilizing.
           * @internal
           */
        this._didWork = false;
        this._callbacks = [];
        this._watchAngularEvents();
        _ngZone.run(function () { _this.taskTrackingZone = Zone.current.get('TaskTrackingZone'); });
    }
    Testability.prototype._watchAngularEvents = function () {
        var _this = this;
        this._ngZone.onUnstable.subscribe({
            next: function () {
                _this._didWork = true;
                _this._isZoneStable = false;
            }
        });
        this._ngZone.runOutsideAngular(function () {
            _this._ngZone.onStable.subscribe({
                next: function () {
                    ng_zone_1.NgZone.assertNotInAngularZone();
                    util_1.scheduleMicroTask(function () {
                        _this._isZoneStable = true;
                        _this._runCallbacksIfReady();
                    });
                }
            });
        });
    };
    /**
     * Increases the number of pending request
     * @deprecated pending requests are now tracked with zones.
     */
    /**
       * Increases the number of pending request
       * @deprecated pending requests are now tracked with zones.
       */
    Testability.prototype.increasePendingRequestCount = /**
       * Increases the number of pending request
       * @deprecated pending requests are now tracked with zones.
       */
    function () {
        this._pendingCount += 1;
        this._didWork = true;
        return this._pendingCount;
    };
    /**
     * Decreases the number of pending request
     * @deprecated pending requests are now tracked with zones
     */
    /**
       * Decreases the number of pending request
       * @deprecated pending requests are now tracked with zones
       */
    Testability.prototype.decreasePendingRequestCount = /**
       * Decreases the number of pending request
       * @deprecated pending requests are now tracked with zones
       */
    function () {
        this._pendingCount -= 1;
        if (this._pendingCount < 0) {
            throw new Error('pending async requests below zero');
        }
        this._runCallbacksIfReady();
        return this._pendingCount;
    };
    /**
     * Whether an associated application is stable
     */
    /**
       * Whether an associated application is stable
       */
    Testability.prototype.isStable = /**
       * Whether an associated application is stable
       */
    function () {
        return this._isZoneStable && this._pendingCount === 0 && !this._ngZone.hasPendingMacrotasks;
    };
    Testability.prototype._runCallbacksIfReady = function () {
        var _this = this;
        if (this.isStable()) {
            // Schedules the call backs in a new frame so that it is always async.
            // Schedules the call backs in a new frame so that it is always async.
            util_1.scheduleMicroTask(function () {
                while (_this._callbacks.length !== 0) {
                    var cb = (_this._callbacks.pop());
                    clearTimeout(cb.timeoutId);
                    cb.doneCb(_this._didWork);
                }
                _this._didWork = false;
            });
        }
        else {
            // Still not stable, send updates.
            var pending_1 = this.getPendingTasks();
            this._callbacks = this._callbacks.filter(function (cb) {
                if (cb.updateCb && cb.updateCb(pending_1)) {
                    clearTimeout(cb.timeoutId);
                    return false;
                }
                return true;
            });
            this._didWork = true;
        }
    };
    Testability.prototype.getPendingTasks = function () {
        if (!this.taskTrackingZone) {
            return [];
        }
        return this.taskTrackingZone.macroTasks.map(function (t) {
            return {
                source: t.source,
                isPeriodic: t.data.isPeriodic,
                delay: t.data.delay,
                // From TaskTrackingZone:
                // https://github.com/angular/zone.js/blob/master/lib/zone-spec/task-tracking.ts#L40
                creationLocation: t.creationLocation,
                // Added by Zones for XHRs
                // https://github.com/angular/zone.js/blob/master/lib/browser/browser.ts#L133
                xhr: t.data.target
            };
        });
    };
    Testability.prototype.addCallback = function (cb, timeout, updateCb) {
        var _this = this;
        var timeoutId = -1;
        if (timeout && timeout > 0) {
            timeoutId = setTimeout(function () {
                _this._callbacks = _this._callbacks.filter(function (cb) { return cb.timeoutId !== timeoutId; });
                cb(_this._didWork, _this.getPendingTasks());
            }, timeout);
        }
        this._callbacks.push({ doneCb: cb, timeoutId: timeoutId, updateCb: updateCb });
    };
    /**
     * Wait for the application to be stable with a timeout. If the timeout is reached before that
     * happens, the callback receives a list of the macro tasks that were pending, otherwise null.
     *
     * @param doneCb The callback to invoke when Angular is stable or the timeout expires
     *    whichever comes first.
     * @param timeout Optional. The maximum time to wait for Angular to become stable. If not
     *    specified, whenStable() will wait forever.
     * @param updateCb Optional. If specified, this callback will be invoked whenever the set of
     *    pending macrotasks changes. If this callback returns true doneCb will not be invoked
     *    and no further updates will be issued.
     */
    /**
       * Wait for the application to be stable with a timeout. If the timeout is reached before that
       * happens, the callback receives a list of the macro tasks that were pending, otherwise null.
       *
       * @param doneCb The callback to invoke when Angular is stable or the timeout expires
       *    whichever comes first.
       * @param timeout Optional. The maximum time to wait for Angular to become stable. If not
       *    specified, whenStable() will wait forever.
       * @param updateCb Optional. If specified, this callback will be invoked whenever the set of
       *    pending macrotasks changes. If this callback returns true doneCb will not be invoked
       *    and no further updates will be issued.
       */
    Testability.prototype.whenStable = /**
       * Wait for the application to be stable with a timeout. If the timeout is reached before that
       * happens, the callback receives a list of the macro tasks that were pending, otherwise null.
       *
       * @param doneCb The callback to invoke when Angular is stable or the timeout expires
       *    whichever comes first.
       * @param timeout Optional. The maximum time to wait for Angular to become stable. If not
       *    specified, whenStable() will wait forever.
       * @param updateCb Optional. If specified, this callback will be invoked whenever the set of
       *    pending macrotasks changes. If this callback returns true doneCb will not be invoked
       *    and no further updates will be issued.
       */
    function (doneCb, timeout, updateCb) {
        if (updateCb && !this.taskTrackingZone) {
            throw new Error('Task tracking zone is required when passing an update callback to ' +
                'whenStable(). Is "zone.js/dist/task-tracking.js" loaded?');
        }
        // These arguments are 'Function' above to keep the public API simple.
        this.addCallback(doneCb, timeout, updateCb);
        this._runCallbacksIfReady();
    };
    /**
     * Get the number of pending requests
     * @deprecated pending requests are now tracked with zones
     */
    /**
       * Get the number of pending requests
       * @deprecated pending requests are now tracked with zones
       */
    Testability.prototype.getPendingRequestCount = /**
       * Get the number of pending requests
       * @deprecated pending requests are now tracked with zones
       */
    function () { return this._pendingCount; };
    /**
     * Find providers by name
     * @param using The root element to search from
     * @param provider The name of binding variable
     * @param exactMatch Whether using exactMatch
     */
    /**
       * Find providers by name
       * @param using The root element to search from
       * @param provider The name of binding variable
       * @param exactMatch Whether using exactMatch
       */
    Testability.prototype.findProviders = /**
       * Find providers by name
       * @param using The root element to search from
       * @param provider The name of binding variable
       * @param exactMatch Whether using exactMatch
       */
    function (using, provider, exactMatch) {
        // TODO(juliemr): implement.
        return [];
    };
    Testability.decorators = [
        { type: di_1.Injectable },
    ];
    /** @nocollapse */
    Testability.ctorParameters = function () { return [
        { type: ng_zone_1.NgZone, },
    ]; };
    return Testability;
}());
exports.Testability = Testability;
/**
 * A global registry of {@link Testability} instances for specific elements.
 * @experimental
 */
var TestabilityRegistry = /** @class */ (function () {
    function TestabilityRegistry() {
        /** @internal */
        this._applications = new Map();
        _testabilityGetter.addToWindow(this);
    }
    /**
     * Registers an application with a testability hook so that it can be tracked
     * @param token token of application, root element
     * @param testability Testability hook
     */
    /**
       * Registers an application with a testability hook so that it can be tracked
       * @param token token of application, root element
       * @param testability Testability hook
       */
    TestabilityRegistry.prototype.registerApplication = /**
       * Registers an application with a testability hook so that it can be tracked
       * @param token token of application, root element
       * @param testability Testability hook
       */
    function (token, testability) {
        this._applications.set(token, testability);
    };
    /**
     * Unregisters an application.
     * @param token token of application, root element
     */
    /**
       * Unregisters an application.
       * @param token token of application, root element
       */
    TestabilityRegistry.prototype.unregisterApplication = /**
       * Unregisters an application.
       * @param token token of application, root element
       */
    function (token) { this._applications.delete(token); };
    /**
     * Unregisters all applications
     */
    /**
       * Unregisters all applications
       */
    TestabilityRegistry.prototype.unregisterAllApplications = /**
       * Unregisters all applications
       */
    function () { this._applications.clear(); };
    /**
     * Get a testability hook associated with the application
     * @param elem root element
     */
    /**
       * Get a testability hook associated with the application
       * @param elem root element
       */
    TestabilityRegistry.prototype.getTestability = /**
       * Get a testability hook associated with the application
       * @param elem root element
       */
    function (elem) { return this._applications.get(elem) || null; };
    /**
     * Get all registered testabilities
     */
    /**
       * Get all registered testabilities
       */
    TestabilityRegistry.prototype.getAllTestabilities = /**
       * Get all registered testabilities
       */
    function () { return Array.from(this._applications.values()); };
    /**
     * Get all registered applications(root elements)
     */
    /**
       * Get all registered applications(root elements)
       */
    TestabilityRegistry.prototype.getAllRootElements = /**
       * Get all registered applications(root elements)
       */
    function () { return Array.from(this._applications.keys()); };
    /**
     * Find testability of a node in the Tree
     * @param elem node
     * @param findInAncestors whether finding testability in ancestors if testability was not found in
     * current node
     */
    /**
       * Find testability of a node in the Tree
       * @param elem node
       * @param findInAncestors whether finding testability in ancestors if testability was not found in
       * current node
       */
    TestabilityRegistry.prototype.findTestabilityInTree = /**
       * Find testability of a node in the Tree
       * @param elem node
       * @param findInAncestors whether finding testability in ancestors if testability was not found in
       * current node
       */
    function (elem, findInAncestors) {
        if (findInAncestors === void 0) { findInAncestors = true; }
        return _testabilityGetter.findTestabilityInTree(this, elem, findInAncestors);
    };
    TestabilityRegistry.decorators = [
        { type: di_1.Injectable },
    ];
    /** @nocollapse */
    TestabilityRegistry.ctorParameters = function () { return []; };
    return TestabilityRegistry;
}());
exports.TestabilityRegistry = TestabilityRegistry;
var _NoopGetTestability = /** @class */ (function () {
    function _NoopGetTestability() {
    }
    _NoopGetTestability.prototype.addToWindow = function (registry) { };
    _NoopGetTestability.prototype.findTestabilityInTree = function (registry, elem, findInAncestors) {
        return null;
    };
    return _NoopGetTestability;
}());
/**
 * Set the {@link GetTestability} implementation used by the Angular testing framework.
 * @experimental
 */
function setTestabilityGetter(getter) {
    _testabilityGetter = getter;
}
exports.setTestabilityGetter = setTestabilityGetter;
var _testabilityGetter = new _NoopGetTestability();
//# sourceMappingURL=testability.js.map