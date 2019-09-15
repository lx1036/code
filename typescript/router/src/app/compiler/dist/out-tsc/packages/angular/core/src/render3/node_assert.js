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
function assertNodeType(node, type) {
    assert_1.assertNotNull(node, 'should be called with a node');
    assert_1.assertEqual(node.type, type, "should be a " + typeName(type));
}
exports.assertNodeType = assertNodeType;
function assertNodeOfPossibleTypes(node) {
    var types = [];
    for (var _i = 1; _i < arguments.length; _i++) {
        types[_i - 1] = arguments[_i];
    }
    assert_1.assertNotNull(node, 'should be called with a node');
    var found = types.some(function (type) { return node.type === type; });
    assert_1.assertEqual(found, true, "Should be one of " + types.map(typeName).join(', '));
}
exports.assertNodeOfPossibleTypes = assertNodeOfPossibleTypes;
function typeName(type) {
    if (type == 1 /* Projection */)
        return 'Projection';
    if (type == 0 /* Container */)
        return 'Container';
    if (type == 2 /* View */)
        return 'View';
    if (type == 3 /* Element */)
        return 'Element';
    return '<unknown>';
}
//# sourceMappingURL=node_assert.js.map