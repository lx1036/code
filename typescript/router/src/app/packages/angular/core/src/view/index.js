"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
function __export(m) {
    for (var p in m) if (!exports.hasOwnProperty(p)) exports[p] = m[p];
}
exports.__esModule = true;
var element_1 = require("./element");
exports.anchorDef = element_1.anchorDef;
exports.elementDef = element_1.elementDef;
var entrypoint_1 = require("./entrypoint");
exports.clearOverrides = entrypoint_1.clearOverrides;
exports.createNgModuleFactory = entrypoint_1.createNgModuleFactory;
exports.overrideComponentView = entrypoint_1.overrideComponentView;
exports.overrideProvider = entrypoint_1.overrideProvider;
var ng_content_1 = require("./ng_content");
exports.ngContentDef = ng_content_1.ngContentDef;
var ng_module_1 = require("./ng_module");
exports.moduleDef = ng_module_1.moduleDef;
exports.moduleProvideDef = ng_module_1.moduleProvideDef;
var provider_1 = require("./provider");
exports.directiveDef = provider_1.directiveDef;
exports.pipeDef = provider_1.pipeDef;
exports.providerDef = provider_1.providerDef;
var pure_expression_1 = require("./pure_expression");
exports.pureArrayDef = pure_expression_1.pureArrayDef;
exports.pureObjectDef = pure_expression_1.pureObjectDef;
exports.purePipeDef = pure_expression_1.purePipeDef;
var query_1 = require("./query");
exports.queryDef = query_1.queryDef;
var refs_1 = require("./refs");
exports.ViewRef_ = refs_1.ViewRef_;
exports.createComponentFactory = refs_1.createComponentFactory;
exports.getComponentViewDefinitionFactory = refs_1.getComponentViewDefinitionFactory;
exports.nodeValue = refs_1.nodeValue;
var services_1 = require("./services");
exports.initServicesIfNeeded = services_1.initServicesIfNeeded;
var text_1 = require("./text");
exports.textDef = text_1.textDef;
var util_1 = require("./util");
exports.EMPTY_ARRAY = util_1.EMPTY_ARRAY;
exports.EMPTY_MAP = util_1.EMPTY_MAP;
exports.createRendererType2 = util_1.createRendererType2;
exports.elementEventFullName = util_1.elementEventFullName;
exports.inlineInterpolate = util_1.inlineInterpolate;
exports.interpolate = util_1.interpolate;
exports.rootRenderNodes = util_1.rootRenderNodes;
exports.tokenKey = util_1.tokenKey;
exports.unwrapValue = util_1.unwrapValue;
var view_1 = require("./view");
exports.viewDef = view_1.viewDef;
var view_attach_1 = require("./view_attach");
exports.attachEmbeddedView = view_attach_1.attachEmbeddedView;
exports.detachEmbeddedView = view_attach_1.detachEmbeddedView;
exports.moveEmbeddedView = view_attach_1.moveEmbeddedView;
__export(require("./types"));
