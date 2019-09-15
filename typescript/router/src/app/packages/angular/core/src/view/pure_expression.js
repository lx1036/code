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
function purePipeDef(checkIndex, argCount) {
    // argCount + 1 to include the pipe as first arg
    return _pureExpressionDef(128 /* TypePurePipe */, checkIndex, new Array(argCount + 1));
}
exports.purePipeDef = purePipeDef;
function pureArrayDef(checkIndex, argCount) {
    return _pureExpressionDef(32 /* TypePureArray */, checkIndex, new Array(argCount));
}
exports.pureArrayDef = pureArrayDef;
function pureObjectDef(checkIndex, propToIndex) {
    var keys = Object.keys(propToIndex);
    var nbKeys = keys.length;
    var propertyNames = new Array(nbKeys);
    for (var i = 0; i < nbKeys; i++) {
        var key = keys[i];
        var index = propToIndex[key];
        propertyNames[index] = key;
    }
    return _pureExpressionDef(64 /* TypePureObject */, checkIndex, propertyNames);
}
exports.pureObjectDef = pureObjectDef;
function _pureExpressionDef(flags, checkIndex, propertyNames) {
    var bindings = new Array(propertyNames.length);
    for (var i = 0; i < propertyNames.length; i++) {
        var prop = propertyNames[i];
        bindings[i] = {
            flags: 8 /* TypeProperty */,
            name: prop,
            ns: null,
            nonMinifiedName: prop,
            securityContext: null,
            suffix: null
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
        flags: flags,
        childFlags: 0,
        directChildFlags: 0,
        childMatchedQueries: 0,
        matchedQueries: {},
        matchedQueryIds: 0,
        references: {},
        ngContentIndex: -1,
        childCount: 0, bindings: bindings,
        bindingFlags: util_1.calcBindingFlags(bindings),
        outputs: [],
        element: null,
        provider: null,
        text: null,
        query: null,
        ngContent: null
    };
}
function createPureExpression(view, def) {
    return { value: undefined };
}
exports.createPureExpression = createPureExpression;
function checkAndUpdatePureExpressionInline(view, def, v0, v1, v2, v3, v4, v5, v6, v7, v8, v9) {
    var bindings = def.bindings;
    var changed = false;
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
        var data = types_1.asPureExpressionData(view, def.nodeIndex);
        var value = void 0;
        switch (def.flags & 201347067 /* Types */) {
            case 32 /* TypePureArray */:
                value = new Array(bindings.length);
                if (bindLen > 0)
                    value[0] = v0;
                if (bindLen > 1)
                    value[1] = v1;
                if (bindLen > 2)
                    value[2] = v2;
                if (bindLen > 3)
                    value[3] = v3;
                if (bindLen > 4)
                    value[4] = v4;
                if (bindLen > 5)
                    value[5] = v5;
                if (bindLen > 6)
                    value[6] = v6;
                if (bindLen > 7)
                    value[7] = v7;
                if (bindLen > 8)
                    value[8] = v8;
                if (bindLen > 9)
                    value[9] = v9;
                break;
            case 64 /* TypePureObject */:
                value = {};
                if (bindLen > 0)
                    value[bindings[0].name] = v0;
                if (bindLen > 1)
                    value[bindings[1].name] = v1;
                if (bindLen > 2)
                    value[bindings[2].name] = v2;
                if (bindLen > 3)
                    value[bindings[3].name] = v3;
                if (bindLen > 4)
                    value[bindings[4].name] = v4;
                if (bindLen > 5)
                    value[bindings[5].name] = v5;
                if (bindLen > 6)
                    value[bindings[6].name] = v6;
                if (bindLen > 7)
                    value[bindings[7].name] = v7;
                if (bindLen > 8)
                    value[bindings[8].name] = v8;
                if (bindLen > 9)
                    value[bindings[9].name] = v9;
                break;
            case 128 /* TypePurePipe */:
                var pipe = v0;
                switch (bindLen) {
                    case 1:
                        value = pipe.transform(v0);
                        break;
                    case 2:
                        value = pipe.transform(v1);
                        break;
                    case 3:
                        value = pipe.transform(v1, v2);
                        break;
                    case 4:
                        value = pipe.transform(v1, v2, v3);
                        break;
                    case 5:
                        value = pipe.transform(v1, v2, v3, v4);
                        break;
                    case 6:
                        value = pipe.transform(v1, v2, v3, v4, v5);
                        break;
                    case 7:
                        value = pipe.transform(v1, v2, v3, v4, v5, v6);
                        break;
                    case 8:
                        value = pipe.transform(v1, v2, v3, v4, v5, v6, v7);
                        break;
                    case 9:
                        value = pipe.transform(v1, v2, v3, v4, v5, v6, v7, v8);
                        break;
                    case 10:
                        value = pipe.transform(v1, v2, v3, v4, v5, v6, v7, v8, v9);
                        break;
                }
                break;
        }
        data.value = value;
    }
    return changed;
}
exports.checkAndUpdatePureExpressionInline = checkAndUpdatePureExpressionInline;
function checkAndUpdatePureExpressionDynamic(view, def, values) {
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
        var data = types_1.asPureExpressionData(view, def.nodeIndex);
        var value = void 0;
        switch (def.flags & 201347067 /* Types */) {
            case 32 /* TypePureArray */:
                value = values;
                break;
            case 64 /* TypePureObject */:
                value = {};
                for (var i = 0; i < values.length; i++) {
                    value[bindings[i].name] = values[i];
                }
                break;
            case 128 /* TypePurePipe */:
                var pipe = values[0];
                var params = values.slice(1);
                value = pipe.transform.apply(pipe, params);
                break;
        }
        data.value = value;
    }
    return changed;
}
exports.checkAndUpdatePureExpressionDynamic = checkAndUpdatePureExpressionDynamic;
