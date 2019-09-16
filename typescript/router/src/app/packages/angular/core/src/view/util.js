"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
exports.__esModule = true;
var change_detection_1 = require("../change_detection/change_detection");
var injector_1 = require("../di/injector");
var view_1 = require("../metadata/view");
var util_1 = require("../util");
var errors_1 = require("./errors");
var types_1 = require("./types");
exports.NOOP = function () { };
var _tokenKeyCache = new Map();
function tokenKey(token) {
    var key = _tokenKeyCache.get(token);
    if (!key) {
        key = util_1.stringify(token) + '_' + _tokenKeyCache.size;
        _tokenKeyCache.set(token, key);
    }
    return key;
}
exports.tokenKey = tokenKey;
function unwrapValue(view, nodeIdx, bindingIdx, value) {
    if (change_detection_1.WrappedValue.isWrapped(value)) {
        value = change_detection_1.WrappedValue.unwrap(value);
        var globalBindingIdx = view.def.nodes[nodeIdx].bindingIndex + bindingIdx;
        var oldValue = change_detection_1.WrappedValue.unwrap(view.oldValues[globalBindingIdx]);
        view.oldValues[globalBindingIdx] = new change_detection_1.WrappedValue(oldValue);
    }
    return value;
}
exports.unwrapValue = unwrapValue;
var UNDEFINED_RENDERER_TYPE_ID = '$$undefined';
var EMPTY_RENDERER_TYPE_ID = '$$empty';
// Attention: this function is called as top level function.
// Putting any logic in here will destroy closure tree shaking!
function createRendererType2(values) {
    return {
        id: UNDEFINED_RENDERER_TYPE_ID,
        styles: values.styles,
        encapsulation: values.encapsulation,
        data: values.data
    };
}
exports.createRendererType2 = createRendererType2;
var _renderCompCount = 0;
function resolveRendererType2(type) {
    if (type && type.id === UNDEFINED_RENDERER_TYPE_ID) {
        // first time we see this RendererType2. Initialize it...
        var isFilled = ((type.encapsulation != null && type.encapsulation !== view_1.ViewEncapsulation.None) ||
            type.styles.length || Object.keys(type.data).length);
        if (isFilled) {
            type.id = "c" + _renderCompCount++;
        }
        else {
            type.id = EMPTY_RENDERER_TYPE_ID;
        }
    }
    if (type && type.id === EMPTY_RENDERER_TYPE_ID) {
        type = null;
    }
    return type || null;
}
exports.resolveRendererType2 = resolveRendererType2;
function checkBinding(view, def, bindingIdx, value) {
    var oldValues = view.oldValues;
    if ((view.state & 2 /* FirstCheck */) ||
        !util_1.looseIdentical(oldValues[def.bindingIndex + bindingIdx], value)) {
        return true;
    }
    return false;
}
exports.checkBinding = checkBinding;
function checkAndUpdateBinding(view, def, bindingIdx, value) {
    if (checkBinding(view, def, bindingIdx, value)) {
        view.oldValues[def.bindingIndex + bindingIdx] = value;
        return true;
    }
    return false;
}
exports.checkAndUpdateBinding = checkAndUpdateBinding;
function checkBindingNoChanges(view, def, bindingIdx, value) {
    var oldValue = view.oldValues[def.bindingIndex + bindingIdx];
    if ((view.state & 1 /* BeforeFirstCheck */) || !change_detection_1.devModeEqual(oldValue, value)) {
        var bindingName = def.bindings[bindingIdx].name;
        throw errors_1.expressionChangedAfterItHasBeenCheckedError(types_1.Services.createDebugContext(view, def.nodeIndex), bindingName + ": " + oldValue, bindingName + ": " + value, (view.state & 1 /* BeforeFirstCheck */) !== 0);
    }
}
exports.checkBindingNoChanges = checkBindingNoChanges;
function markParentViewsForCheck(view) {
    var currView = view;
    while (currView) {
        if (currView.def.flags & 2 /* OnPush */) {
            currView.state |= 8 /* ChecksEnabled */;
        }
        currView = currView.viewContainerParent || currView.parent;
    }
}
exports.markParentViewsForCheck = markParentViewsForCheck;
function markParentViewsForCheckProjectedViews(view, endView) {
    var currView = view;
    while (currView && currView !== endView) {
        currView.state |= 64 /* CheckProjectedViews */;
        currView = currView.viewContainerParent || currView.parent;
    }
}
exports.markParentViewsForCheckProjectedViews = markParentViewsForCheckProjectedViews;
function dispatchEvent(view, nodeIndex, eventName, event) {
    try {
        var nodeDef = view.def.nodes[nodeIndex];
        var startView = nodeDef.flags & 33554432 /* ComponentView */ ?
            types_1.asElementData(view, nodeIndex).componentView :
            view;
        markParentViewsForCheck(startView);
        return types_1.Services.handleEvent(view, nodeIndex, eventName, event);
    }
    catch (e) {
        // Attention: Don't rethrow, as it would cancel Observable subscriptions!
        view.root.errorHandler.handleError(e);
    }
}
exports.dispatchEvent = dispatchEvent;
function declaredViewContainer(view) {
    if (view.parent) {
        var parentView = view.parent;
        return types_1.asElementData(parentView, view.parentNodeDef.nodeIndex);
    }
    return null;
}
exports.declaredViewContainer = declaredViewContainer;
/**
 * for component views, this is the host element.
 * for embedded views, this is the index of the parent node
 * that contains the view container.
 */
function viewParentEl(view) {
    var parentView = view.parent;
    if (parentView) {
        return view.parentNodeDef.parent;
    }
    else {
        return null;
    }
}
exports.viewParentEl = viewParentEl;
function renderNode(view, def) {
    switch (def.flags & 201347067 /* Types */) {
        case 1 /* TypeElement */:
            return types_1.asElementData(view, def.nodeIndex).renderElement;
        case 2 /* TypeText */:
            return types_1.asTextData(view, def.nodeIndex).renderText;
    }
}
exports.renderNode = renderNode;
function elementEventFullName(target, name) {
    return target ? target + ":" + name : name;
}
exports.elementEventFullName = elementEventFullName;
function isComponentView(view) {
    return !!view.parent && !!(view.parentNodeDef.flags & 32768 /* Component */);
}
exports.isComponentView = isComponentView;
function isEmbeddedView(view) {
    return !!view.parent && !(view.parentNodeDef.flags & 32768 /* Component */);
}
exports.isEmbeddedView = isEmbeddedView;
function filterQueryId(queryId) {
    return 1 << (queryId % 32);
}
exports.filterQueryId = filterQueryId;
function splitMatchedQueriesDsl(matchedQueriesDsl) {
    var matchedQueries = {};
    var matchedQueryIds = 0;
    var references = {};
    if (matchedQueriesDsl) {
        matchedQueriesDsl.forEach(function (_a) {
            var queryId = _a[0], valueType = _a[1];
            if (typeof queryId === 'number') {
                matchedQueries[queryId] = valueType;
                matchedQueryIds |= filterQueryId(queryId);
            }
            else {
                references[queryId] = valueType;
            }
        });
    }
    return { matchedQueries: matchedQueries, references: references, matchedQueryIds: matchedQueryIds };
}
exports.splitMatchedQueriesDsl = splitMatchedQueriesDsl;
function splitDepsDsl(deps, sourceName) {
    return deps.map(function (value) {
        var token;
        var flags;
        if (Array.isArray(value)) {
            flags = value[0], token = value[1];
        }
        else {
            flags = 0 /* None */;
            token = value;
        }
        if (token && (typeof token === 'function' || typeof token === 'object') && sourceName) {
            Object.defineProperty(token, injector_1.SOURCE, { value: sourceName, configurable: true });
        }
        return { flags: flags, token: token, tokenKey: tokenKey(token) };
    });
}
exports.splitDepsDsl = splitDepsDsl;
function getParentRenderElement(view, renderHost, def) {
    var renderParent = def.renderParent;
    if (renderParent) {
        if ((renderParent.flags & 1 /* TypeElement */) === 0 ||
            (renderParent.flags & 33554432 /* ComponentView */) === 0 ||
            (renderParent.element.componentRendererType &&
                renderParent.element.componentRendererType.encapsulation ===
                    view_1.ViewEncapsulation.Native)) {
            // only children of non components, or children of components with native encapsulation should
            // be attached.
            return types_1.asElementData(view, def.renderParent.nodeIndex).renderElement;
        }
    }
    else {
        return renderHost;
    }
}
exports.getParentRenderElement = getParentRenderElement;
var DEFINITION_CACHE = new WeakMap();
function resolveDefinition(factory) {
    var value = DEFINITION_CACHE.get(factory);
    if (!value) {
        value = factory(function () { return exports.NOOP; });
        value.factory = factory;
        DEFINITION_CACHE.set(factory, value);
    }
    return value;
}
exports.resolveDefinition = resolveDefinition;
function rootRenderNodes(view) {
    var renderNodes = [];
    visitRootRenderNodes(view, 0 /* Collect */, undefined, undefined, renderNodes);
    return renderNodes;
}
exports.rootRenderNodes = rootRenderNodes;
function visitRootRenderNodes(view, action, parentNode, nextSibling, target) {
    // We need to re-compute the parent node in case the nodes have been moved around manually
    if (action === 3 /* RemoveChild */) {
        parentNode = view.renderer.parentNode(renderNode(view, view.def.lastRenderRootNode));
    }
    visitSiblingRenderNodes(view, action, 0, view.def.nodes.length - 1, parentNode, nextSibling, target);
}
exports.visitRootRenderNodes = visitRootRenderNodes;
function visitSiblingRenderNodes(view, action, startIndex, endIndex, parentNode, nextSibling, target) {
    for (var i = startIndex; i <= endIndex; i++) {
        var nodeDef = view.def.nodes[i];
        if (nodeDef.flags & (1 /* TypeElement */ | 2 /* TypeText */ | 8 /* TypeNgContent */)) {
            visitRenderNode(view, nodeDef, action, parentNode, nextSibling, target);
        }
        // jump to next sibling
        i += nodeDef.childCount;
    }
}
exports.visitSiblingRenderNodes = visitSiblingRenderNodes;
function visitProjectedRenderNodes(view, ngContentIndex, action, parentNode, nextSibling, target) {
    var compView = view;
    while (compView && !isComponentView(compView)) {
        compView = compView.parent;
    }
    var hostView = compView.parent;
    var hostElDef = viewParentEl(compView);
    var startIndex = hostElDef.nodeIndex + 1;
    var endIndex = hostElDef.nodeIndex + hostElDef.childCount;
    for (var i = startIndex; i <= endIndex; i++) {
        var nodeDef = hostView.def.nodes[i];
        if (nodeDef.ngContentIndex === ngContentIndex) {
            visitRenderNode(hostView, nodeDef, action, parentNode, nextSibling, target);
        }
        // jump to next sibling
        i += nodeDef.childCount;
    }
    if (!hostView.parent) {
        // a root view
        var projectedNodes = view.root.projectableNodes[ngContentIndex];
        if (projectedNodes) {
            for (var i = 0; i < projectedNodes.length; i++) {
                execRenderNodeAction(view, projectedNodes[i], action, parentNode, nextSibling, target);
            }
        }
    }
}
exports.visitProjectedRenderNodes = visitProjectedRenderNodes;
function visitRenderNode(view, nodeDef, action, parentNode, nextSibling, target) {
    if (nodeDef.flags & 8 /* TypeNgContent */) {
        visitProjectedRenderNodes(view, nodeDef.ngContent.index, action, parentNode, nextSibling, target);
    }
    else {
        var rn = renderNode(view, nodeDef);
        if (action === 3 /* RemoveChild */ && (nodeDef.flags & 33554432 /* ComponentView */) &&
            (nodeDef.bindingFlags & 48 /* CatSyntheticProperty */)) {
            // Note: we might need to do both actions.
            if (nodeDef.bindingFlags & (16 /* SyntheticProperty */)) {
                execRenderNodeAction(view, rn, action, parentNode, nextSibling, target);
            }
            if (nodeDef.bindingFlags & (32 /* SyntheticHostProperty */)) {
                var compView = types_1.asElementData(view, nodeDef.nodeIndex).componentView;
                execRenderNodeAction(compView, rn, action, parentNode, nextSibling, target);
            }
        }
        else {
            execRenderNodeAction(view, rn, action, parentNode, nextSibling, target);
        }
        if (nodeDef.flags & 16777216 /* EmbeddedViews */) {
            var embeddedViews = types_1.asElementData(view, nodeDef.nodeIndex).viewContainer._embeddedViews;
            for (var k = 0; k < embeddedViews.length; k++) {
                visitRootRenderNodes(embeddedViews[k], action, parentNode, nextSibling, target);
            }
        }
        if (nodeDef.flags & 1 /* TypeElement */ && !nodeDef.element.name) {
            visitSiblingRenderNodes(view, action, nodeDef.nodeIndex + 1, nodeDef.nodeIndex + nodeDef.childCount, parentNode, nextSibling, target);
        }
    }
}
function execRenderNodeAction(view, renderNode, action, parentNode, nextSibling, target) {
    var renderer = view.renderer;
    switch (action) {
        case 1 /* AppendChild */:
            renderer.appendChild(parentNode, renderNode);
            break;
        case 2 /* InsertBefore */:
            renderer.insertBefore(parentNode, renderNode, nextSibling);
            break;
        case 3 /* RemoveChild */:
            renderer.removeChild(parentNode, renderNode);
            break;
        case 0 /* Collect */:
            target.push(renderNode);
            break;
    }
}
var NS_PREFIX_RE = /^:([^:]+):(.+)$/;
function splitNamespace(name) {
    if (name[0] === ':') {
        var match = name.match(NS_PREFIX_RE);
        return [match[1], match[2]];
    }
    return ['', name];
}
exports.splitNamespace = splitNamespace;
function calcBindingFlags(bindings) {
    var flags = 0;
    for (var i = 0; i < bindings.length; i++) {
        flags |= bindings[i].flags;
    }
    return flags;
}
exports.calcBindingFlags = calcBindingFlags;
function interpolate(valueCount, constAndInterp) {
    var result = '';
    for (var i = 0; i < valueCount * 2; i = i + 2) {
        result = result + constAndInterp[i] + _toStringWithNull(constAndInterp[i + 1]);
    }
    return result + constAndInterp[valueCount * 2];
}
exports.interpolate = interpolate;
function inlineInterpolate(valueCount, c0, a1, c1, a2, c2, a3, c3, a4, c4, a5, c5, a6, c6, a7, c7, a8, c8, a9, c9) {
    switch (valueCount) {
        case 1:
            return c0 + _toStringWithNull(a1) + c1;
        case 2:
            return c0 + _toStringWithNull(a1) + c1 + _toStringWithNull(a2) + c2;
        case 3:
            return c0 + _toStringWithNull(a1) + c1 + _toStringWithNull(a2) + c2 + _toStringWithNull(a3) +
                c3;
        case 4:
            return c0 + _toStringWithNull(a1) + c1 + _toStringWithNull(a2) + c2 + _toStringWithNull(a3) +
                c3 + _toStringWithNull(a4) + c4;
        case 5:
            return c0 + _toStringWithNull(a1) + c1 + _toStringWithNull(a2) + c2 + _toStringWithNull(a3) +
                c3 + _toStringWithNull(a4) + c4 + _toStringWithNull(a5) + c5;
        case 6:
            return c0 + _toStringWithNull(a1) + c1 + _toStringWithNull(a2) + c2 + _toStringWithNull(a3) +
                c3 + _toStringWithNull(a4) + c4 + _toStringWithNull(a5) + c5 + _toStringWithNull(a6) + c6;
        case 7:
            return c0 + _toStringWithNull(a1) + c1 + _toStringWithNull(a2) + c2 + _toStringWithNull(a3) +
                c3 + _toStringWithNull(a4) + c4 + _toStringWithNull(a5) + c5 + _toStringWithNull(a6) +
                c6 + _toStringWithNull(a7) + c7;
        case 8:
            return c0 + _toStringWithNull(a1) + c1 + _toStringWithNull(a2) + c2 + _toStringWithNull(a3) +
                c3 + _toStringWithNull(a4) + c4 + _toStringWithNull(a5) + c5 + _toStringWithNull(a6) +
                c6 + _toStringWithNull(a7) + c7 + _toStringWithNull(a8) + c8;
        case 9:
            return c0 + _toStringWithNull(a1) + c1 + _toStringWithNull(a2) + c2 + _toStringWithNull(a3) +
                c3 + _toStringWithNull(a4) + c4 + _toStringWithNull(a5) + c5 + _toStringWithNull(a6) +
                c6 + _toStringWithNull(a7) + c7 + _toStringWithNull(a8) + c8 + _toStringWithNull(a9) + c9;
        default:
            throw new Error("Does not support more than 9 expressions");
    }
}
exports.inlineInterpolate = inlineInterpolate;
function _toStringWithNull(v) {
    return v != null ? v.toString() : '';
}
exports.EMPTY_ARRAY = [];
exports.EMPTY_MAP = {};
