"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
exports.__esModule = true;
var types_1 = require("./types");
var util_1 = require("./util");
function textDef(checkIndex, ngContentIndex, staticText) {
    var bindings = new Array(staticText.length - 1);
    for (var i = 1; i < staticText.length; i++) {
        bindings[i - 1] = {
            flags: 8 /* TypeProperty */,
            name: null,
            ns: null,
            nonMinifiedName: null,
            securityContext: null,
            suffix: staticText[i]
        };
    }
    return {
        // will bet set by the view definition
        nodeIndex: -1,
        parent: null,
        renderParent: null,
        bindingIndex: -1,
        outputIndex: -1,
        // regular values
        checkIndex: checkIndex,
        flags: 2 /* TypeText */,
        childFlags: 0,
        directChildFlags: 0,
        childMatchedQueries: 0,
        matchedQueries: {},
        matchedQueryIds: 0,
        references: {}, ngContentIndex: ngContentIndex,
        childCount: 0, bindings: bindings,
        bindingFlags: 8 /* TypeProperty */,
        outputs: [],
        element: null,
        provider: null,
        text: { prefix: staticText[0] },
        query: null,
        ngContent: null
    };
}
exports.textDef = textDef;
function createText(view, renderHost, def) {
    var renderNode;
    var renderer = view.renderer;
    renderNode = renderer.createText(def.text.prefix);
    var parentEl = util_1.getParentRenderElement(view, renderHost, def);
    if (parentEl) {
        renderer.appendChild(parentEl, renderNode);
    }
    return { renderText: renderNode };
}
exports.createText = createText;
function checkAndUpdateTextInline(view, def, v0, v1, v2, v3, v4, v5, v6, v7, v8, v9) {
    var changed = false;
    var bindings = def.bindings;
    var bindLen = bindings.length;
    if (bindLen > 0 && util_1.checkAndUpdateBinding(view, def, 0, v0))
        changed = true;
    if (bindLen > 1 && util_1.checkAndUpdateBinding(view, def, 1, v1))
        changed = true;
    if (bindLen > 2 && util_1.checkAndUpdateBinding(view, def, 2, v2))
        changed = true;
    if (bindLen > 3 && util_1.checkAndUpdateBinding(view, def, 3, v3))
        changed = true;
    if (bindLen > 4 && util_1.checkAndUpdateBinding(view, def, 4, v4))
        changed = true;
    if (bindLen > 5 && util_1.checkAndUpdateBinding(view, def, 5, v5))
        changed = true;
    if (bindLen > 6 && util_1.checkAndUpdateBinding(view, def, 6, v6))
        changed = true;
    if (bindLen > 7 && util_1.checkAndUpdateBinding(view, def, 7, v7))
        changed = true;
    if (bindLen > 8 && util_1.checkAndUpdateBinding(view, def, 8, v8))
        changed = true;
    if (bindLen > 9 && util_1.checkAndUpdateBinding(view, def, 9, v9))
        changed = true;
    if (changed) {
        var value = def.text.prefix;
        if (bindLen > 0)
            value += _addInterpolationPart(v0, bindings[0]);
        if (bindLen > 1)
            value += _addInterpolationPart(v1, bindings[1]);
        if (bindLen > 2)
            value += _addInterpolationPart(v2, bindings[2]);
        if (bindLen > 3)
            value += _addInterpolationPart(v3, bindings[3]);
        if (bindLen > 4)
            value += _addInterpolationPart(v4, bindings[4]);
        if (bindLen > 5)
            value += _addInterpolationPart(v5, bindings[5]);
        if (bindLen > 6)
            value += _addInterpolationPart(v6, bindings[6]);
        if (bindLen > 7)
            value += _addInterpolationPart(v7, bindings[7]);
        if (bindLen > 8)
            value += _addInterpolationPart(v8, bindings[8]);
        if (bindLen > 9)
            value += _addInterpolationPart(v9, bindings[9]);
        var renderNode = types_1.asTextData(view, def.nodeIndex).renderText;
        view.renderer.setValue(renderNode, value);
    }
    return changed;
}
exports.checkAndUpdateTextInline = checkAndUpdateTextInline;
function checkAndUpdateTextDynamic(view, def, values) {
    var bindings = def.bindings;
    var changed = false;
    for (var i = 0; i < values.length; i++) {
        // Note: We need to loop over all values, so that
        // the old values are updates as well!
        if (util_1.checkAndUpdateBinding(view, def, i, values[i])) {
            changed = true;
        }
    }
    if (changed) {
        var value = '';
        for (var i = 0; i < values.length; i++) {
            value = value + _addInterpolationPart(values[i], bindings[i]);
        }
        value = def.text.prefix + value;
        var renderNode = types_1.asTextData(view, def.nodeIndex).renderText;
        view.renderer.setValue(renderNode, value);
    }
    return changed;
}
exports.checkAndUpdateTextDynamic = checkAndUpdateTextDynamic;
function _addInterpolationPart(value, binding) {
    var valueStr = value != null ? value.toString() : '';
    return valueStr + binding.suffix;
}
