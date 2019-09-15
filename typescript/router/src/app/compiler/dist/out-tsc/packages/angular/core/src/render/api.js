"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var di_1 = require("../di");
/**
 * @deprecated Use `RendererType2` (and `Renderer2`) instead.
 */
var /**
 * @deprecated Use `RendererType2` (and `Renderer2`) instead.
 */
RenderComponentType = /** @class */ (function () {
    function RenderComponentType(id, templateUrl, slotCount, encapsulation, styles, animations) {
        this.id = id;
        this.templateUrl = templateUrl;
        this.slotCount = slotCount;
        this.encapsulation = encapsulation;
        this.styles = styles;
        this.animations = animations;
    }
    return RenderComponentType;
}());
exports.RenderComponentType = RenderComponentType;
/**
 * @deprecated Debug info is handeled internally in the view engine now.
 */
var /**
 * @deprecated Debug info is handeled internally in the view engine now.
 */
RenderDebugInfo = /** @class */ (function () {
    function RenderDebugInfo() {
    }
    return RenderDebugInfo;
}());
exports.RenderDebugInfo = RenderDebugInfo;
/**
 * @deprecated Use the `Renderer2` instead.
 */
var /**
 * @deprecated Use the `Renderer2` instead.
 */
Renderer = /** @class */ (function () {
    function Renderer() {
    }
    return Renderer;
}());
exports.Renderer = Renderer;
exports.Renderer2Interceptor = new di_1.InjectionToken('Renderer2Interceptor');
/**
 * Injectable service that provides a low-level interface for modifying the UI.
 *
 * Use this service to bypass Angular's templating and make custom UI changes that can't be
 * expressed declaratively. For example if you need to set a property or an attribute whose name is
 * not statically known, use {@link Renderer#setElementProperty setElementProperty} or
 * {@link Renderer#setElementAttribute setElementAttribute} respectively.
 *
 * If you are implementing a custom renderer, you must implement this interface.
 *
 * The default Renderer implementation is `DomRenderer`. Also available is `WebWorkerRenderer`.
 *
 * @deprecated Use `RendererFactory2` instead.
 */
var /**
 * Injectable service that provides a low-level interface for modifying the UI.
 *
 * Use this service to bypass Angular's templating and make custom UI changes that can't be
 * expressed declaratively. For example if you need to set a property or an attribute whose name is
 * not statically known, use {@link Renderer#setElementProperty setElementProperty} or
 * {@link Renderer#setElementAttribute setElementAttribute} respectively.
 *
 * If you are implementing a custom renderer, you must implement this interface.
 *
 * The default Renderer implementation is `DomRenderer`. Also available is `WebWorkerRenderer`.
 *
 * @deprecated Use `RendererFactory2` instead.
 */
RootRenderer = /** @class */ (function () {
    function RootRenderer() {
    }
    return RootRenderer;
}());
exports.RootRenderer = RootRenderer;
/**
 * @experimental
 */
var /**
 * @experimental
 */
RendererFactory2 = /** @class */ (function () {
    function RendererFactory2() {
    }
    return RendererFactory2;
}());
exports.RendererFactory2 = RendererFactory2;
/**
 * @experimental
 */
/**
 * @experimental
 */
var RendererStyleFlags2;
/**
 * @experimental
 */
(function (RendererStyleFlags2) {
    RendererStyleFlags2[RendererStyleFlags2["Important"] = 1] = "Important";
    RendererStyleFlags2[RendererStyleFlags2["DashCase"] = 2] = "DashCase";
})(RendererStyleFlags2 = exports.RendererStyleFlags2 || (exports.RendererStyleFlags2 = {}));
/**
 * @experimental
 */
var /**
 * @experimental
 */
Renderer2 = /** @class */ (function () {
    function Renderer2() {
    }
    return Renderer2;
}());
exports.Renderer2 = Renderer2;
//# sourceMappingURL=api.js.map