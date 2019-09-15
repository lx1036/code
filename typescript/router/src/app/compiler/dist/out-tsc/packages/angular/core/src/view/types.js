"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
// Called before each cycle of a view's check to detect whether this is in the
// initState for which we need to call ngOnInit, ngAfterContentInit or ngAfterViewInit
// lifecycle methods. Returns true if this check cycle should call lifecycle
// methods.
function shiftInitState(view, priorInitState, newInitState) {
    // Only update the InitState if we are currently in the prior state.
    // For example, only move into CallingInit if we are in BeforeInit. Only
    // move into CallingContentInit if we are in CallingInit. Normally this will
    // always be true because of how checkCycle is called in checkAndUpdateView.
    // However, if checkAndUpdateView is called recursively or if an exception is
    // thrown while checkAndUpdateView is running, checkAndUpdateView starts over
    // from the beginning. This ensures the state is monotonically increasing,
    // terminating in the AfterInit state, which ensures the Init methods are called
    // at least once and only once.
    var state = view.state;
    var initState = state & 1792 /* InitState_Mask */;
    if (initState === priorInitState) {
        view.state = (state & ~1792 /* InitState_Mask */) | newInitState;
        view.initIndex = -1;
        return true;
    }
    return initState === newInitState;
}
exports.shiftInitState = shiftInitState;
// Returns true if the lifecycle init method should be called for the node with
// the given init index.
function shouldCallLifecycleInitHook(view, initState, index) {
    if ((view.state & 1792 /* InitState_Mask */) === initState && view.initIndex <= index) {
        view.initIndex = index + 1;
        return true;
    }
    return false;
}
exports.shouldCallLifecycleInitHook = shouldCallLifecycleInitHook;
/**
 * Node instance data.
 *
 * We have a separate type per NodeType to save memory
 * (TextData | ElementData | ProviderData | PureExpressionData | QueryList<any>)
 *
 * To keep our code monomorphic,
 * we prohibit using `NodeData` directly but enforce the use of accessors (`asElementData`, ...).
 * This way, no usage site can get a `NodeData` from view.nodes and then use it for different
 * purposes.
 */
var /**
 * Node instance data.
 *
 * We have a separate type per NodeType to save memory
 * (TextData | ElementData | ProviderData | PureExpressionData | QueryList<any>)
 *
 * To keep our code monomorphic,
 * we prohibit using `NodeData` directly but enforce the use of accessors (`asElementData`, ...).
 * This way, no usage site can get a `NodeData` from view.nodes and then use it for different
 * purposes.
 */
NodeData = /** @class */ (function () {
    function NodeData() {
    }
    return NodeData;
}());
exports.NodeData = NodeData;
/**
 * Accessor for view.nodes, enforcing that every usage site stays monomorphic.
 */
function asTextData(view, index) {
    return view.nodes[index];
}
exports.asTextData = asTextData;
/**
 * Accessor for view.nodes, enforcing that every usage site stays monomorphic.
 */
function asElementData(view, index) {
    return view.nodes[index];
}
exports.asElementData = asElementData;
/**
 * Accessor for view.nodes, enforcing that every usage site stays monomorphic.
 */
function asProviderData(view, index) {
    return view.nodes[index];
}
exports.asProviderData = asProviderData;
/**
 * Accessor for view.nodes, enforcing that every usage site stays monomorphic.
 */
function asPureExpressionData(view, index) {
    return view.nodes[index];
}
exports.asPureExpressionData = asPureExpressionData;
/**
 * Accessor for view.nodes, enforcing that every usage site stays monomorphic.
 */
function asQueryList(view, index) {
    return view.nodes[index];
}
exports.asQueryList = asQueryList;
var DebugContext = /** @class */ (function () {
    function DebugContext() {
    }
    return DebugContext;
}());
exports.DebugContext = DebugContext;
/**
 * This object is used to prevent cycles in the source files and to have a place where
 * debug mode can hook it. It is lazily filled when `isDevMode` is known.
 */
exports.Services = {
    setCurrentNode: (undefined),
    createRootView: (undefined),
    createEmbeddedView: (undefined),
    createComponentView: (undefined),
    createNgModuleRef: (undefined),
    overrideProvider: (undefined),
    overrideComponentView: (undefined),
    clearOverrides: (undefined),
    checkAndUpdateView: (undefined),
    checkNoChangesView: (undefined),
    destroyView: (undefined),
    resolveDep: (undefined),
    createDebugContext: (undefined),
    handleEvent: (undefined),
    updateDirectives: (undefined),
    updateRenderer: (undefined),
    dirtyParentQueries: (undefined),
};
//# sourceMappingURL=types.js.map