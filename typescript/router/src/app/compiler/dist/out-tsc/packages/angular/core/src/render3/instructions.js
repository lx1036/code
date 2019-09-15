"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
require("./ng_dev_mode");
var assert_1 = require("./assert");
var projection_1 = require("./interfaces/projection");
var node_assert_1 = require("./node_assert");
var node_manipulation_1 = require("./node_manipulation");
var node_selector_matcher_1 = require("./node_selector_matcher");
var renderer_1 = require("./interfaces/renderer");
var util_1 = require("./util");
var hooks_1 = require("./hooks");
var errors_1 = require("./errors");
/**
 * Directive (D) sets a property on all component instances using this constant as a key and the
 * component's host node (LElement) as the value. This is used in methods like detectChanges to
 * facilitate jumping from an instance to the host node.
 */
exports.NG_HOST_SYMBOL = '__ngHostLNode__';
/**
 * A permanent marker promise which signifies that the current CD tree is
 * clean.
 */
var _CLEAN_PROMISE = Promise.resolve(null);
/**
 * Directive and element indices for top-level directive.
 *
 * Saved here to avoid re-instantiating an array on every change detection run.
 */
exports._ROOT_DIRECTIVE_INDICES = [0, 0];
/**
 * Token set in currentMatches while dependencies are being resolved.
 *
 * If we visit a directive that has a value set to CIRCULAR, we know we've
 * already seen it, and thus have a circular dependency.
 */
exports.CIRCULAR = '__CIRCULAR__';
/**
 * This property gets set before entering a template.
 *
 * This renderer can be one of two varieties of Renderer3:
 *
 * - ObjectedOrientedRenderer3
 *
 * This is the native browser API style, e.g. operations are methods on individual objects
 * like HTMLElement. With this style, no additional code is needed as a facade (reducing payload
 * size).
 *
 * - ProceduralRenderer3
 *
 * In non-native browser environments (e.g. platforms such as web-workers), this is the facade
 * that enables element manipulation. This also facilitates backwards compatibility with
 * Renderer2.
 */
var renderer;
var rendererFactory;
function getRenderer() {
    // top level variables should not be exported for performance reason (PERF_NOTES.md)
    return renderer;
}
exports.getRenderer = getRenderer;
/** Used to set the parent property when nodes are created. */
var previousOrParentNode;
function getPreviousOrParentNode() {
    // top level variables should not be exported for performance reason (PERF_NOTES.md)
    return previousOrParentNode;
}
exports.getPreviousOrParentNode = getPreviousOrParentNode;
/**
 * If `isParent` is:
 *  - `true`: then `previousOrParentNode` points to a parent node.
 *  - `false`: then `previousOrParentNode` points to previous node (sibling).
 */
var isParent;
/**
 * Static data that corresponds to the instance-specific data array on an LView.
 *
 * Each node's static data is stored in tData at the same index that it's stored
 * in the data array. Any nodes that do not have static data store a null value in
 * tData to avoid a sparse array.
 */
var tData;
/**
 * State of the current view being processed.
 *
 * NOTE: we cheat here and initialize it to `null` even thought the type does not
 * contain `null`. This is because we expect this value to be not `null` as soon
 * as we enter the view. Declaring the type as `null` would require us to place `!`
 * in most instructions since they all assume that `currentView` is defined.
 */
var currentView = (null);
var currentQueries;
function getCurrentQueries(QueryType) {
    // top level variables should not be exported for performance reason (PERF_NOTES.md)
    return currentQueries || (currentQueries = new QueryType());
}
exports.getCurrentQueries = getCurrentQueries;
/**
 * This property gets set before entering a template.
 */
var creationMode;
function getCreationMode() {
    // top level variables should not be exported for performance reason (PERF_NOTES.md)
    return creationMode;
}
exports.getCreationMode = getCreationMode;
/**
 * An array of nodes (text, element, container, etc), pipes, their bindings, and
 * any local variables that need to be stored between invocations.
 */
var data;
/**
 * An array of directive instances in the current view.
 *
 * These must be stored separately from LNodes because their presence is
 * unknown at compile-time and thus space cannot be reserved in data[].
 */
var directives;
/**
 * When a view is destroyed, listeners need to be released and outputs need to be
 * unsubscribed. This cleanup array stores both listener data (in chunks of 4)
 * and output data (in chunks of 2) for a particular view. Combining the arrays
 * saves on memory (70 bytes per array) and on a few bytes of code size (for two
 * separate for loops).
 *
 * If it's a listener being stored:
 * 1st index is: event name to remove
 * 2nd index is: native element
 * 3rd index is: listener function
 * 4th index is: useCapture boolean
 *
 * If it's an output subscription:
 * 1st index is: unsubscribe function
 * 2nd index is: context for function
 */
var cleanup;
/**
 * In this mode, any changes in bindings will throw an ExpressionChangedAfterChecked error.
 *
 * Necessary to support ChangeDetectorRef.checkNoChanges().
 */
var checkNoChangesMode = false;
/** Whether or not this is the first time the current view has been processed. */
var firstTemplatePass = true;
/**
 * Swap the current state with a new state.
 *
 * For performance reasons we store the state in the top level of the module.
 * This way we minimize the number of properties to read. Whenever a new view
 * is entered we have to store the state for later, and when the view is
 * exited the state has to be restored
 *
 * @param newView New state to become active
 * @param host Element to which the View is a child of
 * @returns the previous state;
 */
function enterView(newView, host) {
    var oldView = currentView;
    data = newView && newView.data;
    directives = newView && newView.directives;
    tData = newView && newView.tView.data;
    creationMode = newView && (newView.flags & 1 /* CreationMode */) === 1 /* CreationMode */;
    firstTemplatePass = newView && newView.tView.firstTemplatePass;
    cleanup = newView && newView.cleanup;
    renderer = newView && newView.renderer;
    if (newView && newView.bindingIndex < 0) {
        newView.bindingIndex = newView.bindingStartIndex;
    }
    if (host != null) {
        previousOrParentNode = host;
        isParent = true;
    }
    currentView = newView;
    currentQueries = newView && newView.queries;
    return oldView;
}
exports.enterView = enterView;
/**
 * Used in lieu of enterView to make it clear when we are exiting a child view. This makes
 * the direction of traversal (up or down the view tree) a bit clearer.
 */
function leaveView(newView) {
    if (!checkNoChangesMode) {
        hooks_1.executeHooks((directives), currentView.tView.viewHooks, currentView.tView.viewCheckHooks, creationMode);
    }
    // Views should be clean and in update mode after being checked, so these bits are cleared
    currentView.flags &= ~(1 /* CreationMode */ | 4 /* Dirty */);
    currentView.lifecycleStage = 1 /* Init */;
    currentView.bindingIndex = -1;
    enterView(newView, null);
}
exports.leaveView = leaveView;
/**  Refreshes directives in this view and triggers any init/content hooks.  */
function refreshDirectives() {
    executeInitAndContentHooks();
    var tView = currentView.tView;
    // This needs to be set before children are processed to support recursive components
    tView.firstTemplatePass = firstTemplatePass = false;
    setHostBindings(tView.hostBindings);
    refreshChildComponents(tView.components);
}
/** Sets the host bindings for the current view. */
function setHostBindings(bindings) {
    if (bindings != null) {
        var defs = (currentView.tView.directives);
        for (var i = 0; i < bindings.length; i += 2) {
            var dirIndex = bindings[i];
            var def = defs[dirIndex];
            def.hostBindings && def.hostBindings(dirIndex, bindings[i + 1]);
        }
    }
}
exports.setHostBindings = setHostBindings;
/** Refreshes child components in the current view. */
function refreshChildComponents(components) {
    if (components != null) {
        for (var i = 0; i < components.length; i += 2) {
            componentRefresh(components[i], components[i + 1]);
        }
    }
}
function executeInitAndContentHooks() {
    if (!checkNoChangesMode) {
        var tView = currentView.tView;
        hooks_1.executeInitHooks(currentView, tView, creationMode);
        hooks_1.executeHooks((directives), tView.contentHooks, tView.contentCheckHooks, creationMode);
    }
}
exports.executeInitAndContentHooks = executeInitAndContentHooks;
function createLView(viewId, renderer, tView, template, context, flags) {
    var newView = {
        parent: currentView,
        id: viewId,
        // -1 for component views
        flags: flags | 1 /* CreationMode */ | 8 /* Attached */,
        node: (null),
        // until we initialize it in createNode.
        data: [],
        directives: null,
        tView: tView,
        cleanup: null,
        renderer: renderer,
        child: null,
        tail: null,
        next: null,
        bindingStartIndex: -1,
        bindingIndex: -1,
        template: template,
        context: context,
        dynamicViewCount: 0,
        lifecycleStage: 1 /* Init */,
        queries: null,
        injector: currentView && currentView.injector,
    };
    return newView;
}
exports.createLView = createLView;
/**
 * Creation of LNode object is extracted to a separate function so we always create LNode object
 * with the same shape
 * (same properties assigned in the same order).
 */
function createLNodeObject(type, currentView, parent, native, state, queries) {
    return {
        type: type,
        native: native,
        view: currentView,
        parent: parent,
        child: null,
        next: null,
        nodeInjector: parent ? parent.nodeInjector : null,
        data: state,
        queries: queries,
        tNode: null,
        pNextOrParent: null,
        dynamicLContainerNode: null
    };
}
exports.createLNodeObject = createLNodeObject;
function createLNode(index, type, native, state) {
    var parent = isParent ? previousOrParentNode :
        previousOrParentNode && previousOrParentNode.parent;
    var queries = (isParent ? currentQueries : previousOrParentNode && previousOrParentNode.queries) ||
        parent && parent.queries && parent.queries.child();
    var isState = state != null;
    var node = createLNodeObject(type, currentView, parent, native, isState ? state : null, queries);
    if ((type & 2 /* ViewOrElement */) === 2 /* ViewOrElement */ && isState) {
        // Bit of a hack to bust through the readonly because there is a circular dep between
        // LView and LNode.
        ngDevMode && assert_1.assertNull(state.node, 'LView.node should not have been initialized');
        state.node = node;
    }
    if (index != null) {
        // We are Element or Container
        ngDevMode && assertDataNext(index);
        data[index] = node;
        // Every node adds a value to the static data array to avoid a sparse array
        if (index >= tData.length) {
            tData[index] = null;
        }
        else {
            node.tNode = tData[index];
        }
        // Now link ourselves into the tree.
        if (isParent) {
            currentQueries = null;
            if (previousOrParentNode.view === currentView ||
                previousOrParentNode.type === 2 /* View */) {
                // We are in the same view, which means we are adding content node to the parent View.
                ngDevMode && assert_1.assertNull(previousOrParentNode.child, "previousOrParentNode's child should not have been set.");
                previousOrParentNode.child = node;
            }
            else {
                // We are adding component view, so we don't link parent node child to this node.
            }
        }
        else if (previousOrParentNode) {
            ngDevMode && assert_1.assertNull(previousOrParentNode.next, "previousOrParentNode's next property should not have been set " + index + ".");
            previousOrParentNode.next = node;
            if (previousOrParentNode.dynamicLContainerNode) {
                previousOrParentNode.dynamicLContainerNode.next = node;
            }
        }
    }
    previousOrParentNode = node;
    isParent = true;
    return node;
}
exports.createLNode = createLNode;
/**
 * Resets the application state.
 */
function resetApplicationState() {
    isParent = false;
    previousOrParentNode = (null);
}
/**
 *
 * @param hostNode Existing node to render into.
 * @param template Template function with the instructions.
 * @param context to pass into the template.
 * @param providedRendererFactory renderer factory to use
 * @param host The host element node to use
 * @param directives Directive defs that should be used for matching
 * @param pipes Pipe defs that should be used for matching
 */
function renderTemplate(hostNode, template, context, providedRendererFactory, host, directives, pipes) {
    if (host == null) {
        resetApplicationState();
        rendererFactory = providedRendererFactory;
        var tView = getOrCreateTView(template, directives || null, pipes || null);
        host = createLNode(null, 3 /* Element */, hostNode, createLView(-1, providedRendererFactory.createRenderer(null, null), tView, null, {}, 2 /* CheckAlways */));
    }
    var hostView = (host.data);
    ngDevMode && assert_1.assertNotNull(hostView, 'Host node should have an LView defined in host.data.');
    renderComponentOrTemplate(host, hostView, context, template);
    return host;
}
exports.renderTemplate = renderTemplate;
function renderEmbeddedTemplate(viewNode, template, context, renderer, directives, pipes) {
    var _isParent = isParent;
    var _previousOrParentNode = previousOrParentNode;
    var oldView;
    try {
        isParent = true;
        previousOrParentNode = (null);
        var rf = 2 /* Update */;
        if (viewNode == null) {
            var tView = getOrCreateTView(template, directives || null, pipes || null);
            var lView = createLView(-1, renderer, tView, template, context, 2 /* CheckAlways */);
            viewNode = createLNode(null, 2 /* View */, null, lView);
            rf = 1 /* Create */;
        }
        oldView = enterView(viewNode.data, viewNode);
        template(rf, context);
        refreshDirectives();
        refreshDynamicChildren();
    }
    finally {
        leaveView((oldView));
        isParent = _isParent;
        previousOrParentNode = _previousOrParentNode;
    }
    return viewNode;
}
exports.renderEmbeddedTemplate = renderEmbeddedTemplate;
function renderComponentOrTemplate(node, hostView, componentOrContext, template) {
    var oldView = enterView(hostView, node);
    try {
        if (rendererFactory.begin) {
            rendererFactory.begin();
        }
        if (template) {
            template(getRenderFlags(hostView), (componentOrContext));
            refreshDynamicChildren();
            refreshDirectives();
        }
        else {
            executeInitAndContentHooks();
            // Element was stored at 0 in data and directive was stored at 0 in directives
            // in renderComponent()
            setHostBindings(exports._ROOT_DIRECTIVE_INDICES);
            componentRefresh(0, 0);
        }
    }
    finally {
        if (rendererFactory.end) {
            rendererFactory.end();
        }
        leaveView(oldView);
    }
}
exports.renderComponentOrTemplate = renderComponentOrTemplate;
/**
 * This function returns the default configuration of rendering flags depending on when the
 * template is in creation mode or update mode. By default, the update block is run with the
 * creation block when the view is in creation mode. Otherwise, the update block is run
 * alone.
 *
 * Dynamically created views do NOT use this configuration (update block and create block are
 * always run separately).
 */
function getRenderFlags(view) {
    return view.flags & 1 /* CreationMode */ ? 1 /* Create */ | 2 /* Update */ :
        2 /* Update */;
}
/**
 * Create DOM element. The instruction must later be followed by `elementEnd()` call.
 *
 * @param index Index of the element in the data array
 * @param name Name of the DOM Node
 * @param attrs Statically bound set of attributes to be written into the DOM element on creation.
 * @param localRefs A set of local reference bindings on the element.
 *
 * Attributes and localRefs are passed as an array of strings where elements with an even index
 * hold an attribute name and elements with an odd index hold an attribute value, ex.:
 * ['id', 'warning5', 'class', 'alert']
 */
function elementStart(index, name, attrs, localRefs) {
    ngDevMode &&
        assert_1.assertEqual(currentView.bindingStartIndex, -1, 'elements should be created before any bindings');
    var native = renderer.createElement(name);
    var node = createLNode(index, 3 /* Element */, (native), null);
    if (attrs)
        setUpAttributes(native, attrs);
    node_manipulation_1.appendChild((node.parent), native, currentView);
    createDirectivesAndLocals(index, name, attrs, localRefs, null);
    return native;
}
exports.elementStart = elementStart;
function createDirectivesAndLocals(index, name, attrs, localRefs, containerData) {
    var node = previousOrParentNode;
    if (firstTemplatePass) {
        ngDevMode && assertDataInRange(index - 1);
        node.tNode = tData[index] = createTNode(name, attrs || null, containerData);
        cacheMatchingDirectivesForNode(node.tNode, currentView.tView, localRefs || null);
    }
    else {
        instantiateDirectivesDirectly();
    }
    saveResolvedLocalsInData();
}
/**
 * On first template pass, we match each node against available directive selectors and save
 * the resulting defs in the correct instantiation order for subsequent change detection runs
 * (so dependencies are always created before the directives that inject them).
 */
function cacheMatchingDirectivesForNode(tNode, tView, localRefs) {
    // Please make sure to have explicit type for `exportsMap`. Inferred type triggers bug in tsickle.
    var exportsMap = localRefs ? { '': -1 } : null;
    var matches = tView.currentMatches = findDirectiveMatches(tNode);
    if (matches) {
        for (var i = 0; i < matches.length; i += 2) {
            var def = matches[i];
            var valueIndex = i + 1;
            resolveDirective(def, valueIndex, matches, tView);
            saveNameToExportMap(matches[valueIndex], def, exportsMap);
        }
    }
    if (exportsMap)
        cacheMatchingLocalNames(tNode, localRefs, exportsMap);
}
/** Matches the current node against all available selectors. */
function findDirectiveMatches(tNode) {
    var registry = currentView.tView.directiveRegistry;
    var matches = null;
    if (registry) {
        for (var i = 0; i < registry.length; i++) {
            var def = registry[i];
            if (node_selector_matcher_1.isNodeMatchingSelectorList(tNode, (def.selectors))) {
                if (def.template) {
                    if (tNode.flags & 4096 /* isComponent */)
                        errors_1.throwMultipleComponentError(tNode);
                    tNode.flags = 4096 /* isComponent */;
                }
                if (def.diPublic)
                    def.diPublic(def);
                (matches || (matches = [])).push(def, null);
            }
        }
    }
    return matches;
}
function resolveDirective(def, valueIndex, matches, tView) {
    if (matches[valueIndex] === null) {
        matches[valueIndex] = exports.CIRCULAR;
        var instance = def.factory();
        (tView.directives || (tView.directives = [])).push(def);
        return directiveCreate(matches[valueIndex] = tView.directives.length - 1, instance, def);
    }
    else if (matches[valueIndex] === exports.CIRCULAR) {
        // If we revisit this directive before it's resolved, we know it's circular
        // If we revisit this directive before it's resolved, we know it's circular
        errors_1.throwCyclicDependencyError(def.type);
    }
    return null;
}
exports.resolveDirective = resolveDirective;
/** Stores index of component's host element so it will be queued for view refresh during CD. */
function queueComponentIndexForCheck(dirIndex) {
    if (firstTemplatePass) {
        (currentView.tView.components || (currentView.tView.components = [])).push(dirIndex, data.length - 1);
    }
}
/** Stores index of directive and host element so it will be queued for binding refresh during CD.
 */
function queueHostBindingForCheck(dirIndex) {
    ngDevMode &&
        assert_1.assertEqual(firstTemplatePass, true, 'Should only be called in first template pass.');
    (currentView.tView.hostBindings || (currentView.tView.hostBindings = [])).push(dirIndex, data.length - 1);
}
/** Sets the context for a ChangeDetectorRef to the given instance. */
function initChangeDetectorIfExisting(injector, instance, view) {
    if (injector && injector.changeDetectorRef != null) {
        injector.changeDetectorRef._setComponentContext(view, instance);
    }
}
exports.initChangeDetectorIfExisting = initChangeDetectorIfExisting;
function isComponent(tNode) {
    return (tNode.flags & 4096 /* isComponent */) === 4096 /* isComponent */;
}
exports.isComponent = isComponent;
/**
 * This function instantiates the given directives.
 */
function instantiateDirectivesDirectly() {
    var tNode = (previousOrParentNode.tNode);
    var count = tNode.flags & 4095 /* DirectiveCountMask */;
    if (count > 0) {
        var start = tNode.flags >> 13 /* DirectiveStartingIndexShift */;
        var end = start + count;
        var tDirectives = (currentView.tView.directives);
        for (var i = start; i < end; i++) {
            var def = tDirectives[i];
            directiveCreate(i, def.factory(), def);
        }
    }
}
/** Caches local names and their matching directive indices for query and template lookups. */
function cacheMatchingLocalNames(tNode, localRefs, exportsMap) {
    if (localRefs) {
        var localNames = tNode.localNames = [];
        // Local names must be stored in tNode in the same order that localRefs are defined
        // in the template to ensure the data is loaded in the same slots as their refs
        // in the template (for template queries).
        for (var i = 0; i < localRefs.length; i += 2) {
            var index = exportsMap[localRefs[i + 1]];
            if (index == null)
                throw new Error("Export of name '" + localRefs[i + 1] + "' not found!");
            localNames.push(localRefs[i], index);
        }
    }
}
/**
 * Builds up an export map as directives are created, so local refs can be quickly mapped
 * to their directive instances.
 */
function saveNameToExportMap(index, def, exportsMap) {
    if (exportsMap) {
        if (def.exportAs)
            exportsMap[def.exportAs] = index;
        if (def.template)
            exportsMap[''] = index;
    }
}
/**
 * Takes a list of local names and indices and pushes the resolved local variable values
 * to data[] in the same order as they are loaded in the template with load().
 */
function saveResolvedLocalsInData() {
    var localNames = previousOrParentNode.tNode.localNames;
    if (localNames) {
        for (var i = 0; i < localNames.length; i += 2) {
            var index = localNames[i + 1];
            var value = index === -1 ? previousOrParentNode.native : directives[index];
            data.push(value);
        }
    }
}
/**
 * Gets TView from a template function or creates a new TView
 * if it doesn't already exist.
 *
 * @param template The template from which to get static data
 * @param directives Directive defs that should be saved on TView
 * @param pipes Pipe defs that should be saved on TView
 * @returns TView
 */
function getOrCreateTView(template, directives, pipes) {
    return template.ngPrivateData ||
        (template.ngPrivateData = createTView(directives, pipes));
}
/** Creates a TView instance */
function createTView(defs, pipes) {
    return {
        data: [],
        directives: null,
        firstTemplatePass: true,
        initHooks: null,
        checkHooks: null,
        contentHooks: null,
        contentCheckHooks: null,
        viewHooks: null,
        viewCheckHooks: null,
        destroyHooks: null,
        pipeDestroyHooks: null,
        hostBindings: null,
        components: null,
        directiveRegistry: typeof defs === 'function' ? defs() : defs,
        pipeRegistry: typeof pipes === 'function' ? pipes() : pipes,
        currentMatches: null
    };
}
exports.createTView = createTView;
function setUpAttributes(native, attrs) {
    ngDevMode && assert_1.assertEqual(attrs.length % 2, 0, 'each attribute should have a key and a value');
    var isProc = renderer_1.isProceduralRenderer(renderer);
    for (var i = 0; i < attrs.length; i += 2) {
        var attrName = attrs[i];
        if (attrName !== projection_1.NG_PROJECT_AS_ATTR_NAME) {
            var attrVal = attrs[i + 1];
            isProc ? renderer.setAttribute(native, attrName, attrVal) :
                native.setAttribute(attrName, attrVal);
        }
    }
}
function createError(text, token) {
    return new Error("Renderer: " + text + " [" + util_1.stringify(token) + "]");
}
exports.createError = createError;
/**
 * Locates the host native element, used for bootstrapping existing nodes into rendering pipeline.
 *
 * @param elementOrSelector Render element or CSS selector to locate the element.
 */
function locateHostElement(factory, elementOrSelector) {
    ngDevMode && assertDataInRange(-1);
    rendererFactory = factory;
    var defaultRenderer = factory.createRenderer(null, null);
    var rNode = typeof elementOrSelector === 'string' ?
        (renderer_1.isProceduralRenderer(defaultRenderer) ?
            defaultRenderer.selectRootElement(elementOrSelector) :
            defaultRenderer.querySelector(elementOrSelector)) :
        elementOrSelector;
    if (ngDevMode && !rNode) {
        if (typeof elementOrSelector === 'string') {
            throw createError('Host node with selector not found:', elementOrSelector);
        }
        else {
            throw createError('Host node is required:', elementOrSelector);
        }
    }
    return rNode;
}
exports.locateHostElement = locateHostElement;
/**
 * Creates the host LNode.
 *
 * @param rNode Render host element.
 * @param def ComponentDef
 *
 * @returns LElementNode created
 */
function hostElement(tag, rNode, def) {
    resetApplicationState();
    var node = createLNode(0, 3 /* Element */, rNode, createLView(-1, renderer, getOrCreateTView(def.template, def.directiveDefs, def.pipeDefs), null, null, def.onPush ? 4 /* Dirty */ : 2 /* CheckAlways */));
    if (firstTemplatePass) {
        node.tNode = createTNode(tag, null, null);
        node.tNode.flags = 4096 /* isComponent */;
        if (def.diPublic)
            def.diPublic(def);
        currentView.tView.directives = [def];
    }
    return node;
}
exports.hostElement = hostElement;
/**
 * Adds an event listener to the current node.
 *
 * If an output exists on one of the node's directives, it also subscribes to the output
 * and saves the subscription for later cleanup.
 *
 * @param eventName Name of the event
 * @param listenerFn The function to be called when event emits
 * @param useCapture Whether or not to use capture in event listener.
 */
function listener(eventName, listenerFn, useCapture) {
    if (useCapture === void 0) { useCapture = false; }
    ngDevMode && assertPreviousIsParent();
    var node = previousOrParentNode;
    var native = node.native;
    // In order to match current behavior, native DOM event listeners must be added for all
    // events (including outputs).
    var cleanupFns = cleanup || (cleanup = currentView.cleanup = []);
    if (renderer_1.isProceduralRenderer(renderer)) {
        var wrappedListener = wrapListenerWithDirtyLogic(currentView, listenerFn);
        var cleanupFn = renderer.listen(native, eventName, wrappedListener);
        cleanupFns.push(cleanupFn, null);
    }
    else {
        var wrappedListener = wrapListenerWithDirtyAndDefault(currentView, listenerFn);
        native.addEventListener(eventName, wrappedListener, useCapture);
        cleanupFns.push(eventName, native, wrappedListener, useCapture);
    }
    var tNode = (node.tNode);
    if (tNode.outputs === undefined) {
        // if we create TNode here, inputs must be undefined so we know they still need to be
        // checked
        tNode.outputs = generatePropertyAliases(node.tNode.flags, 1 /* Output */);
    }
    var outputs = tNode.outputs;
    var outputData;
    if (outputs && (outputData = outputs[eventName])) {
        createOutput(outputData, listenerFn);
    }
}
exports.listener = listener;
/**
 * Iterates through the outputs associated with a particular event name and subscribes to
 * each output.
 */
function createOutput(outputs, listener) {
    for (var i = 0; i < outputs.length; i += 2) {
        ngDevMode && assertDataInRange(outputs[i], (directives));
        var subscription = directives[outputs[i]][outputs[i + 1]].subscribe(listener);
        cleanup.push(subscription.unsubscribe, subscription);
    }
}
/** Mark the end of the element. */
function elementEnd() {
    if (isParent) {
        isParent = false;
    }
    else {
        ngDevMode && assertHasParent();
        previousOrParentNode = (previousOrParentNode.parent);
    }
    ngDevMode && node_assert_1.assertNodeType(previousOrParentNode, 3 /* Element */);
    var queries = previousOrParentNode.queries;
    queries && queries.addNode(previousOrParentNode);
    hooks_1.queueLifecycleHooks(previousOrParentNode.tNode.flags, currentView);
}
exports.elementEnd = elementEnd;
/**
 * Updates the value of removes an attribute on an Element.
 *
 * @param number index The index of the element in the data array
 * @param name name The name of the attribute.
 * @param value value The attribute is removed when value is `null` or `undefined`.
 *                  Otherwise the attribute value is set to the stringified value.
 * @param sanitizer An optional function used to sanitize the value.
 */
function elementAttribute(index, name, value, sanitizer) {
    if (value !== exports.NO_CHANGE) {
        var element = data[index];
        if (value == null) {
            renderer_1.isProceduralRenderer(renderer) ? renderer.removeAttribute(element.native, name) :
                element.native.removeAttribute(name);
        }
        else {
            var strValue = sanitizer == null ? util_1.stringify(value) : sanitizer(value);
            renderer_1.isProceduralRenderer(renderer) ? renderer.setAttribute(element.native, name, strValue) :
                element.native.setAttribute(name, strValue);
        }
    }
}
exports.elementAttribute = elementAttribute;
/**
 * Update a property on an Element.
 *
 * If the property name also exists as an input property on one of the element's directives,
 * the component property will be set instead of the element property. This check must
 * be conducted at runtime so child components that add new @Inputs don't have to be re-compiled.
 *
 * @param index The index of the element to update in the data array
 * @param propName Name of property. Because it is going to DOM, this is not subject to
 *        renaming as part of minification.
 * @param value New value to write.
 * @param sanitizer An optional function used to sanitize the value.
 */
function elementProperty(index, propName, value, sanitizer) {
    if (value === exports.NO_CHANGE)
        return;
    var node = data[index];
    var tNode = (node.tNode);
    // if tNode.inputs is undefined, a listener has created outputs, but inputs haven't
    // yet been checked
    if (tNode && tNode.inputs === undefined) {
        // mark inputs as checked
        tNode.inputs = generatePropertyAliases(node.tNode.flags, 0 /* Input */);
    }
    var inputData = tNode && tNode.inputs;
    var dataValue;
    if (inputData && (dataValue = inputData[propName])) {
        setInputsForProperty(dataValue, value);
        markDirtyIfOnPush(node);
    }
    else {
        // It is assumed that the sanitizer is only added when the compiler determines that the property
        // is risky, so sanitization can be done without further checks.
        value = sanitizer != null ? sanitizer(value) : value;
        var native = node.native;
        renderer_1.isProceduralRenderer(renderer) ? renderer.setProperty(native, propName, value) :
            (native.setProperty ? native.setProperty(propName, value) :
                native[propName] = value);
    }
}
exports.elementProperty = elementProperty;
/**
 * Constructs a TNode object from the arguments.
 *
 * @param tagName
 * @param attrs
 * @param data
 * @param localNames A list of local names and their matching indices
 * @returns the TNode object
 */
function createTNode(tagName, attrs, data) {
    return {
        flags: 0,
        tagName: tagName,
        attrs: attrs,
        localNames: null,
        initialInputs: undefined,
        inputs: undefined,
        outputs: undefined,
        data: data
    };
}
/**
 * Given a list of directive indices and minified input names, sets the
 * input properties on the corresponding directives.
 */
function setInputsForProperty(inputs, value) {
    for (var i = 0; i < inputs.length; i += 2) {
        ngDevMode && assertDataInRange(inputs[i], (directives));
        directives[inputs[i]][inputs[i + 1]] = value;
    }
}
/**
 * Consolidates all inputs or outputs of all directives on this logical node.
 *
 * @param number lNodeFlags logical node flags
 * @param Direction direction whether to consider inputs or outputs
 * @returns PropertyAliases|null aggregate of all properties if any, `null` otherwise
 */
function generatePropertyAliases(tNodeFlags, direction) {
    var count = tNodeFlags & 4095 /* DirectiveCountMask */;
    var propStore = null;
    if (count > 0) {
        var start = tNodeFlags >> 13 /* DirectiveStartingIndexShift */;
        var end = start + count;
        var isInput = direction === 0 /* Input */;
        var defs = (currentView.tView.directives);
        for (var i = start; i < end; i++) {
            var directiveDef = defs[i];
            var propertyAliasMap = isInput ? directiveDef.inputs : directiveDef.outputs;
            for (var publicName in propertyAliasMap) {
                if (propertyAliasMap.hasOwnProperty(publicName)) {
                    propStore = propStore || {};
                    var internalName = propertyAliasMap[publicName];
                    var hasProperty = propStore.hasOwnProperty(publicName);
                    hasProperty ? propStore[publicName].push(i, internalName) :
                        (propStore[publicName] = [i, internalName]);
                }
            }
        }
    }
    return propStore;
}
/**
 * Add or remove a class in a `classList` on a DOM element.
 *
 * This instruction is meant to handle the [class.foo]="exp" case
 *
 * @param index The index of the element to update in the data array
 * @param className Name of class to toggle. Because it is going to DOM, this is not subject to
 *        renaming as part of minification.
 * @param value A value indicating if a given class should be added or removed.
 */
function elementClassNamed(index, className, value) {
    if (value !== exports.NO_CHANGE) {
        var lElement = data[index];
        if (value) {
            renderer_1.isProceduralRenderer(renderer) ? renderer.addClass(lElement.native, className) :
                lElement.native.classList.add(className);
        }
        else {
            renderer_1.isProceduralRenderer(renderer) ? renderer.removeClass(lElement.native, className) :
                lElement.native.classList.remove(className);
        }
    }
}
exports.elementClassNamed = elementClassNamed;
/**
 * Set the `className` property on a DOM element.
 *
 * This instruction is meant to handle the `[class]="exp"` usage.
 *
 * `elementClass` instruction writes the value to the "element's" `className` property.
 *
 * @param index The index of the element to update in the data array
 * @param value A value indicating a set of classes which should be applied. The method overrides
 *   any existing classes. The value is stringified (`toString`) before it is applied to the
 *   element.
 */
function elementClass(index, value) {
    if (value !== exports.NO_CHANGE) {
        // TODO: This is a naive implementation which simply writes value to the `className`. In the
        // future
        // we will add logic here which would work with the animation code.
        var lElement = data[index];
        renderer_1.isProceduralRenderer(renderer) ? renderer.setProperty(lElement.native, 'className', value) :
            lElement.native['className'] = util_1.stringify(value);
    }
}
exports.elementClass = elementClass;
function elementStyleNamed(index, styleName, value, suffixOrSanitizer) {
    if (value !== exports.NO_CHANGE) {
        var lElement = data[index];
        if (value == null) {
            renderer_1.isProceduralRenderer(renderer) ?
                renderer.removeStyle(lElement.native, styleName, renderer_1.RendererStyleFlags3.DashCase) :
                lElement.native['style'].removeProperty(styleName);
        }
        else {
            var strValue = typeof suffixOrSanitizer == 'function' ? suffixOrSanitizer(value) : util_1.stringify(value);
            if (typeof suffixOrSanitizer == 'string')
                strValue = strValue + suffixOrSanitizer;
            renderer_1.isProceduralRenderer(renderer) ?
                renderer.setStyle(lElement.native, styleName, strValue, renderer_1.RendererStyleFlags3.DashCase) :
                lElement.native['style'].setProperty(styleName, strValue);
        }
    }
}
exports.elementStyleNamed = elementStyleNamed;
/**
 * Set the `style` property on a DOM element.
 *
 * This instruction is meant to handle the `[style]="exp"` usage.
 *
 *
 * @param index The index of the element to update in the data array
 * @param value A value indicating if a given style should be added or removed.
 *   The expected shape of `value` is an object where keys are style names and the values
 *   are their corresponding values to set. If value is falsy than the style is remove. An absence
 *   of style does not cause that style to be removed. `NO_CHANGE` implies that no update should be
 *   performed.
 */
function elementStyle(index, value) {
    if (value !== exports.NO_CHANGE) {
        // TODO: This is a naive implementation which simply writes value to the `style`. In the future
        // we will add logic here which would work with the animation code.
        var lElement = data[index];
        if (renderer_1.isProceduralRenderer(renderer)) {
            renderer.setProperty(lElement.native, 'style', value);
        }
        else {
            var style = lElement.native['style'];
            for (var i = 0, keys = Object.keys(value); i < keys.length; i++) {
                var styleName = keys[i];
                var styleValue = value[styleName];
                styleValue == null ? style.removeProperty(styleName) :
                    style.setProperty(styleName, styleValue);
            }
        }
    }
}
exports.elementStyle = elementStyle;
/**
 * Create static text node
 *
 * @param index Index of the node in the data array.
 * @param value Value to write. This value will be stringified.
 *   If value is not provided than the actual creation of the text node is delayed.
 */
function text(index, value) {
    ngDevMode &&
        assert_1.assertEqual(currentView.bindingStartIndex, -1, 'text nodes should be created before bindings');
    var textNode = value != null ? node_manipulation_1.createTextNode(value, renderer) : null;
    var node = createLNode(index, 3 /* Element */, textNode);
    // Text nodes are self closing.
    isParent = false;
    node_manipulation_1.appendChild((node.parent), textNode, currentView);
}
exports.text = text;
/**
 * Create text node with binding
 * Bindings should be handled externally with the proper bind(1-8) method
 *
 * @param index Index of the node in the data array.
 * @param value Stringified value to write.
 */
function textBinding(index, value) {
    ngDevMode && assertDataInRange(index);
    var existingNode = data[index];
    ngDevMode && assert_1.assertNotNull(existingNode, 'existing node');
    if (existingNode.native) {
        // If DOM node exists and value changed, update textContent
        value !== exports.NO_CHANGE &&
            (renderer_1.isProceduralRenderer(renderer) ? renderer.setValue(existingNode.native, util_1.stringify(value)) :
                existingNode.native.textContent = util_1.stringify(value));
    }
    else {
        // Node was created but DOM node creation was delayed. Create and append now.
        existingNode.native = node_manipulation_1.createTextNode(value, renderer);
        node_manipulation_1.insertChild(existingNode, currentView);
    }
}
exports.textBinding = textBinding;
/**
 * Create a directive.
 *
 * NOTE: directives can be created in order other than the index order. They can also
 *       be retrieved before they are created in which case the value will be null.
 *
 * @param directive The directive instance.
 * @param directiveDef DirectiveDef object which contains information about the template.
 */
function directiveCreate(index, directive, directiveDef) {
    var instance = baseDirectiveCreate(index, directive, directiveDef);
    ngDevMode && assert_1.assertNotNull(previousOrParentNode.tNode, 'previousOrParentNode.tNode');
    var tNode = previousOrParentNode.tNode;
    var isComponent = directiveDef.template;
    if (isComponent) {
        addComponentLogic(index, directive, directiveDef);
    }
    if (firstTemplatePass) {
        // Init hooks are queued now so ngOnInit is called in host components before
        // any projected components.
        // Init hooks are queued now so ngOnInit is called in host components before
        // any projected components.
        hooks_1.queueInitHooks(index, directiveDef.onInit, directiveDef.doCheck, currentView.tView);
        if (directiveDef.hostBindings)
            queueHostBindingForCheck(index);
    }
    if (tNode && tNode.attrs) {
        setInputsFromAttrs(index, instance, directiveDef.inputs, tNode);
    }
    return instance;
}
exports.directiveCreate = directiveCreate;
function addComponentLogic(index, instance, def) {
    var tView = getOrCreateTView(def.template, def.directiveDefs, def.pipeDefs);
    // Only component views should be added to the view tree directly. Embedded views are
    // accessed through their containers because they may be removed / re-added later.
    var hostView = addToViewTree(currentView, createLView(-1, rendererFactory.createRenderer(previousOrParentNode.native, def.rendererType), tView, null, null, def.onPush ? 4 /* Dirty */ : 2 /* CheckAlways */));
    previousOrParentNode.data = hostView;
    hostView.node = previousOrParentNode;
    initChangeDetectorIfExisting(previousOrParentNode.nodeInjector, instance, hostView);
    if (firstTemplatePass)
        queueComponentIndexForCheck(index);
}
/**
 * A lighter version of directiveCreate() that is used for the root component
 *
 * This version does not contain features that we don't already support at root in
 * current Angular. Example: local refs and inputs on root component.
 */
function baseDirectiveCreate(index, directive, directiveDef) {
    ngDevMode &&
        assert_1.assertEqual(currentView.bindingStartIndex, -1, 'directives should be created before any bindings');
    ngDevMode && assertPreviousIsParent();
    Object.defineProperty(directive, exports.NG_HOST_SYMBOL, { enumerable: false, value: previousOrParentNode });
    if (directives == null)
        currentView.directives = directives = [];
    ngDevMode && assertDataNext(index, directives);
    directives[index] = directive;
    if (firstTemplatePass) {
        var flags = previousOrParentNode.tNode.flags;
        if ((flags & 4095 /* DirectiveCountMask */) === 0) {
            // When the first directive is created:
            // - save the index,
            // - set the number of directives to 1
            // When the first directive is created:
            // - save the index,
            // - set the number of directives to 1
            previousOrParentNode.tNode.flags =
                index << 13 /* DirectiveStartingIndexShift */ | flags & 4096 /* isComponent */ | 1;
        }
        else {
            // Only need to bump the size when subsequent directives are created
            ngDevMode && assert_1.assertNotEqual(flags & 4095 /* DirectiveCountMask */, 4095 /* DirectiveCountMask */, 'Reached the max number of directives');
            previousOrParentNode.tNode.flags++;
        }
    }
    else {
        var diPublic = directiveDef.diPublic;
        if (diPublic)
            diPublic((directiveDef));
    }
    if (directiveDef.attributes != null && previousOrParentNode.type == 3 /* Element */) {
        setUpAttributes(previousOrParentNode.native, directiveDef.attributes);
    }
    return directive;
}
exports.baseDirectiveCreate = baseDirectiveCreate;
/**
 * Sets initial input properties on directive instances from attribute data
 *
 * @param directiveIndex Index of the directive in directives array
 * @param instance Instance of the directive on which to set the initial inputs
 * @param inputs The list of inputs from the directive def
 * @param tNode The static data for this node
 */
function setInputsFromAttrs(directiveIndex, instance, inputs, tNode) {
    var initialInputData = tNode.initialInputs;
    if (initialInputData === undefined || directiveIndex >= initialInputData.length) {
        initialInputData = generateInitialInputs(directiveIndex, inputs, tNode);
    }
    var initialInputs = initialInputData[directiveIndex];
    if (initialInputs) {
        for (var i = 0; i < initialInputs.length; i += 2) {
            instance[initialInputs[i]] = initialInputs[i + 1];
        }
    }
}
/**
 * Generates initialInputData for a node and stores it in the template's static storage
 * so subsequent template invocations don't have to recalculate it.
 *
 * initialInputData is an array containing values that need to be set as input properties
 * for directives on this node, but only once on creation. We need this array to support
 * the case where you set an @Input property of a directive using attribute-like syntax.
 * e.g. if you have a `name` @Input, you can set it once like this:
 *
 * <my-component name="Bess"></my-component>
 *
 * @param directiveIndex Index to store the initial input data
 * @param inputs The list of inputs from the directive def
 * @param tNode The static data on this node
 */
function generateInitialInputs(directiveIndex, inputs, tNode) {
    var initialInputData = tNode.initialInputs || (tNode.initialInputs = []);
    initialInputData[directiveIndex] = null;
    var attrs = (tNode.attrs);
    for (var i = 0; i < attrs.length; i += 2) {
        var attrName = attrs[i];
        var minifiedInputName = inputs[attrName];
        if (minifiedInputName !== undefined) {
            var inputsToStore = initialInputData[directiveIndex] || (initialInputData[directiveIndex] = []);
            inputsToStore.push(minifiedInputName, attrs[i + 1]);
        }
    }
    return initialInputData;
}
function createLContainer(parentLNode, currentView, template) {
    ngDevMode && assert_1.assertNotNull(parentLNode, 'containers should have a parent');
    return {
        views: [],
        nextIndex: 0,
        // If the direct parent of the container is a view, its views will need to be added
        // through insertView() when its parent view is being inserted:
        renderParent: node_manipulation_1.canInsertNativeNode(parentLNode, currentView) ? parentLNode : null,
        template: template == null ? null : template,
        next: null,
        parent: currentView,
        dynamicViewCount: 0,
        queries: null
    };
}
exports.createLContainer = createLContainer;
/**
 * Creates an LContainerNode.
 *
 * Only `LViewNodes` can go into `LContainerNodes`.
 *
 * @param index The index of the container in the data array
 * @param template Optional inline template
 * @param tagName The name of the container element, if applicable
 * @param attrs The attrs attached to the container, if applicable
 * @param localRefs A set of local reference bindings on the element.
 */
function container(index, template, tagName, attrs, localRefs) {
    ngDevMode && assert_1.assertEqual(currentView.bindingStartIndex, -1, 'container nodes should be created before any bindings');
    var currentParent = isParent ? previousOrParentNode : previousOrParentNode.parent;
    var lContainer = createLContainer(currentParent, currentView, template);
    var node = createLNode(index, 0 /* Container */, undefined, lContainer);
    // Containers are added to the current view tree instead of their embedded views
    // because views can be removed and re-inserted.
    addToViewTree(currentView, node.data);
    createDirectivesAndLocals(index, tagName || null, attrs, localRefs, []);
    isParent = false;
    ngDevMode && node_assert_1.assertNodeType(previousOrParentNode, 0 /* Container */);
    var queries = node.queries;
    if (queries) {
        // check if a given container node matches
        queries.addNode(node);
        // prepare place for matching nodes from views inserted into a given container
        lContainer.queries = queries.container();
    }
}
exports.container = container;
/**
 * Sets a container up to receive views.
 *
 * @param index The index of the container in the data array
 */
function containerRefreshStart(index) {
    ngDevMode && assertDataInRange(index);
    previousOrParentNode = data[index];
    ngDevMode && node_assert_1.assertNodeType(previousOrParentNode, 0 /* Container */);
    isParent = true;
    previousOrParentNode.data.nextIndex = 0;
    ngDevMode && assert_1.assertSame(previousOrParentNode.native, undefined, "the container's native element should not have been set yet.");
    if (!checkNoChangesMode) {
        // We need to execute init hooks here so ngOnInit hooks are called in top level views
        // before they are called in embedded views (for backwards compatibility).
        // We need to execute init hooks here so ngOnInit hooks are called in top level views
        // before they are called in embedded views (for backwards compatibility).
        hooks_1.executeInitHooks(currentView, currentView.tView, creationMode);
    }
}
exports.containerRefreshStart = containerRefreshStart;
/**
 * Marks the end of the LContainerNode.
 *
 * Marking the end of LContainerNode is the time when to child Views get inserted or removed.
 */
function containerRefreshEnd() {
    if (isParent) {
        isParent = false;
    }
    else {
        ngDevMode && node_assert_1.assertNodeType(previousOrParentNode, 2 /* View */);
        ngDevMode && assertHasParent();
        previousOrParentNode = (previousOrParentNode.parent);
    }
    ngDevMode && node_assert_1.assertNodeType(previousOrParentNode, 0 /* Container */);
    var container = previousOrParentNode;
    container.native = undefined;
    ngDevMode && node_assert_1.assertNodeType(container, 0 /* Container */);
    var nextIndex = container.data.nextIndex;
    // remove extra views at the end of the container
    while (nextIndex < container.data.views.length) {
        node_manipulation_1.removeView(container, nextIndex);
    }
}
exports.containerRefreshEnd = containerRefreshEnd;
function refreshDynamicChildren() {
    for (var current = currentView.child; current !== null; current = current.next) {
        if (current.dynamicViewCount !== 0 && current.views) {
            var container_1 = current;
            for (var i = 0; i < container_1.views.length; i++) {
                var view = container_1.views[i];
                // The directives and pipes are not needed here as an existing view is only being refreshed.
                renderEmbeddedTemplate(view, (view.data.template), (view.data.context), renderer);
            }
        }
    }
}
/**
 * Looks for a view with a given view block id inside a provided LContainer.
 * Removes views that need to be deleted in the process.
 *
 * @param containerNode where to search for views
 * @param startIdx starting index in the views array to search from
 * @param viewBlockId exact view block id to look for
 * @returns index of a found view or -1 if not found
 */
function scanForView(containerNode, startIdx, viewBlockId) {
    var views = containerNode.data.views;
    for (var i = startIdx; i < views.length; i++) {
        var viewAtPositionId = views[i].data.id;
        if (viewAtPositionId === viewBlockId) {
            return views[i];
        }
        else if (viewAtPositionId < viewBlockId) {
            // found a view that should not be at this position - remove
            // found a view that should not be at this position - remove
            node_manipulation_1.removeView(containerNode, i);
        }
        else {
            // found a view with id grater than the one we are searching for
            // which means that required view doesn't exist and can't be found at
            // later positions in the views array - stop the search here
            break;
        }
    }
    return null;
}
/**
 * Marks the start of an embedded view.
 *
 * @param viewBlockId The ID of this view
 * @return boolean Whether or not this view is in creation mode
 */
function embeddedViewStart(viewBlockId) {
    var container = (isParent ? previousOrParentNode : previousOrParentNode.parent);
    ngDevMode && node_assert_1.assertNodeType(container, 0 /* Container */);
    var lContainer = container.data;
    var viewNode = scanForView(container, lContainer.nextIndex, viewBlockId);
    if (viewNode) {
        previousOrParentNode = viewNode;
        ngDevMode && node_assert_1.assertNodeType(previousOrParentNode, 2 /* View */);
        isParent = true;
        enterView(viewNode.data, viewNode);
    }
    else {
        // When we create a new LView, we always reset the state of the instructions.
        var newView = createLView(viewBlockId, renderer, getOrCreateEmbeddedTView(viewBlockId, container), null, null, 2 /* CheckAlways */);
        if (lContainer.queries) {
            newView.queries = lContainer.queries.enterView(lContainer.nextIndex);
        }
        enterView(newView, viewNode = createLNode(null, 2 /* View */, null, newView));
    }
    return getRenderFlags(viewNode.data);
}
exports.embeddedViewStart = embeddedViewStart;
/**
 * Initialize the TView (e.g. static data) for the active embedded view.
 *
 * Each embedded view needs to set the global tData variable to the static data for
 * that view. Otherwise, the view's static data for a particular node would overwrite
 * the static data for a node in the view above it with the same index (since it's in the
 * same template).
 *
 * @param viewIndex The index of the TView in TContainer
 * @param parent The parent container in which to look for the view's static data
 * @returns TView
 */
function getOrCreateEmbeddedTView(viewIndex, parent) {
    ngDevMode && node_assert_1.assertNodeType(parent, 0 /* Container */);
    var tContainer = parent.tNode.data;
    if (viewIndex >= tContainer.length || tContainer[viewIndex] == null) {
        var tView = currentView.tView;
        tContainer[viewIndex] = createTView(tView.directiveRegistry, tView.pipeRegistry);
    }
    return tContainer[viewIndex];
}
/** Marks the end of an embedded view. */
function embeddedViewEnd() {
    refreshDirectives();
    isParent = false;
    var viewNode = previousOrParentNode = currentView.node;
    var containerNode = previousOrParentNode.parent;
    if (containerNode) {
        ngDevMode && node_assert_1.assertNodeType(viewNode, 2 /* View */);
        ngDevMode && node_assert_1.assertNodeType(containerNode, 0 /* Container */);
        var lContainer = containerNode.data;
        if (creationMode) {
            // When projected nodes are going to be inserted, the renderParent of the dynamic container
            // used by the ViewContainerRef must be set.
            setRenderParentInProjectedNodes(lContainer.renderParent, viewNode);
            // it is a new view, insert it into collection of views for a given container
            // it is a new view, insert it into collection of views for a given container
            node_manipulation_1.insertView(containerNode, viewNode, lContainer.nextIndex);
        }
        lContainer.nextIndex++;
    }
    leaveView((currentView.parent));
    ngDevMode && assert_1.assertEqual(isParent, false, 'isParent');
    ngDevMode && node_assert_1.assertNodeType(previousOrParentNode, 2 /* View */);
}
exports.embeddedViewEnd = embeddedViewEnd;
/**
 * For nodes which are projected inside an embedded view, this function sets the renderParent
 * of their dynamic LContainerNode.
 * @param renderParent the renderParent of the LContainer which contains the embedded view.
 * @param viewNode the embedded view.
 */
function setRenderParentInProjectedNodes(renderParent, viewNode) {
    if (renderParent != null) {
        var node = viewNode.child;
        while (node) {
            if (node.type === 1 /* Projection */) {
                var nodeToProject = node.data.head;
                var lastNodeToProject = node.data.tail;
                while (nodeToProject) {
                    if (nodeToProject.dynamicLContainerNode) {
                        nodeToProject.dynamicLContainerNode.data.renderParent = renderParent;
                    }
                    nodeToProject = nodeToProject === lastNodeToProject ? null : nodeToProject.pNextOrParent;
                }
            }
            node = node.next;
        }
    }
}
/**
 * Refreshes components by entering the component view and processing its bindings, queries, etc.
 *
 * @param directiveIndex
 * @param elementIndex
 */
function componentRefresh(directiveIndex, elementIndex) {
    ngDevMode && assertDataInRange(elementIndex);
    var element = data[elementIndex];
    ngDevMode && node_assert_1.assertNodeType(element, 3 /* Element */);
    ngDevMode && assert_1.assertNotNull(element.data, "Component's host node should have an LView attached.");
    var hostView = (element.data);
    // Only attached CheckAlways components or attached, dirty OnPush components should be checked
    if (viewAttached(hostView) && hostView.flags & (2 /* CheckAlways */ | 4 /* Dirty */)) {
        ngDevMode && assertDataInRange(directiveIndex, (directives));
        var def = currentView.tView.directives[directiveIndex];
        detectChangesInternal(hostView, element, def, getDirectiveInstance(directives[directiveIndex]));
    }
}
exports.componentRefresh = componentRefresh;
/** Returns a boolean for whether the view is attached */
function viewAttached(view) {
    return (view.flags & 8 /* Attached */) === 8 /* Attached */;
}
/**
 * Instruction to distribute projectable nodes among <ng-content> occurrences in a given template.
 * It takes all the selectors from the entire component's template and decides where
 * each projected node belongs (it re-distributes nodes among "buckets" where each "bucket" is
 * backed by a selector).
 *
 * This function requires CSS selectors to be provided in 2 forms: parsed (by a compiler) and text,
 * un-parsed form.
 *
 * The parsed form is needed for efficient matching of a node against a given CSS selector.
 * The un-parsed, textual form is needed for support of the ngProjectAs attribute.
 *
 * Having a CSS selector in 2 different formats is not ideal, but alternatives have even more
 * drawbacks:
 * - having only a textual form would require runtime parsing of CSS selectors;
 * - we can't have only a parsed as we can't re-construct textual form from it (as entered by a
 * template author).
 *
 * @param selectors A collection of parsed CSS selectors
 * @param rawSelectors A collection of CSS selectors in the raw, un-parsed form
 */
function projectionDef(index, selectors, textSelectors) {
    var noOfNodeBuckets = selectors ? selectors.length + 1 : 1;
    var distributedNodes = new Array(noOfNodeBuckets);
    for (var i = 0; i < noOfNodeBuckets; i++) {
        distributedNodes[i] = [];
    }
    var componentNode = findComponentHost(currentView);
    var componentChild = componentNode.child;
    while (componentChild !== null) {
        // execute selector matching logic if and only if:
        // - there are selectors defined
        // - a node has a tag name / attributes that can be matched
        if (selectors && componentChild.tNode) {
            var matchedIdx = node_selector_matcher_1.matchingSelectorIndex(componentChild.tNode, selectors, (textSelectors));
            distributedNodes[matchedIdx].push(componentChild);
        }
        else {
            distributedNodes[0].push(componentChild);
        }
        componentChild = componentChild.next;
    }
    ngDevMode && assertDataNext(index);
    data[index] = distributedNodes;
}
exports.projectionDef = projectionDef;
/**
 * Updates the linked list of a projection node, by appending another linked list.
 *
 * @param projectionNode Projection node whose projected nodes linked list has to be updated
 * @param appendedFirst First node of the linked list to append.
 * @param appendedLast Last node of the linked list to append.
 */
function appendToProjectionNode(projectionNode, appendedFirst, appendedLast) {
    ngDevMode && assert_1.assertEqual(!!appendedFirst, !!appendedLast, 'appendedFirst can be null if and only if appendedLast is also null');
    if (!appendedLast) {
        // nothing to append
        return;
    }
    var projectionNodeData = projectionNode.data;
    if (projectionNodeData.tail) {
        projectionNodeData.tail.pNextOrParent = appendedFirst;
    }
    else {
        projectionNodeData.head = appendedFirst;
    }
    projectionNodeData.tail = appendedLast;
    appendedLast.pNextOrParent = projectionNode;
}
/**
 * Inserts previously re-distributed projected nodes. This instruction must be preceded by a call
 * to the projectionDef instruction.
 *
 * @param nodeIndex
 * @param localIndex - index under which distribution of projected nodes was memorized
 * @param selectorIndex - 0 means <ng-content> without any selector
 * @param attrs - attributes attached to the ng-content node, if present
 */
function projection(nodeIndex, localIndex, selectorIndex, attrs) {
    if (selectorIndex === void 0) { selectorIndex = 0; }
    var node = createLNode(nodeIndex, 1 /* Projection */, null, { head: null, tail: null });
    if (node.tNode == null) {
        node.tNode = createTNode(null, attrs || null, null);
    }
    isParent = false; // self closing
    var currentParent = node.parent;
    // re-distribution of projectable nodes is memorized on a component's view level
    var componentNode = findComponentHost(currentView);
    // make sure that nodes to project were memorized
    var nodesForSelector = componentNode.data.data[localIndex][selectorIndex];
    // build the linked list of projected nodes:
    for (var i = 0; i < nodesForSelector.length; i++) {
        var nodeToProject = nodesForSelector[i];
        if (nodeToProject.type === 1 /* Projection */) {
            var previouslyProjected = nodeToProject.data;
            appendToProjectionNode(node, previouslyProjected.head, previouslyProjected.tail);
        }
        else {
            appendToProjectionNode(node, nodeToProject, nodeToProject);
        }
    }
    if (node_manipulation_1.canInsertNativeNode(currentParent, currentView)) {
        ngDevMode && node_assert_1.assertNodeType(currentParent, 3 /* Element */);
        // process each node in the list of projected nodes:
        var nodeToProject = node.data.head;
        var lastNodeToProject = node.data.tail;
        while (nodeToProject) {
            node_manipulation_1.appendProjectedNode(nodeToProject, currentParent, currentView);
            nodeToProject = nodeToProject === lastNodeToProject ? null : nodeToProject.pNextOrParent;
        }
    }
}
exports.projection = projection;
/**
 * Given a current view, finds the nearest component's host (LElement).
 *
 * @param lView LView for which we want a host element node
 * @returns The host node
 */
function findComponentHost(lView) {
    var viewRootLNode = lView.node;
    while (viewRootLNode.type === 2 /* View */) {
        ngDevMode && assert_1.assertNotNull(lView.parent, 'lView.parent');
        lView = (lView.parent);
        viewRootLNode = lView.node;
    }
    ngDevMode && node_assert_1.assertNodeType(viewRootLNode, 3 /* Element */);
    ngDevMode && assert_1.assertNotNull(viewRootLNode.data, 'node.data');
    return viewRootLNode;
}
/**
 * Adds a LView or a LContainer to the end of the current view tree.
 *
 * This structure will be used to traverse through nested views to remove listeners
 * and call onDestroy callbacks.
 *
 * @param currentView The view where LView or LContainer should be added
 * @param state The LView or LContainer to add to the view tree
 * @returns The state passed in
 */
function addToViewTree(currentView, state) {
    currentView.tail ? (currentView.tail.next = state) : (currentView.child = state);
    currentView.tail = state;
    return state;
}
exports.addToViewTree = addToViewTree;
/** If node is an OnPush component, marks its LView dirty. */
function markDirtyIfOnPush(node) {
    // Because data flows down the component tree, ancestors do not need to be marked dirty
    if (node.data && !(node.data.flags & 2 /* CheckAlways */)) {
        node.data.flags |= 4 /* Dirty */;
    }
}
exports.markDirtyIfOnPush = markDirtyIfOnPush;
/**
 * Wraps an event listener so its host view and its ancestor views will be marked dirty
 * whenever the event fires. Necessary to support OnPush components.
 */
function wrapListenerWithDirtyLogic(view, listenerFn) {
    return function (e) {
        markViewDirty(view);
        return listenerFn(e);
    };
}
exports.wrapListenerWithDirtyLogic = wrapListenerWithDirtyLogic;
/**
 * Wraps an event listener so its host view and its ancestor views will be marked dirty
 * whenever the event fires. Also wraps with preventDefault behavior.
 */
function wrapListenerWithDirtyAndDefault(view, listenerFn) {
    return function wrapListenerIn_markViewDirty(e) {
        markViewDirty(view);
        if (listenerFn(e) === false) {
            e.preventDefault();
            // Necessary for legacy browsers that don't support preventDefault (e.g. IE)
            e.returnValue = false;
        }
    };
}
exports.wrapListenerWithDirtyAndDefault = wrapListenerWithDirtyAndDefault;
/** Marks current view and all ancestors dirty */
function markViewDirty(view) {
    var currentView = view;
    while (currentView.parent != null) {
        currentView.flags |= 4 /* Dirty */;
        currentView = currentView.parent;
    }
    currentView.flags |= 4 /* Dirty */;
    ngDevMode && assert_1.assertNotNull(currentView.context, 'rootContext');
    scheduleTick(currentView.context);
}
exports.markViewDirty = markViewDirty;
/**
 * Used to schedule change detection on the whole application.
 *
 * Unlike `tick`, `scheduleTick` coalesces multiple calls into one change detection run.
 * It is usually called indirectly by calling `markDirty` when the view needs to be
 * re-rendered.
 *
 * Typically `scheduleTick` uses `requestAnimationFrame` to coalesce multiple
 * `scheduleTick` requests. The scheduling function can be overridden in
 * `renderComponent`'s `scheduler` option.
 */
function scheduleTick(rootContext) {
    if (rootContext.clean == _CLEAN_PROMISE) {
        var res_1;
        rootContext.clean = new Promise(function (r) { return res_1 = r; });
        rootContext.scheduler(function () {
            tick(rootContext.component);
            res_1(null);
            rootContext.clean = _CLEAN_PROMISE;
        });
    }
}
exports.scheduleTick = scheduleTick;
/**
 * Used to perform change detection on the whole application.
 *
 * This is equivalent to `detectChanges`, but invoked on root component. Additionally, `tick`
 * executes lifecycle hooks and conditionally checks components based on their
 * `ChangeDetectionStrategy` and dirtiness.
 *
 * The preferred way to trigger change detection is to call `markDirty`. `markDirty` internally
 * schedules `tick` using a scheduler in order to coalesce multiple `markDirty` calls into a
 * single change detection run. By default, the scheduler is `requestAnimationFrame`, but can
 * be changed when calling `renderComponent` and providing the `scheduler` option.
 */
function tick(component) {
    var rootView = getRootView(component);
    var rootComponent = rootView.context.component;
    var hostNode = _getComponentHostLElementNode(rootComponent);
    ngDevMode && assert_1.assertNotNull(hostNode.data, 'Component host node should be attached to an LView');
    renderComponentOrTemplate(hostNode, rootView, rootComponent);
}
exports.tick = tick;
/**
 * Retrieve the root view from any component by walking the parent `LView` until
 * reaching the root `LView`.
 *
 * @param component any component
 */
function getRootView(component) {
    ngDevMode && assert_1.assertNotNull(component, 'component');
    var lElementNode = _getComponentHostLElementNode(component);
    var lView = lElementNode.view;
    while (lView.parent) {
        lView = lView.parent;
    }
    return lView;
}
exports.getRootView = getRootView;
/**
 * Synchronously perform change detection on a component (and possibly its sub-components).
 *
 * This function triggers change detection in a synchronous way on a component. There should
 * be very little reason to call this function directly since a preferred way to do change
 * detection is to {@link markDirty} the component and wait for the scheduler to call this method
 * at some future point in time. This is because a single user action often results in many
 * components being invalidated and calling change detection on each component synchronously
 * would be inefficient. It is better to wait until all components are marked as dirty and
 * then perform single change detection across all of the components
 *
 * @param component The component which the change detection should be performed on.
 */
function detectChanges(component) {
    var hostNode = _getComponentHostLElementNode(component);
    ngDevMode && assert_1.assertNotNull(hostNode.data, 'Component host node should be attached to an LView');
    var componentIndex = hostNode.tNode.flags >> 13 /* DirectiveStartingIndexShift */;
    var def = hostNode.view.tView.directives[componentIndex];
    detectChangesInternal(hostNode.data, hostNode, def, component);
}
exports.detectChanges = detectChanges;
/**
 * Checks the change detector and its children, and throws if any changes are detected.
 *
 * This is used in development mode to verify that running change detection doesn't
 * introduce other changes.
 */
function checkNoChanges(component) {
    checkNoChangesMode = true;
    try {
        detectChanges(component);
    }
    finally {
        checkNoChangesMode = false;
    }
}
exports.checkNoChanges = checkNoChanges;
/** Checks the view of the component provided. Does not gate on dirty checks or execute doCheck. */
function detectChangesInternal(hostView, hostNode, def, component) {
    var oldView = enterView(hostView, hostNode);
    var template = def.template;
    try {
        template(getRenderFlags(hostView), component);
        refreshDirectives();
        refreshDynamicChildren();
    }
    finally {
        leaveView(oldView);
    }
}
exports.detectChangesInternal = detectChangesInternal;
/**
 * Mark the component as dirty (needing change detection).
 *
 * Marking a component dirty will schedule a change detection on this
 * component at some point in the future. Marking an already dirty
 * component as dirty is a noop. Only one outstanding change detection
 * can be scheduled per component tree. (Two components bootstrapped with
 * separate `renderComponent` will have separate schedulers)
 *
 * When the root component is bootstrapped with `renderComponent`, a scheduler
 * can be provided.
 *
 * @param component Component to mark as dirty.
 */
function markDirty(component) {
    ngDevMode && assert_1.assertNotNull(component, 'component');
    var lElementNode = _getComponentHostLElementNode(component);
    markViewDirty(lElementNode.view);
}
exports.markDirty = markDirty;
/** A special value which designates that a value has not changed. */
exports.NO_CHANGE = {};
/**
 *  Initializes the binding start index. Will get inlined.
 *
 *  This function must be called before any binding related function is called
 *  (ie `bind()`, `interpolationX()`, `pureFunctionX()`)
 */
function initBindings() {
    ngDevMode && assert_1.assertEqual(currentView.bindingStartIndex, -1, 'Binding start index should only be set once, when null');
    ngDevMode && assert_1.assertEqual(currentView.bindingIndex, -1, 'Binding index should not yet be set ' + currentView.bindingIndex);
    currentView.bindingIndex = currentView.bindingStartIndex = data.length;
}
/**
 * Creates a single value binding.
 *
 * @param value Value to diff
 */
function bind(value) {
    if (currentView.bindingStartIndex < 0) {
        initBindings();
        return data[currentView.bindingIndex++] = value;
    }
    var changed = value !== exports.NO_CHANGE && util_1.isDifferent(data[currentView.bindingIndex], value);
    if (changed) {
        errors_1.throwErrorIfNoChangesMode(creationMode, checkNoChangesMode, data[currentView.bindingIndex], value);
        data[currentView.bindingIndex] = value;
    }
    currentView.bindingIndex++;
    return changed ? value : exports.NO_CHANGE;
}
exports.bind = bind;
/**
 * Create interpolation bindings with a variable number of expressions.
 *
 * If there are 1 to 8 expressions `interpolation1()` to `interpolation8()` should be used instead.
 * Those are faster because there is no need to create an array of expressions and iterate over it.
 *
 * `values`:
 * - has static text at even indexes,
 * - has evaluated expressions at odd indexes.
 *
 * Returns the concatenated string when any of the arguments changes, `NO_CHANGE` otherwise.
 */
function interpolationV(values) {
    ngDevMode && assert_1.assertLessThan(2, values.length, 'should have at least 3 values');
    ngDevMode && assert_1.assertEqual(values.length % 2, 1, 'should have an odd number of values');
    var different = false;
    for (var i = 1; i < values.length; i += 2) {
        // Check if bindings (odd indexes) have changed
        bindingUpdated(values[i]) && (different = true);
    }
    if (!different) {
        return exports.NO_CHANGE;
    }
    // Build the updated content
    var content = values[0];
    for (var i = 1; i < values.length; i += 2) {
        content += util_1.stringify(values[i]) + values[i + 1];
    }
    return content;
}
exports.interpolationV = interpolationV;
/**
 * Creates an interpolation binding with 1 expression.
 *
 * @param prefix static value used for concatenation only.
 * @param v0 value checked for change.
 * @param suffix static value used for concatenation only.
 */
function interpolation1(prefix, v0, suffix) {
    var different = bindingUpdated(v0);
    return different ? prefix + util_1.stringify(v0) + suffix : exports.NO_CHANGE;
}
exports.interpolation1 = interpolation1;
/** Creates an interpolation binding with 2 expressions. */
function interpolation2(prefix, v0, i0, v1, suffix) {
    var different = bindingUpdated2(v0, v1);
    return different ? prefix + util_1.stringify(v0) + i0 + util_1.stringify(v1) + suffix : exports.NO_CHANGE;
}
exports.interpolation2 = interpolation2;
/** Creates an interpolation bindings with 3 expressions. */
function interpolation3(prefix, v0, i0, v1, i1, v2, suffix) {
    var different = bindingUpdated2(v0, v1);
    different = bindingUpdated(v2) || different;
    return different ? prefix + util_1.stringify(v0) + i0 + util_1.stringify(v1) + i1 + util_1.stringify(v2) + suffix :
        exports.NO_CHANGE;
}
exports.interpolation3 = interpolation3;
/** Create an interpolation binding with 4 expressions. */
function interpolation4(prefix, v0, i0, v1, i1, v2, i2, v3, suffix) {
    var different = bindingUpdated4(v0, v1, v2, v3);
    return different ?
        prefix + util_1.stringify(v0) + i0 + util_1.stringify(v1) + i1 + util_1.stringify(v2) + i2 + util_1.stringify(v3) +
            suffix :
        exports.NO_CHANGE;
}
exports.interpolation4 = interpolation4;
/** Creates an interpolation binding with 5 expressions. */
function interpolation5(prefix, v0, i0, v1, i1, v2, i2, v3, i3, v4, suffix) {
    var different = bindingUpdated4(v0, v1, v2, v3);
    different = bindingUpdated(v4) || different;
    return different ?
        prefix + util_1.stringify(v0) + i0 + util_1.stringify(v1) + i1 + util_1.stringify(v2) + i2 + util_1.stringify(v3) + i3 +
            util_1.stringify(v4) + suffix :
        exports.NO_CHANGE;
}
exports.interpolation5 = interpolation5;
/** Creates an interpolation binding with 6 expressions. */
function interpolation6(prefix, v0, i0, v1, i1, v2, i2, v3, i3, v4, i4, v5, suffix) {
    var different = bindingUpdated4(v0, v1, v2, v3);
    different = bindingUpdated2(v4, v5) || different;
    return different ?
        prefix + util_1.stringify(v0) + i0 + util_1.stringify(v1) + i1 + util_1.stringify(v2) + i2 + util_1.stringify(v3) + i3 +
            util_1.stringify(v4) + i4 + util_1.stringify(v5) + suffix :
        exports.NO_CHANGE;
}
exports.interpolation6 = interpolation6;
/** Creates an interpolation binding with 7 expressions. */
function interpolation7(prefix, v0, i0, v1, i1, v2, i2, v3, i3, v4, i4, v5, i5, v6, suffix) {
    var different = bindingUpdated4(v0, v1, v2, v3);
    different = bindingUpdated2(v4, v5) || different;
    different = bindingUpdated(v6) || different;
    return different ?
        prefix + util_1.stringify(v0) + i0 + util_1.stringify(v1) + i1 + util_1.stringify(v2) + i2 + util_1.stringify(v3) + i3 +
            util_1.stringify(v4) + i4 + util_1.stringify(v5) + i5 + util_1.stringify(v6) + suffix :
        exports.NO_CHANGE;
}
exports.interpolation7 = interpolation7;
/** Creates an interpolation binding with 8 expressions. */
function interpolation8(prefix, v0, i0, v1, i1, v2, i2, v3, i3, v4, i4, v5, i5, v6, i6, v7, suffix) {
    var different = bindingUpdated4(v0, v1, v2, v3);
    different = bindingUpdated4(v4, v5, v6, v7) || different;
    return different ?
        prefix + util_1.stringify(v0) + i0 + util_1.stringify(v1) + i1 + util_1.stringify(v2) + i2 + util_1.stringify(v3) + i3 +
            util_1.stringify(v4) + i4 + util_1.stringify(v5) + i5 + util_1.stringify(v6) + i6 + util_1.stringify(v7) + suffix :
        exports.NO_CHANGE;
}
exports.interpolation8 = interpolation8;
/** Store a value in the `data` at a given `index`. */
function store(index, value) {
    // We don't store any static data for local variables, so the first time
    // we see the template, we should store as null to avoid a sparse array
    if (index >= tData.length) {
        tData[index] = null;
    }
    data[index] = value;
}
exports.store = store;
/** Retrieves a value from the `data`. */
function load(index) {
    ngDevMode && assertDataInRange(index);
    return data[index];
}
exports.load = load;
/** Retrieves a value from the `directives` array. */
function loadDirective(index) {
    ngDevMode && assert_1.assertNotNull(directives, 'Directives array should be defined if reading a dir.');
    ngDevMode && assertDataInRange(index, (directives));
    return directives[index];
}
exports.loadDirective = loadDirective;
/** Gets the current binding value and increments the binding index. */
function consumeBinding() {
    ngDevMode && assertDataInRange(currentView.bindingIndex);
    ngDevMode &&
        assert_1.assertNotEqual(data[currentView.bindingIndex], exports.NO_CHANGE, 'Stored value should never be NO_CHANGE.');
    return data[currentView.bindingIndex++];
}
exports.consumeBinding = consumeBinding;
/** Updates binding if changed, then returns whether it was updated. */
function bindingUpdated(value) {
    ngDevMode && assert_1.assertNotEqual(value, exports.NO_CHANGE, 'Incoming value should never be NO_CHANGE.');
    if (currentView.bindingStartIndex < 0) {
        initBindings();
    }
    else if (util_1.isDifferent(data[currentView.bindingIndex], value)) {
        errors_1.throwErrorIfNoChangesMode(creationMode, checkNoChangesMode, data[currentView.bindingIndex], value);
    }
    else {
        currentView.bindingIndex++;
        return false;
    }
    data[currentView.bindingIndex++] = value;
    return true;
}
exports.bindingUpdated = bindingUpdated;
/** Updates binding if changed, then returns the latest value. */
function checkAndUpdateBinding(value) {
    bindingUpdated(value);
    return value;
}
exports.checkAndUpdateBinding = checkAndUpdateBinding;
/** Updates 2 bindings if changed, then returns whether either was updated. */
function bindingUpdated2(exp1, exp2) {
    var different = bindingUpdated(exp1);
    return bindingUpdated(exp2) || different;
}
exports.bindingUpdated2 = bindingUpdated2;
/** Updates 4 bindings if changed, then returns whether any was updated. */
function bindingUpdated4(exp1, exp2, exp3, exp4) {
    var different = bindingUpdated2(exp1, exp2);
    return bindingUpdated2(exp3, exp4) || different;
}
exports.bindingUpdated4 = bindingUpdated4;
function getTView() {
    return currentView.tView;
}
exports.getTView = getTView;
function getDirectiveInstance(instanceOrArray) {
    // Directives with content queries store an array in directives[directiveIndex]
    // with the instance as the first index
    return Array.isArray(instanceOrArray) ? instanceOrArray[0] : instanceOrArray;
}
exports.getDirectiveInstance = getDirectiveInstance;
function assertPreviousIsParent() {
    assert_1.assertEqual(isParent, true, 'previousOrParentNode should be a parent');
}
exports.assertPreviousIsParent = assertPreviousIsParent;
function assertHasParent() {
    assert_1.assertNotNull(previousOrParentNode.parent, 'previousOrParentNode should have a parent');
}
function assertDataInRange(index, arr) {
    if (arr == null)
        arr = data;
    assert_1.assertLessThan(index, arr ? arr.length : 0, 'index expected to be a valid data index');
}
function assertDataNext(index, arr) {
    if (arr == null)
        arr = data;
    assert_1.assertEqual(arr.length, index, "index " + index + " expected to be at the end of arr (length " + arr.length + ")");
}
function _getComponentHostLElementNode(component) {
    ngDevMode && assert_1.assertNotNull(component, 'expecting component got null');
    var lElementNode = component[exports.NG_HOST_SYMBOL];
    ngDevMode && assert_1.assertNotNull(component, 'object is not a component');
    return lElementNode;
}
exports._getComponentHostLElementNode = _getComponentHostLElementNode;
exports.CLEAN_PROMISE = _CLEAN_PROMISE;
exports.ROOT_DIRECTIVE_INDICES = exports._ROOT_DIRECTIVE_INDICES;
//# sourceMappingURL=instructions.js.map