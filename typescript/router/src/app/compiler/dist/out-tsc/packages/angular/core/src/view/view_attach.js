"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var types_1 = require("./types");
var util_1 = require("./util");
function attachEmbeddedView(parentView, elementData, viewIndex, view) {
    var embeddedViews = elementData.viewContainer._embeddedViews;
    if (viewIndex === null || viewIndex === undefined) {
        viewIndex = embeddedViews.length;
    }
    view.viewContainerParent = parentView;
    addToArray(embeddedViews, (viewIndex), view);
    attachProjectedView(elementData, view);
    types_1.Services.dirtyParentQueries(view);
    var prevView = (viewIndex) > 0 ? embeddedViews[(viewIndex) - 1] : null;
    renderAttachEmbeddedView(elementData, prevView, view);
}
exports.attachEmbeddedView = attachEmbeddedView;
function attachProjectedView(vcElementData, view) {
    var dvcElementData = util_1.declaredViewContainer(view);
    if (!dvcElementData || dvcElementData === vcElementData ||
        view.state & 16 /* IsProjectedView */) {
        return;
    }
    // Note: For performance reasons, we
    // - add a view to template._projectedViews only 1x throughout its lifetime,
    //   and remove it not until the view is destroyed.
    //   (hard, as when a parent view is attached/detached we would need to attach/detach all
    //    nested projected views as well, even across component boundaries).
    // - don't track the insertion order of views in the projected views array
    //   (hard, as when the views of the same template are inserted different view containers)
    view.state |= 16 /* IsProjectedView */;
    var projectedViews = dvcElementData.template._projectedViews;
    if (!projectedViews) {
        projectedViews = dvcElementData.template._projectedViews = [];
    }
    projectedViews.push(view);
    // Note: we are changing the NodeDef here as we cannot calculate
    // the fact whether a template is used for projection during compilation.
    markNodeAsProjectedTemplate(view.parent.def, (view.parentNodeDef));
}
function markNodeAsProjectedTemplate(viewDef, nodeDef) {
    if (nodeDef.flags & 4 /* ProjectedTemplate */) {
        return;
    }
    viewDef.nodeFlags |= 4 /* ProjectedTemplate */;
    nodeDef.flags |= 4 /* ProjectedTemplate */;
    var parentNodeDef = nodeDef.parent;
    while (parentNodeDef) {
        parentNodeDef.childFlags |= 4 /* ProjectedTemplate */;
        parentNodeDef = parentNodeDef.parent;
    }
}
function detachEmbeddedView(elementData, viewIndex) {
    var embeddedViews = elementData.viewContainer._embeddedViews;
    if (viewIndex == null || viewIndex >= embeddedViews.length) {
        viewIndex = embeddedViews.length - 1;
    }
    if (viewIndex < 0) {
        return null;
    }
    var view = embeddedViews[viewIndex];
    view.viewContainerParent = null;
    removeFromArray(embeddedViews, viewIndex);
    // See attachProjectedView for why we don't update projectedViews here.
    // See attachProjectedView for why we don't update projectedViews here.
    types_1.Services.dirtyParentQueries(view);
    renderDetachView(view);
    return view;
}
exports.detachEmbeddedView = detachEmbeddedView;
function detachProjectedView(view) {
    if (!(view.state & 16 /* IsProjectedView */)) {
        return;
    }
    var dvcElementData = util_1.declaredViewContainer(view);
    if (dvcElementData) {
        var projectedViews = dvcElementData.template._projectedViews;
        if (projectedViews) {
            removeFromArray(projectedViews, projectedViews.indexOf(view));
            types_1.Services.dirtyParentQueries(view);
        }
    }
}
exports.detachProjectedView = detachProjectedView;
function moveEmbeddedView(elementData, oldViewIndex, newViewIndex) {
    var embeddedViews = elementData.viewContainer._embeddedViews;
    var view = embeddedViews[oldViewIndex];
    removeFromArray(embeddedViews, oldViewIndex);
    if (newViewIndex == null) {
        newViewIndex = embeddedViews.length;
    }
    addToArray(embeddedViews, newViewIndex, view);
    // Note: Don't need to change projectedViews as the order in there
    // as always invalid...
    // Note: Don't need to change projectedViews as the order in there
    // as always invalid...
    types_1.Services.dirtyParentQueries(view);
    renderDetachView(view);
    var prevView = newViewIndex > 0 ? embeddedViews[newViewIndex - 1] : null;
    renderAttachEmbeddedView(elementData, prevView, view);
    return view;
}
exports.moveEmbeddedView = moveEmbeddedView;
function renderAttachEmbeddedView(elementData, prevView, view) {
    var prevRenderNode = prevView ? util_1.renderNode(prevView, (prevView.def.lastRenderRootNode)) :
        elementData.renderElement;
    var parentNode = view.renderer.parentNode(prevRenderNode);
    var nextSibling = view.renderer.nextSibling(prevRenderNode);
    // Note: We can't check if `nextSibling` is present, as on WebWorkers it will always be!
    // However, browsers automatically do `appendChild` when there is no `nextSibling`.
    // Note: We can't check if `nextSibling` is present, as on WebWorkers it will always be!
    // However, browsers automatically do `appendChild` when there is no `nextSibling`.
    util_1.visitRootRenderNodes(view, 2 /* InsertBefore */, parentNode, nextSibling, undefined);
}
function renderDetachView(view) {
    util_1.visitRootRenderNodes(view, 3 /* RemoveChild */, null, null, undefined);
}
exports.renderDetachView = renderDetachView;
function addToArray(arr, index, value) {
    // perf: array.push is faster than array.splice!
    if (index >= arr.length) {
        arr.push(value);
    }
    else {
        arr.splice(index, 0, value);
    }
}
function removeFromArray(arr, index) {
    // perf: array.pop is faster than array.splice!
    if (index >= arr.length - 1) {
        arr.pop();
    }
    else {
        arr.splice(index, 1);
    }
}
//# sourceMappingURL=view_attach.js.map