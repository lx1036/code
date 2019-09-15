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
var dom_adapter_1 = require("../dom_adapter");
var dom_tokens_1 = require("../dom_tokens");
var event_manager_1 = require("./event_manager");
/**
 * Defines supported modifiers for key events.
 */
var MODIFIER_KEYS = ['alt', 'control', 'meta', 'shift'];
var ɵ0 = function (event) { return event.altKey; }, ɵ1 = function (event) { return event.ctrlKey; }, ɵ2 = function (event) { return event.metaKey; }, ɵ3 = function (event) { return event.shiftKey; };
exports.ɵ0 = ɵ0;
exports.ɵ1 = ɵ1;
exports.ɵ2 = ɵ2;
exports.ɵ3 = ɵ3;
/**
 * Retrieves modifiers from key-event objects.
 */
var MODIFIER_KEY_GETTERS = {
    'alt': ɵ0,
    'control': ɵ1,
    'meta': ɵ2,
    'shift': ɵ3
};
/**
 * @experimental
 * A browser plug-in that provides support for handling of key events in Angular.
 */
var KeyEventsPlugin = /** @class */ (function (_super) {
    __extends(KeyEventsPlugin, _super);
    /**
     * Initializes an instance of the browser plug-in.
     * @param doc The document in which key events will be detected.
     */
    function KeyEventsPlugin(doc) {
        return _super.call(this, doc) || this;
    }
    /**
      * Reports whether a named key event is supported.
      * @param eventName The event name to query.
      * @return True if the named key event is supported.
     */
    /**
        * Reports whether a named key event is supported.
        * @param eventName The event name to query.
        * @return True if the named key event is supported.
       */
    KeyEventsPlugin.prototype.supports = /**
        * Reports whether a named key event is supported.
        * @param eventName The event name to query.
        * @return True if the named key event is supported.
       */
    function (eventName) { return KeyEventsPlugin.parseEventName(eventName) != null; };
    /**
     * Registers a handler for a specific element and key event.
     * @param element The HTML element to receive event notifications.
     * @param eventName The name of the key event to listen for.
     * @param handler A function to call when the notification occurs. Receives the
     * event object as an argument.
     * @returns The key event that was registered.
    */
    /**
       * Registers a handler for a specific element and key event.
       * @param element The HTML element to receive event notifications.
       * @param eventName The name of the key event to listen for.
       * @param handler A function to call when the notification occurs. Receives the
       * event object as an argument.
       * @returns The key event that was registered.
      */
    KeyEventsPlugin.prototype.addEventListener = /**
       * Registers a handler for a specific element and key event.
       * @param element The HTML element to receive event notifications.
       * @param eventName The name of the key event to listen for.
       * @param handler A function to call when the notification occurs. Receives the
       * event object as an argument.
       * @returns The key event that was registered.
      */
    function (element, eventName, handler) {
        var parsedEvent = (KeyEventsPlugin.parseEventName(eventName));
        var outsideHandler = KeyEventsPlugin.eventCallback(parsedEvent['fullKey'], handler, this.manager.getZone());
        return this.manager.getZone().runOutsideAngular(function () {
            return dom_adapter_1.getDOM().onAndCancel(element, parsedEvent['domEventName'], outsideHandler);
        });
    };
    KeyEventsPlugin.parseEventName = function (eventName) {
        var parts = eventName.toLowerCase().split('.');
        var domEventName = parts.shift();
        if ((parts.length === 0) || !(domEventName === 'keydown' || domEventName === 'keyup')) {
            return null;
        }
        var key = KeyEventsPlugin._normalizeKey((parts.pop()));
        var fullKey = '';
        MODIFIER_KEYS.forEach(function (modifierName) {
            var index = parts.indexOf(modifierName);
            if (index > -1) {
                parts.splice(index, 1);
                fullKey += modifierName + '.';
            }
        });
        fullKey += key;
        if (parts.length != 0 || key.length === 0) {
            // returning null instead of throwing to let another plugin process the event
            return null;
        }
        var result = {};
        result['domEventName'] = domEventName;
        result['fullKey'] = fullKey;
        return result;
    };
    KeyEventsPlugin.getEventFullKey = function (event) {
        var fullKey = '';
        var key = dom_adapter_1.getDOM().getEventKey(event);
        key = key.toLowerCase();
        if (key === ' ') {
            key = 'space'; // for readability
        }
        else if (key === '.') {
            key = 'dot'; // because '.' is used as a separator in event names
        }
        MODIFIER_KEYS.forEach(function (modifierName) {
            if (modifierName != key) {
                var modifierGetter = MODIFIER_KEY_GETTERS[modifierName];
                if (modifierGetter(event)) {
                    fullKey += modifierName + '.';
                }
            }
        });
        fullKey += key;
        return fullKey;
    };
    /**
     * Configures a handler callback for a key event.
     * @param fullKey The event name that combines all simultaneous keystrokes.
     * @param handler The function that responds to the key event.
     * @param zone The zone in which the event occurred.
     * @returns A callback function.
     */
    /**
       * Configures a handler callback for a key event.
       * @param fullKey The event name that combines all simultaneous keystrokes.
       * @param handler The function that responds to the key event.
       * @param zone The zone in which the event occurred.
       * @returns A callback function.
       */
    KeyEventsPlugin.eventCallback = /**
       * Configures a handler callback for a key event.
       * @param fullKey The event name that combines all simultaneous keystrokes.
       * @param handler The function that responds to the key event.
       * @param zone The zone in which the event occurred.
       * @returns A callback function.
       */
    function (fullKey, handler, zone) {
        return function (event /** TODO #9100 */) {
            if (KeyEventsPlugin.getEventFullKey(event) === fullKey) {
                zone.runGuarded(function () { return handler(event); });
            }
        };
    };
    /** @internal */
    /** @internal */
    KeyEventsPlugin._normalizeKey = /** @internal */
    function (keyName) {
        // TODO: switch to a Map if the mapping grows too much
        switch (keyName) {
            case 'esc':
                return 'escape';
            default:
                return keyName;
        }
    };
    KeyEventsPlugin.decorators = [
        { type: core_1.Injectable },
    ];
    /** @nocollapse */
    KeyEventsPlugin.ctorParameters = function () { return [
        { type: undefined, decorators: [{ type: core_1.Inject, args: [dom_tokens_1.DOCUMENT,] },] },
    ]; };
    return KeyEventsPlugin;
}(event_manager_1.EventManagerPlugin));
exports.KeyEventsPlugin = KeyEventsPlugin;
//# sourceMappingURL=key_events.js.map