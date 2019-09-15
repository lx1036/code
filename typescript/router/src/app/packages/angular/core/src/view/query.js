"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
exports.__esModule = true;
var element_ref_1 = require("../linker/element_ref");
var query_list_1 = require("../linker/query_list");
var types_1 = require("./types");
var util_1 = require("./util");
function queryDef(flags, id, bindings) {
    var bindingDefs = [];
    for (var propName in bindings) {
        var bindingType = bindings[propName];
        bindingDefs.push({ propName: propName, bindingType: bindingType });
    }
    return {
        // will bet set by the view definition
        nodeIndex: -1,
        parent: null,
        renderParent: null,
        bindingIndex: -1,
        outputIndex: -1,
        // regular values
        // TODO(vicb): check
        checkIndex: -1, flags: flags,
        childFlags: 0,
        directChildFlags: 0,
        childMatchedQueries: 0,
        ngContentIndex: -1,
        matchedQueries: {},
        matchedQueryIds: 0,
        references: {},
        childCount: 0,
        bindings: [],
        bindingFlags: 0,
        outputs: [],
        element: null,
        provider: null,
        text: null,
        query: { id: id, filterId: util_1.filterQueryId(id), bindings: bindingDefs },
        ngContent: null
    };
}
exports.queryDef = queryDef;
function createQuery() {
    return new query_list_1.QueryList();
}
exports.createQuery = createQuery;
function dirtyParentQueries(view) {
    var queryIds = view.def.nodeMatchedQueries;
    while (view.parent && util_1.isEmbeddedView(view)) {
        var tplDef = view.parentNodeDef;
        view = view.parent;
        // content queries
        var end = tplDef.nodeIndex + tplDef.childCount;
        for (var i = 0; i <= end; i++) {
            var nodeDef = view.def.nodes[i];
            if ((nodeDef.flags & 67108864 /* TypeContentQuery */) &&
                (nodeDef.flags & 536870912 /* DynamicQuery */) &&
                (nodeDef.query.filterId & queryIds) === nodeDef.query.filterId) {
                types_1.asQueryList(view, i).setDirty();
            }
            if ((nodeDef.flags & 1 /* TypeElement */ && i + nodeDef.childCount < tplDef.nodeIndex) ||
                !(nodeDef.childFlags & 67108864 /* TypeContentQuery */) ||
                !(nodeDef.childFlags & 536870912 /* DynamicQuery */)) {
                // skip elements that don't contain the template element or no query.
                i += nodeDef.childCount;
            }
        }
    }
    // view queries
    if (view.def.nodeFlags & 134217728 /* TypeViewQuery */) {
        for (var i = 0; i < view.def.nodes.length; i++) {
            var nodeDef = view.def.nodes[i];
            if ((nodeDef.flags & 134217728 /* TypeViewQuery */) && (nodeDef.flags & 536870912 /* DynamicQuery */)) {
                types_1.asQueryList(view, i).setDirty();
            }
            // only visit the root nodes
            i += nodeDef.childCount;
        }
    }
}
exports.dirtyParentQueries = dirtyParentQueries;
function checkAndUpdateQuery(view, nodeDef) {
    var queryList = types_1.asQueryList(view, nodeDef.nodeIndex);
    if (!queryList.dirty) {
        return;
    }
    var directiveInstance;
    var newValues = undefined;
    if (nodeDef.flags & 67108864 /* TypeContentQuery */) {
        var elementDef = nodeDef.parent.parent;
        newValues = calcQueryValues(view, elementDef.nodeIndex, elementDef.nodeIndex + elementDef.childCount, nodeDef.query, []);
        directiveInstance = types_1.asProviderData(view, nodeDef.parent.nodeIndex).instance;
    }
    else if (nodeDef.flags & 134217728 /* TypeViewQuery */) {
        newValues = calcQueryValues(view, 0, view.def.nodes.length - 1, nodeDef.query, []);
        directiveInstance = view.component;
    }
    queryList.reset(newValues);
    var bindings = nodeDef.query.bindings;
    var notify = false;
    for (var i = 0; i < bindings.length; i++) {
        var binding = bindings[i];
        var boundValue = void 0;
        switch (binding.bindingType) {
            case 0 /* First */:
                boundValue = queryList.first;
                break;
            case 1 /* All */:
                boundValue = queryList;
                notify = true;
                break;
        }
        directiveInstance[binding.propName] = boundValue;
    }
    if (notify) {
        queryList.notifyOnChanges();
    }
}
exports.checkAndUpdateQuery = checkAndUpdateQuery;
function calcQueryValues(view, startIndex, endIndex, queryDef, values) {
    for (var i = startIndex; i <= endIndex; i++) {
        var nodeDef = view.def.nodes[i];
        var valueType = nodeDef.matchedQueries[queryDef.id];
        if (valueType != null) {
            values.push(getQueryValue(view, nodeDef, valueType));
        }
        if (nodeDef.flags & 1 /* TypeElement */ && nodeDef.element.template &&
            (nodeDef.element.template.nodeMatchedQueries & queryDef.filterId) ===
                queryDef.filterId) {
            var elementData = types_1.asElementData(view, i);
            // check embedded views that were attached at the place of their template,
            // but process child nodes first if some match the query (see issue #16568)
            if ((nodeDef.childMatchedQueries & queryDef.filterId) === queryDef.filterId) {
                calcQueryValues(view, i + 1, i + nodeDef.childCount, queryDef, values);
                i += nodeDef.childCount;
            }
            if (nodeDef.flags & 16777216 /* EmbeddedViews */) {
                var embeddedViews = elementData.viewContainer._embeddedViews;
                for (var k = 0; k < embeddedViews.length; k++) {
                    var embeddedView = embeddedViews[k];
                    var dvc = util_1.declaredViewContainer(embeddedView);
                    if (dvc && dvc === elementData) {
                        calcQueryValues(embeddedView, 0, embeddedView.def.nodes.length - 1, queryDef, values);
                    }
                }
            }
            var projectedViews = elementData.template._projectedViews;
            if (projectedViews) {
                for (var k = 0; k < projectedViews.length; k++) {
                    var projectedView = projectedViews[k];
                    calcQueryValues(projectedView, 0, projectedView.def.nodes.length - 1, queryDef, values);
                }
            }
        }
        if ((nodeDef.childMatchedQueries & queryDef.filterId) !== queryDef.filterId) {
            // if no child matches the query, skip the children.
            i += nodeDef.childCount;
        }
    }
    return values;
}
function getQueryValue(view, nodeDef, queryValueType) {
    if (queryValueType != null) {
        // a match
        switch (queryValueType) {
            case 1 /* RenderElement */:
                return types_1.asElementData(view, nodeDef.nodeIndex).renderElement;
            case 0 /* ElementRef */:
                return new element_ref_1.ElementRef(types_1.asElementData(view, nodeDef.nodeIndex).renderElement);
            case 2 /* TemplateRef */:
                return types_1.asElementData(view, nodeDef.nodeIndex).template;
            case 3 /* ViewContainerRef */:
                return types_1.asElementData(view, nodeDef.nodeIndex).viewContainer;
            case 4 /* Provider */:
                return types_1.asProviderData(view, nodeDef.nodeIndex).instance;
        }
    }
}
exports.getQueryValue = getQueryValue;
