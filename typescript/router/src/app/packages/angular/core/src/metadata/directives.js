"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
var __assign = (this && this.__assign) || Object.assign || function(t) {
    for (var s, i = 1, n = arguments.length; i < n; i++) {
        s = arguments[i];
        for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p))
            t[p] = s[p];
    }
    return t;
};
exports.__esModule = true;
var constants_1 = require("../change_detection/constants");
var decorators_1 = require("../util/decorators");
/**
 * Directive decorator and metadata.
 *
 *
 * @Annotation
 */
exports.Directive = decorators_1.makeDecorator('Directive', function (dir) {
    if (dir === void 0) { dir = {}; }
    return dir;
});
/**
 * Component decorator and metadata.
 *
 *
 * @Annotation
 */
exports.Component = decorators_1.makeDecorator('Component', function (c) {
    if (c === void 0) { c = {}; }
    return (__assign({ changeDetection: constants_1.ChangeDetectionStrategy.Default }, c));
}, exports.Directive);
/**
 * Pipe decorator and metadata.
 *
 * Use the `@Pipe` annotation to declare that a given class is a pipe. A pipe
 * class must also implement {@link PipeTransform} interface.
 *
 * To use the pipe include a reference to the pipe class in
 * {@link NgModule#declarations}.
 *
 *
 * @Annotation
 */
exports.Pipe = decorators_1.makeDecorator('Pipe', function (p) { return (__assign({ pure: true }, p)); });
/**
 * Input decorator and metadata.
 *
 *
 * @Annotation
 */
exports.Input = decorators_1.makePropDecorator('Input', function (bindingPropertyName) { return ({ bindingPropertyName: bindingPropertyName }); });
/**
 * Output decorator and metadata.
 *
 *
 * @Annotation
 */
exports.Output = decorators_1.makePropDecorator('Output', function (bindingPropertyName) { return ({ bindingPropertyName: bindingPropertyName }); });
/**
 * HostBinding decorator and metadata.
 *
 *
 * @Annotation
 */
exports.HostBinding = decorators_1.makePropDecorator('HostBinding', function (hostPropertyName) { return ({ hostPropertyName: hostPropertyName }); });
/**
 * HostListener decorator and metadata.
 *
 *
 * @Annotation
 */
exports.HostListener = decorators_1.makePropDecorator('HostListener', function (eventName, args) { return ({ eventName: eventName, args: args }); });
