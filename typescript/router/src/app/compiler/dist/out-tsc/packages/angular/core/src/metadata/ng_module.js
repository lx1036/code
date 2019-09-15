"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var defs_1 = require("../di/defs");
var injectable_1 = require("../di/injectable");
var decorators_1 = require("../util/decorators");
/**
 * Defines a schema that will allow:
 * - any non-Angular elements with a `-` in their name,
 * - any properties on elements with a `-` in their name which is the common rule for custom
 * elements.
 *
 *
 */
exports.CUSTOM_ELEMENTS_SCHEMA = {
    name: 'custom-elements'
};
/**
 * Defines a schema that will allow any property on any element.
 *
 * @experimental
 */
exports.NO_ERRORS_SCHEMA = {
    name: 'no-errors-schema'
};
/**
 * NgModule decorator and metadata.
 *
 *
 * @Annotation
 */
exports.NgModule = decorators_1.makeDecorator('NgModule', function (ngModule) { return ngModule; }, undefined, undefined, function (moduleType, metadata) {
    var imports = (metadata && metadata.imports) || [];
    if (metadata && metadata.exports) {
        imports = imports.concat([metadata.exports]);
    }
    moduleType.ngInjectorDef = defs_1.defineInjector({
        factory: injectable_1.convertInjectableProviderToFactory(moduleType, { useClass: moduleType }),
        providers: metadata && metadata.providers,
        imports: imports,
    });
});
//# sourceMappingURL=ng_module.js.map