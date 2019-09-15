"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var assert_1 = require("./assert");
var hooks_1 = require("./hooks");
var instructions_1 = require("./instructions");
var renderer_1 = require("./interfaces/renderer");
var util_1 = require("./util");
var view_ref_1 = require("./view_ref");
/**
 * Bootstraps a component, then creates and returns a `ComponentRef` for that component.
 *
 * @param componentType Component to bootstrap
 * @param options Optional parameters which control bootstrapping
 */
function createComponentRef(componentType, opts) {
    var component = renderComponent(componentType, opts);
    var hostView = instructions_1._getComponentHostLElementNode(component).data;
    var hostViewRef = view_ref_1.createViewRef(hostView, component);
    return {
        location: { nativeElement: getHostElement(component) },
        injector: opts.injector || exports.NULL_INJECTOR,
        instance: component,
        hostView: hostViewRef,
        changeDetectorRef: hostViewRef,
        componentType: componentType,
        // TODO: implement destroy and onDestroy
        destroy: function () { },
        onDestroy: function (cb) { }
    };
}
exports.createComponentRef = createComponentRef;
// TODO: A hack to not pull in the NullInjector from @angular/core.
exports.NULL_INJECTOR = {
    get: function (token, notFoundValue) {
        throw new Error('NullInjector: Not found: ' + util_1.stringify(token));
    }
};
/**
 * Bootstraps a Component into an existing host element and returns an instance
 * of the component.
 *
 * Use this function to bootstrap a component into the DOM tree. Each invocation
 * of this function will create a separate tree of components, injectors and
 * change detection cycles and lifetimes. To dynamically insert a new component
 * into an existing tree such that it shares the same injection, change detection
 * and object lifetime, use {@link ViewContainer#createComponent}.
 *
 * @param componentType Component to bootstrap
 * @param options Optional parameters which control bootstrapping
 */
function renderComponent(componentType /* Type as workaround for: Microsoft/TypeScript/issues/4881 */, opts) {
    if (opts === void 0) { opts = {}; }
    ngDevMode && assert_1.assertComponentType(componentType);
    var rendererFactory = opts.rendererFactory || renderer_1.domRendererFactory3;
    var componentDef = componentType.ngComponentDef;
    if (componentDef.type != componentType)
        componentDef.type = componentType;
    var component;
    // The first index of the first selector is the tag name.
    var componentTag = componentDef.selectors[0][0];
    var hostNode = instructions_1.locateHostElement(rendererFactory, opts.host || componentTag);
    var rootContext = {
        // Incomplete initialization due to circular reference.
        component: (null),
        scheduler: opts.scheduler || requestAnimationFrame.bind(window),
        clean: instructions_1.CLEAN_PROMISE,
    };
    var rootView = instructions_1.createLView(-1, rendererFactory.createRenderer(hostNode, componentDef.rendererType), instructions_1.createTView(null, null), null, rootContext, componentDef.onPush ? 4 /* Dirty */ : 2 /* CheckAlways */);
    rootView.injector = opts.injector || null;
    var oldView = instructions_1.enterView(rootView, (null));
    var elementNode;
    try {
        if (rendererFactory.begin)
            rendererFactory.begin();
        // Create element node at index 0 in data array
        elementNode = instructions_1.hostElement(componentTag, hostNode, componentDef);
        // Create directive instance with factory() and store at index 0 in directives array
        component = rootContext.component = instructions_1.baseDirectiveCreate(0, componentDef.factory(), componentDef);
        instructions_1.initChangeDetectorIfExisting(elementNode.nodeInjector, component, (elementNode.data));
        opts.hostFeatures && opts.hostFeatures.forEach(function (feature) { return feature(component, componentDef); });
        instructions_1.executeInitAndContentHooks();
        instructions_1.setHostBindings(instructions_1.ROOT_DIRECTIVE_INDICES);
        instructions_1.detectChangesInternal(elementNode.data, elementNode, componentDef, component);
    }
    finally {
        instructions_1.leaveView(oldView);
        if (rendererFactory.end)
            rendererFactory.end();
    }
    return component;
}
exports.renderComponent = renderComponent;
/**
 * Used to enable lifecycle hooks on the root component.
 *
 * Include this feature when calling `renderComponent` if the root component
 * you are rendering has lifecycle hooks defined. Otherwise, the hooks won't
 * be called properly.
 *
 * Example:
 *
 * ```
 * renderComponent(AppComponent, {features: [RootLifecycleHooks]});
 * ```
 */
function LifecycleHooksFeature(component, def) {
    var elementNode = instructions_1._getComponentHostLElementNode(component);
    // Root component is always created at dir index 0
    // Root component is always created at dir index 0
    hooks_1.queueInitHooks(0, def.onInit, def.doCheck, elementNode.view.tView);
    hooks_1.queueLifecycleHooks(elementNode.tNode.flags, elementNode.view);
}
exports.LifecycleHooksFeature = LifecycleHooksFeature;
/**
 * Retrieve the root context for any component by walking the parent `LView` until
 * reaching the root `LView`.
 *
 * @param component any component
 */
function getRootContext(component) {
    var rootContext = instructions_1.getRootView(component).context;
    ngDevMode && assert_1.assertNotNull(rootContext, 'rootContext');
    return rootContext;
}
/**
 * Retrieve the host element of the component.
 *
 * Use this function to retrieve the host element of the component. The host
 * element is the element which the component is associated with.
 *
 * @param component Component for which the host element should be retrieved.
 */
function getHostElement(component) {
    return instructions_1._getComponentHostLElementNode(component).native;
}
exports.getHostElement = getHostElement;
/**
 * Retrieves the rendered text for a given component.
 *
 * This function retrieves the host element of a component and
 * and then returns the `textContent` for that element. This implies
 * that the text returned will include re-projected content of
 * the component as well.
 *
 * @param component The component to return the content text for.
 */
function getRenderedText(component) {
    var hostElement = getHostElement(component);
    return hostElement.textContent || '';
}
exports.getRenderedText = getRenderedText;
/**
 * Wait on component until it is rendered.
 *
 * This function returns a `Promise` which is resolved when the component's
 * change detection is executed. This is determined by finding the scheduler
 * associated with the `component`'s render tree and waiting until the scheduler
 * flushes. If nothing is scheduled, the function returns a resolved promise.
 *
 * Example:
 * ```
 * await whenRendered(myComponent);
 * ```
 *
 * @param component Component to wait upon
 * @returns Promise which resolves when the component is rendered.
 */
function whenRendered(component) {
    return getRootContext(component).clean;
}
exports.whenRendered = whenRendered;
//# sourceMappingURL=component.js.map