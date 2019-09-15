"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = Object.setPrototypeOf ||
        ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
        function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var dom_tokens_1 = require("../dom_tokens");
var event_manager_1 = require("./event_manager");
/**
 * Supported HammerJS recognizer event names.
 */
var EVENT_NAMES = {
    // pan
    'pan': true,
    'panstart': true,
    'panmove': true,
    'panend': true,
    'pancancel': true,
    'panleft': true,
    'panright': true,
    'panup': true,
    'pandown': true,
    // pinch
    'pinch': true,
    'pinchstart': true,
    'pinchmove': true,
    'pinchend': true,
    'pinchcancel': true,
    'pinchin': true,
    'pinchout': true,
    // press
    'press': true,
    'pressup': true,
    // rotate
    'rotate': true,
    'rotatestart': true,
    'rotatemove': true,
    'rotateend': true,
    'rotatecancel': true,
    // swipe
    'swipe': true,
    'swipeleft': true,
    'swiperight': true,
    'swipeup': true,
    'swipedown': true,
    // tap
    'tap': true,
};
/**
 * DI token for providing [HammerJS](http://hammerjs.github.io/) support to Angular.
 * @see `HammerGestureConfig`
 *
 * @experimental
 */
exports.HAMMER_GESTURE_CONFIG = new core_1.InjectionToken('HammerGestureConfig');
/**
 * An injectable [HammerJS Manager](http://hammerjs.github.io/api/#hammer.manager)
 * for gesture recognition. Configures specific event recognition.
 * @experimental
 */
var HammerGestureConfig = /** @class */ (function () {
    function HammerGestureConfig() {
        /**
           * A set of supported event names for gestures to be used in Angular.
           * Angular supports all built-in recognizers, as listed in
           * [HammerJS documentation](http://hammerjs.github.io/).
           */
        this.events = [];
        /**
          * Maps gesture event names to a set of configuration options
          * that specify overrides to the default values for specific properties.
          *
          * The key is a supported event name to be configured,
          * and the options object contains a set of properties, with override values
          * to be applied to the named recognizer event.
          * For example, to disable recognition of the rotate event, specify
          *  `{"rotate": {"enable": false}}`.
          *
          * Properties that are not present take the HammerJS default values.
          * For information about which properties are supported for which events,
          * and their allowed and default values, see
          * [HammerJS documentation](http://hammerjs.github.io/).
          *
          */
        this.overrides = {};
    }
    /**
     * Creates a [HammerJS Manager](http://hammerjs.github.io/api/#hammer.manager)
     * and attaches it to a given HTML element.
     * @param element The element that will recognize gestures.
     * @returns A HammerJS event-manager object.
     */
    /**
       * Creates a [HammerJS Manager](http://hammerjs.github.io/api/#hammer.manager)
       * and attaches it to a given HTML element.
       * @param element The element that will recognize gestures.
       * @returns A HammerJS event-manager object.
       */
    HammerGestureConfig.prototype.buildHammer = /**
       * Creates a [HammerJS Manager](http://hammerjs.github.io/api/#hammer.manager)
       * and attaches it to a given HTML element.
       * @param element The element that will recognize gestures.
       * @returns A HammerJS event-manager object.
       */
    function (element) {
        var mc = new Hammer(element, this.options);
        mc.get('pinch').set({ enable: true });
        mc.get('rotate').set({ enable: true });
        for (var eventName in this.overrides) {
            mc.get(eventName).set(this.overrides[eventName]);
        }
        return mc;
    };
    HammerGestureConfig.decorators = [
        { type: core_1.Injectable },
    ];
    return HammerGestureConfig;
}());
exports.HammerGestureConfig = HammerGestureConfig;
var HammerGesturesPlugin = /** @class */ (function (_super) {
    __extends(HammerGesturesPlugin, _super);
    function HammerGesturesPlugin(doc, _config, console) {
        var _this = _super.call(this, doc) || this;
        _this._config = _config;
        _this.console = console;
        return _this;
    }
    HammerGesturesPlugin.prototype.supports = function (eventName) {
        if (!EVENT_NAMES.hasOwnProperty(eventName.toLowerCase()) && !this.isCustomEvent(eventName)) {
            return false;
        }
        if (!window.Hammer) {
            this.console.warn("Hammer.js is not loaded, can not bind '" + eventName + "' event.");
            return false;
        }
        return true;
    };
    HammerGesturesPlugin.prototype.addEventListener = function (element, eventName, handler) {
        var _this = this;
        var zone = this.manager.getZone();
        eventName = eventName.toLowerCase();
        return zone.runOutsideAngular(function () {
            // Creating the manager bind events, must be done outside of angular
            var mc = _this._config.buildHammer(element);
            var callback = function (eventObj) {
                zone.runGuarded(function () { handler(eventObj); });
            };
            mc.on(eventName, callback);
            return function () { return mc.off(eventName, callback); };
        });
    };
    HammerGesturesPlugin.prototype.isCustomEvent = function (eventName) { return this._config.events.indexOf(eventName) > -1; };
    HammerGesturesPlugin.decorators = [
        { type: core_1.Injectable },
    ];
    /** @nocollapse */
    HammerGesturesPlugin.ctorParameters = function () { return [
        { type: undefined, decorators: [{ type: core_1.Inject, args: [dom_tokens_1.DOCUMENT,] },] },
        { type: HammerGestureConfig, decorators: [{ type: core_1.Inject, args: [exports.HAMMER_GESTURE_CONFIG,] },] },
        { type: core_1.ɵConsole, },
    ]; };
    return HammerGesturesPlugin;
}(event_manager_1.EventManagerPlugin));
exports.HammerGesturesPlugin = HammerGesturesPlugin;
//# sourceMappingURL=hammer_gestures.js.map