"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.PLATFORM_BROWSER_ID = 'browser';
exports.PLATFORM_SERVER_ID = 'server';
exports.PLATFORM_WORKER_APP_ID = 'browserWorkerApp';
exports.PLATFORM_WORKER_UI_ID = 'browserWorkerUi';
/**
 * Returns whether a platform id represents a browser platform.
 * @experimental
 */
function isPlatformBrowser(platformId) {
    return platformId === exports.PLATFORM_BROWSER_ID;
}
exports.isPlatformBrowser = isPlatformBrowser;
/**
 * Returns whether a platform id represents a server platform.
 * @experimental
 */
function isPlatformServer(platformId) {
    return platformId === exports.PLATFORM_SERVER_ID;
}
exports.isPlatformServer = isPlatformServer;
/**
 * Returns whether a platform id represents a web worker app platform.
 * @experimental
 */
function isPlatformWorkerApp(platformId) {
    return platformId === exports.PLATFORM_WORKER_APP_ID;
}
exports.isPlatformWorkerApp = isPlatformWorkerApp;
/**
 * Returns whether a platform id represents a web worker UI platform.
 * @experimental
 */
function isPlatformWorkerUi(platformId) {
    return platformId === exports.PLATFORM_WORKER_UI_ID;
}
exports.isPlatformWorkerUi = isPlatformWorkerUi;
//# sourceMappingURL=platform_id.js.map