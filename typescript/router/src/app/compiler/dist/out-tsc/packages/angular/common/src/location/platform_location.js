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
/**
 * This class should not be used directly by an application developer. Instead, use
 * {@link Location}.
 *
 * `PlatformLocation` encapsulates all calls to DOM apis, which allows the Router to be platform
 * agnostic.
 * This means that we can have different implementation of `PlatformLocation` for the different
 * platforms that angular supports. For example, `@angular/platform-browser` provides an
 * implementation specific to the browser environment, while `@angular/platform-webworker` provides
 * one suitable for use with web workers.
 *
 * The `PlatformLocation` class is used directly by all implementations of {@link LocationStrategy}
 * when they need to interact with the DOM apis like pushState, popState, etc...
 *
 * {@link LocationStrategy} in turn is used by the {@link Location} service which is used directly
 * by the {@link Router} in order to navigate between routes. Since all interactions between {@link
 * Router} /
 * {@link Location} / {@link LocationStrategy} and DOM apis flow through the `PlatformLocation`
 * class they are all platform independent.
 *
 *
 */
var /**
 * This class should not be used directly by an application developer. Instead, use
 * {@link Location}.
 *
 * `PlatformLocation` encapsulates all calls to DOM apis, which allows the Router to be platform
 * agnostic.
 * This means that we can have different implementation of `PlatformLocation` for the different
 * platforms that angular supports. For example, `@angular/platform-browser` provides an
 * implementation specific to the browser environment, while `@angular/platform-webworker` provides
 * one suitable for use with web workers.
 *
 * The `PlatformLocation` class is used directly by all implementations of {@link LocationStrategy}
 * when they need to interact with the DOM apis like pushState, popState, etc...
 *
 * {@link LocationStrategy} in turn is used by the {@link Location} service which is used directly
 * by the {@link Router} in order to navigate between routes. Since all interactions between {@link
 * Router} /
 * {@link Location} / {@link LocationStrategy} and DOM apis flow through the `PlatformLocation`
 * class they are all platform independent.
 *
 *
 */
PlatformLocation = /** @class */ (function () {
    function PlatformLocation() {
    }
    return PlatformLocation;
}());
exports.PlatformLocation = PlatformLocation;
/**
 * @description Indicates when a location is initialized.
 * @experimental
 */
exports.LOCATION_INITIALIZED = new core_1.InjectionToken('Location Initialized');
//# sourceMappingURL=platform_location.js.map