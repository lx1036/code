"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
exports.__esModule = true;
var reflection_capabilities_1 = require("../reflection/reflection_capabilities");
var decorators_1 = require("../util/decorators");
var property_1 = require("../util/property");
var defs_1 = require("./defs");
var injector_1 = require("./injector");
var GET_PROPERTY_NAME = {};
var USE_VALUE = property_1.getClosureSafeProperty({ provide: String, useValue: GET_PROPERTY_NAME }, GET_PROPERTY_NAME);
var EMPTY_ARRAY = [];
function convertInjectableProviderToFactory(type, provider) {
    if (!provider) {
        var reflectionCapabilities = new reflection_capabilities_1.ReflectionCapabilities();
        var deps_1 = reflectionCapabilities.parameters(type);
        // TODO - convert to flags.
        return function () { return new (type.bind.apply(type, [void 0].concat(injector_1.injectArgs(deps_1))))(); };
    }
    if (USE_VALUE in provider) {
        var valueProvider_1 = provider;
        return function () { return valueProvider_1.useValue; };
    }
    else if (provider.useExisting) {
        var existingProvider_1 = provider;
        return function () { return injector_1.inject(existingProvider_1.useExisting); };
    }
    else if (provider.useFactory) {
        var factoryProvider_1 = provider;
        return function () { return factoryProvider_1.useFactory.apply(factoryProvider_1, injector_1.injectArgs(factoryProvider_1.deps || EMPTY_ARRAY)); };
    }
    else if (provider.useClass) {
        var classProvider_1 = provider;
        var deps_2 = provider.deps;
        if (!deps_2) {
            var reflectionCapabilities = new reflection_capabilities_1.ReflectionCapabilities();
            deps_2 = reflectionCapabilities.parameters(type);
        }
        return function () {
            return new ((_a = classProvider_1.useClass).bind.apply(_a, [void 0].concat(injector_1.injectArgs(deps_2))))();
            var _a;
        };
    }
    else {
        var deps_3 = provider.deps;
        if (!deps_3) {
            var reflectionCapabilities = new reflection_capabilities_1.ReflectionCapabilities();
            deps_3 = reflectionCapabilities.parameters(type);
        }
        return function () { return new (type.bind.apply(type, [void 0].concat(injector_1.injectArgs(deps_3))))(); };
    }
}
exports.convertInjectableProviderToFactory = convertInjectableProviderToFactory;
/**
* Injectable decorator and metadata.
*
*
* @Annotation
*/
exports.Injectable = decorators_1.makeDecorator('Injectable', undefined, undefined, undefined, function (injectableType, options) {
    if (options && options.providedIn !== undefined &&
        injectableType.ngInjectableDef === undefined) {
        injectableType.ngInjectableDef = defs_1.defineInjectable({
            providedIn: options.providedIn,
            factory: convertInjectableProviderToFactory(injectableType, options)
        });
    }
});
