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
var dom_adapter_1 = require("../dom_adapter");
/**
 * The injection token for the event-manager plug-in service.
 */
exports.EVENT_MANAGER_PLUGINS = new core_1.InjectionToken('EventManagerPlugins');
/**
 * An injectable service that provides event management for Angular
 * through a browser plug-in.
 */
var EventManager = /** @class */ (function () {
    /**
     * Initializes an instance of the event-manager service.
     */
    function EventManager(plugins, _zone) {
        var _this = this;
        this._zone = _zone;
        this._eventNameToPlugin = new Map();
        plugins.forEach(function (p) { return p.manager = _this; });
        this._plugins = plugins.slice().reverse();
    }
    /**
     * Registers a handler for a specific element and event.
     *
     * @param element The HTML element to receive event notifications.
     * @param eventName The name of the event to listen for.
     * @param handler A function to call when the notification occurs. Receives the
     * event object as an argument.
     * @returns  A callback function that can be used to remove the handler.
     */
    /**
       * Registers a handler for a specific element and event.
       *
       * @param element The HTML element to receive event notifications.
       * @param eventName The name of the event to listen for.
       * @param handler A function to call when the notification occurs. Receives the
       * event object as an argument.
       * @returns  A callback function that can be used to remove the handler.
       */
    EventManager.prototype.addEventListener = /**
       * Registers a handler for a specific element and event.
       *
       * @param element The HTML element to receive event notifications.
       * @param eventName The name of the event to listen for.
       * @param handler A function to call when the notification occurs. Receives the
       * event object as an argument.
       * @returns  A callback function that can be used to remove the handler.
       */
    function (element, eventName, handler) {
        var plugin = this._findPluginFor(eventName);
        return plugin.addEventListener(element, eventName, handler);
    };
    /**
     * Registers a global handler for an event in a target view.
     *
     * @param target A target for global event notifications. One of "window", "document", or "body".
     * @param eventName The name of the event to listen for.
     * @param handler A function to call when the notification occurs. Receives the
     * event object as an argument.
     * @returns A callback function that can be used to remove the handler.
     */
    /**
       * Registers a global handler for an event in a target view.
       *
       * @param target A target for global event notifications. One of "window", "document", or "body".
       * @param eventName The name of the event to listen for.
       * @param handler A function to call when the notification occurs. Receives the
       * event object as an argument.
       * @returns A callback function that can be used to remove the handler.
       */
    EventManager.prototype.addGlobalEventListener = /**
       * Registers a global handler for an event in a target view.
       *
       * @param target A target for global event notifications. One of "window", "document", or "body".
       * @param eventName The name of the event to listen for.
       * @param handler A function to call when the notification occurs. Receives the
       * event object as an argument.
       * @returns A callback function that can be used to remove the handler.
       */
    function (target, eventName, handler) {
        var plugin = this._findPluginFor(eventName);
        return plugin.addGlobalEventListener(target, eventName, handler);
    };
    /**
     * Retrieves the compilation zone in which event listeners are registered.
     */
    /**
       * Retrieves the compilation zone in which event listeners are registered.
       */
    EventManager.prototype.getZone = /**
       * Retrieves the compilation zone in which event listeners are registered.
       */
    function () { return this._zone; };
    /** @internal */
    /** @internal */
    EventManager.prototype._findPluginFor = /** @internal */
    function (eventName) {
        var plugin = this._eventNameToPlugin.get(eventName);
        if (plugin) {
            return plugin;
        }
        var plugins = this._plugins;
        for (var i = 0; i < plugins.length; i++) {
            var plugin_1 = plugins[i];
            if (plugin_1.supports(eventName)) {
                this._eventNameToPlugin.set(eventName, plugin_1);
                return plugin_1;
            }
        }
        throw new Error("No event manager plugin found for event " + eventName);
    };
    EventManager.decorators = [
        { type: core_1.Injectable },
    ];
    /** @nocollapse */
    EventManager.ctorParameters = function () { return [
        { type: Array, decorators: [{ type: core_1.Inject, args: [exports.EVENT_MANAGER_PLUGINS,] },] },
        { type: core_1.NgZone, },
    ]; };
    return EventManager;
}());
exports.EventManager = EventManager;
var EventManagerPlugin = /** @class */ (function () {
    function EventManagerPlugin(_doc) {
        this._doc = _doc;
    }
    EventManagerPlugin.prototype.addGlobalEventListener = function (element, eventName, handler) {
        var target = dom_adapter_1.getDOM().getGlobalEventTarget(this._doc, element);
        if (!target) {
            throw new Error("Unsupported event target " + target + " for event " + eventName);
        }
        return this.addEventListener(target, eventName, handler);
    };
    return EventManagerPlugin;
}());
exports.EventManagerPlugin = EventManagerPlugin;
//# sourceMappingURL=event_manager.js.map