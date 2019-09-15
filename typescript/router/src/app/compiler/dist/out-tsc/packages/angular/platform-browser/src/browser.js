"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var common_1 = require("@angular/common");
var core_1 = require("@angular/core");
var browser_adapter_1 = require("./browser/browser_adapter");
var browser_platform_location_1 = require("./browser/location/browser_platform_location");
var meta_1 = require("./browser/meta");
var server_transition_1 = require("./browser/server-transition");
var testability_1 = require("./browser/testability");
var title_1 = require("./browser/title");
var ng_probe_1 = require("./dom/debug/ng_probe");
var dom_renderer_1 = require("./dom/dom_renderer");
var dom_tokens_1 = require("./dom/dom_tokens");
var dom_events_1 = require("./dom/events/dom_events");
var event_manager_1 = require("./dom/events/event_manager");
var hammer_gestures_1 = require("./dom/events/hammer_gestures");
var key_events_1 = require("./dom/events/key_events");
var shared_styles_host_1 = require("./dom/shared_styles_host");
var dom_sanitization_service_1 = require("./security/dom_sanitization_service");
exports.INTERNAL_BROWSER_PLATFORM_PROVIDERS = [
    { provide: core_1.PLATFORM_ID, useValue: common_1.ɵPLATFORM_BROWSER_ID },
    { provide: core_1.PLATFORM_INITIALIZER, useValue: initDomAdapter, multi: true },
    { provide: common_1.PlatformLocation, useClass: browser_platform_location_1.BrowserPlatformLocation, deps: [dom_tokens_1.DOCUMENT] },
    { provide: dom_tokens_1.DOCUMENT, useFactory: _document, deps: [] },
];
/**
 * @security Replacing built-in sanitization providers exposes the application to XSS risks.
 * Attacker-controlled data introduced by an unsanitized provider could expose your
 * application to XSS risks. For more detail, see the [Security Guide](http://g.co/ng/security).
 * @experimental
 */
exports.BROWSER_SANITIZATION_PROVIDERS = [
    { provide: core_1.Sanitizer, useExisting: dom_sanitization_service_1.DomSanitizer },
    { provide: dom_sanitization_service_1.DomSanitizer, useClass: dom_sanitization_service_1.DomSanitizerImpl, deps: [dom_tokens_1.DOCUMENT] },
];
exports.platformBrowser = core_1.createPlatformFactory(core_1.platformCore, 'browser', exports.INTERNAL_BROWSER_PLATFORM_PROVIDERS);
function initDomAdapter() {
    browser_adapter_1.BrowserDomAdapter.makeCurrent();
    testability_1.BrowserGetTestability.init();
}
exports.initDomAdapter = initDomAdapter;
function errorHandler() {
    return new core_1.ErrorHandler();
}
exports.errorHandler = errorHandler;
function _document() {
    return document;
}
exports._document = _document;
/**
 * The ng module for the browser.
 *
 *
 */
var BrowserModule = /** @class */ (function () {
    function BrowserModule(parentModule) {
        if (parentModule) {
            throw new Error("BrowserModule has already been loaded. If you need access to common directives such as NgIf and NgFor from a lazy loaded module, import CommonModule instead.");
        }
    }
    /**
     * Configures a browser-based application to transition from a server-rendered app, if
     * one is present on the page. The specified parameters must include an application id,
     * which must match between the client and server applications.
     *
     * @experimental
     */
    /**
       * Configures a browser-based application to transition from a server-rendered app, if
       * one is present on the page. The specified parameters must include an application id,
       * which must match between the client and server applications.
       *
       * @experimental
       */
    BrowserModule.withServerTransition = /**
       * Configures a browser-based application to transition from a server-rendered app, if
       * one is present on the page. The specified parameters must include an application id,
       * which must match between the client and server applications.
       *
       * @experimental
       */
    function (params) {
        return {
            ngModule: BrowserModule,
            providers: [
                { provide: core_1.APP_ID, useValue: params.appId },
                { provide: server_transition_1.TRANSITION_ID, useExisting: core_1.APP_ID },
                server_transition_1.SERVER_TRANSITION_PROVIDERS,
            ],
        };
    };
    BrowserModule.decorators = [
        { type: core_1.NgModule, args: [{
                    providers: [
                        exports.BROWSER_SANITIZATION_PROVIDERS,
                        { provide: core_1.ɵAPP_ROOT, useValue: true },
                        { provide: core_1.ErrorHandler, useFactory: errorHandler, deps: [] },
                        { provide: event_manager_1.EVENT_MANAGER_PLUGINS, useClass: dom_events_1.DomEventsPlugin, multi: true },
                        { provide: event_manager_1.EVENT_MANAGER_PLUGINS, useClass: key_events_1.KeyEventsPlugin, multi: true },
                        { provide: event_manager_1.EVENT_MANAGER_PLUGINS, useClass: hammer_gestures_1.HammerGesturesPlugin, multi: true },
                        { provide: hammer_gestures_1.HAMMER_GESTURE_CONFIG, useClass: hammer_gestures_1.HammerGestureConfig },
                        dom_renderer_1.DomRendererFactory2,
                        { provide: core_1.RendererFactory2, useExisting: dom_renderer_1.DomRendererFactory2 },
                        { provide: shared_styles_host_1.SharedStylesHost, useExisting: shared_styles_host_1.DomSharedStylesHost },
                        shared_styles_host_1.DomSharedStylesHost,
                        core_1.Testability,
                        event_manager_1.EventManager,
                        ng_probe_1.ELEMENT_PROBE_PROVIDERS,
                        meta_1.Meta,
                        title_1.Title,
                    ],
                    exports: [common_1.CommonModule, core_1.ApplicationModule]
                },] },
    ];
    /** @nocollapse */
    BrowserModule.ctorParameters = function () { return [
        { type: BrowserModule, decorators: [{ type: core_1.Optional }, { type: core_1.SkipSelf },] },
    ]; };
    return BrowserModule;
}());
exports.BrowserModule = BrowserModule;
//# sourceMappingURL=browser.js.map