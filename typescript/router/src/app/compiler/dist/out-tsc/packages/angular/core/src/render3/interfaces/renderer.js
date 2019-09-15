"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
// TODO: cleanup once the code is merged in angular/angular
// TODO: cleanup once the code is merged in angular/angular
var RendererStyleFlags3;
// TODO: cleanup once the code is merged in angular/angular
(function (RendererStyleFlags3) {
    RendererStyleFlags3[RendererStyleFlags3["Important"] = 1] = "Important";
    RendererStyleFlags3[RendererStyleFlags3["DashCase"] = 2] = "DashCase";
})(RendererStyleFlags3 = exports.RendererStyleFlags3 || (exports.RendererStyleFlags3 = {}));
/** Returns whether the `renderer` is a `ProceduralRenderer3` */
function isProceduralRenderer(renderer) {
    return !!(renderer.listen);
}
exports.isProceduralRenderer = isProceduralRenderer;
exports.domRendererFactory3 = {
    createRenderer: function (hostElement, rendererType) { return document; }
};
// Note: This hack is necessary so we don't erroneously get a circular dependency
// failure based on types.
exports.unusedValueExportToPlacateAjd = 1;
//# sourceMappingURL=renderer.js.map