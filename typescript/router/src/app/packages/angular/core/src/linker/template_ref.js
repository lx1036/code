"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
exports.__esModule = true;
/**
 * Represents an Embedded Template that can be used to instantiate Embedded Views.
 *
 * You can access a `TemplateRef`, in two ways. Via a directive placed on a `<ng-template>` element
 * (or directive prefixed with `*`) and have the `TemplateRef` for this Embedded View injected into
 * the constructor of the directive using the `TemplateRef` Token. Alternatively you can query for
 * the `TemplateRef` from a Component or a Directive via {@link Query}.
 *
 * To instantiate Embedded Views based on a Template, use {@link ViewContainerRef#
 * createEmbeddedView}, which will create the View and attach it to the View Container.
 *
 */
var TemplateRef = /** @class */ (function () {
    function TemplateRef() {
    }
    return TemplateRef;
}());
exports.TemplateRef = TemplateRef;
