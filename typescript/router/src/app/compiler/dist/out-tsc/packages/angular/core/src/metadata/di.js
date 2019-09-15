"use strict";
var __assign = (this && this.__assign) || Object.assign || function(t) {
    for (var s, i = 1, n = arguments.length; i < n; i++) {
        s = arguments[i];
        for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p))
            t[p] = s[p];
    }
    return t;
};
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var injection_token_1 = require("../di/injection_token");
var decorators_1 = require("../util/decorators");
/**
 * This token can be used to create a virtual provider that will populate the
 * `entryComponents` fields of components and ng modules based on its `useValue`.
 * All components that are referenced in the `useValue` value (either directly
 * or in a nested array or map) will be added to the `entryComponents` property.
 *
 * ### Example
 * The following example shows how the router can populate the `entryComponents`
 * field of an NgModule based on the router configuration which refers
 * to components.
 *
 * ```typescript
 * // helper function inside the router
 * function provideRoutes(routes) {
 *   return [
 *     {provide: ROUTES, useValue: routes},
 *     {provide: ANALYZE_FOR_ENTRY_COMPONENTS, useValue: routes, multi: true}
 *   ];
 * }
 *
 * // user code
 * let routes = [
 *   {path: '/root', component: RootComp},
 *   {path: '/teams', component: TeamsComp}
 * ];
 *
 * @NgModule({
 *   providers: [provideRoutes(routes)]
 * })
 * class ModuleWithRoutes {}
 * ```
 *
 * @experimental
 */
exports.ANALYZE_FOR_ENTRY_COMPONENTS = new injection_token_1.InjectionToken('AnalyzeForEntryComponents');
/**
 * Attribute decorator and metadata.
 *
 *
 * @Annotation
 */
exports.Attribute = decorators_1.makeParamDecorator('Attribute', function (attributeName) { return ({ attributeName: attributeName }); });
/**
 * Base class for query metadata.
 *
 * See {@link ContentChildren}, {@link ContentChild}, {@link ViewChildren}, {@link ViewChild} for
 * more information.
 *
 *
 */
var /**
 * Base class for query metadata.
 *
 * See {@link ContentChildren}, {@link ContentChild}, {@link ViewChildren}, {@link ViewChild} for
 * more information.
 *
 *
 */
Query = /** @class */ (function () {
    function Query() {
    }
    return Query;
}());
exports.Query = Query;
/**
 * ContentChildren decorator and metadata.
 *
 *
 *  @Annotation
 */
exports.ContentChildren = decorators_1.makePropDecorator('ContentChildren', function (selector, data) {
    if (data === void 0) { data = {}; }
    return (__assign({ selector: selector, first: false, isViewQuery: false, descendants: false }, data));
}, Query);
/**
 * ContentChild decorator and metadata.
 *
 *
 * @Annotation
 */
exports.ContentChild = decorators_1.makePropDecorator('ContentChild', function (selector, data) {
    if (data === void 0) { data = {}; }
    return (__assign({ selector: selector, first: true, isViewQuery: false, descendants: true }, data));
}, Query);
/**
 * ViewChildren decorator and metadata.
 *
 *
 * @Annotation
 */
exports.ViewChildren = decorators_1.makePropDecorator('ViewChildren', function (selector, data) {
    if (data === void 0) { data = {}; }
    return (__assign({ selector: selector, first: false, isViewQuery: true, descendants: true }, data));
}, Query);
/**
 * ViewChild decorator and metadata.
 *
 *
 * @Annotation
 */
exports.ViewChild = decorators_1.makePropDecorator('ViewChild', function (selector, data) {
    return (__assign({ selector: selector, first: true, isViewQuery: true, descendants: true }, data));
}, Query);
//# sourceMappingURL=di.js.map