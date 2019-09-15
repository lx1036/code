"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
exports.__esModule = true;
var forward_ref_1 = require("../di/forward_ref");
var injector_1 = require("../di/injector");
var scope_1 = require("../di/scope");
var ng_module_factory_1 = require("../linker/ng_module_factory");
var util_1 = require("../util");
var util_2 = require("./util");
var UNDEFINED_VALUE = new Object();
var InjectorRefTokenKey = util_2.tokenKey(injector_1.Injector);
var INJECTORRefTokenKey = util_2.tokenKey(injector_1.INJECTOR);
var NgModuleRefTokenKey = util_2.tokenKey(ng_module_factory_1.NgModuleRef);
function moduleProvideDef(flags, token, value, deps) {
    // Need to resolve forwardRefs as e.g. for `useValue` we
    // lowered the expression and then stopped evaluating it,
    // i.e. also didn't unwrap it.
    value = forward_ref_1.resolveForwardRef(value);
    var depDefs = util_2.splitDepsDsl(deps, util_1.stringify(token));
    return {
        // will bet set by the module definition
        index: -1,
        deps: depDefs, flags: flags, token: token, value: value
    };
}
exports.moduleProvideDef = moduleProvideDef;
function moduleDef(providers) {
    var providersByKey = {};
    var modules = [];
    var isRoot = false;
    for (var i = 0; i < providers.length; i++) {
        var provider = providers[i];
        if (provider.token === scope_1.APP_ROOT) {
            isRoot = true;
        }
        if (provider.flags & 1073741824 /* TypeNgModule */) {
            modules.push(provider.token);
        }
        provider.index = i;
        providersByKey[util_2.tokenKey(provider.token)] = provider;
    }
    return {
        // Will be filled later...
        factory: null,
        providersByKey: providersByKey,
        providers: providers,
        modules: modules,
        isRoot: isRoot
    };
}
exports.moduleDef = moduleDef;
function initNgModule(data) {
    var def = data._def;
    var providers = data._providers = new Array(def.providers.length);
    for (var i = 0; i < def.providers.length; i++) {
        var provDef = def.providers[i];
        if (!(provDef.flags & 4096 /* LazyProvider */)) {
            // Make sure the provider has not been already initialized outside this loop.
            if (providers[i] === undefined) {
                providers[i] = _createProviderInstance(data, provDef);
            }
        }
    }
}
exports.initNgModule = initNgModule;
function resolveNgModuleDep(data, depDef, notFoundValue) {
    if (notFoundValue === void 0) { notFoundValue = injector_1.Injector.THROW_IF_NOT_FOUND; }
    var former = injector_1.setCurrentInjector(data);
    try {
        if (depDef.flags & 8 /* Value */) {
            return depDef.token;
        }
        if (depDef.flags & 2 /* Optional */) {
            notFoundValue = null;
        }
        if (depDef.flags & 1 /* SkipSelf */) {
            return data._parent.get(depDef.token, notFoundValue);
        }
        var tokenKey_1 = depDef.tokenKey;
        switch (tokenKey_1) {
            case InjectorRefTokenKey:
            case INJECTORRefTokenKey:
            case NgModuleRefTokenKey:
                return data;
        }
        var providerDef = data._def.providersByKey[tokenKey_1];
        if (providerDef) {
            var providerInstance = data._providers[providerDef.index];
            if (providerInstance === undefined) {
                providerInstance = data._providers[providerDef.index] =
                    _createProviderInstance(data, providerDef);
            }
            return providerInstance === UNDEFINED_VALUE ? undefined : providerInstance;
        }
        else if (depDef.token.ngInjectableDef && targetsModule(data, depDef.token.ngInjectableDef)) {
            var injectableDef = depDef.token.ngInjectableDef;
            var key = tokenKey_1;
            var index = data._providers.length;
            data._def.providersByKey[depDef.tokenKey] = {
                flags: 1024 /* TypeFactoryProvider */ | 4096 /* LazyProvider */,
                value: injectableDef.factory,
                deps: [], index: index,
                token: depDef.token
            };
            data._providers[index] = UNDEFINED_VALUE;
            return (data._providers[index] =
                _createProviderInstance(data, data._def.providersByKey[depDef.tokenKey]));
        }
        return data._parent.get(depDef.token, notFoundValue);
    }
    finally {
        injector_1.setCurrentInjector(former);
    }
}
exports.resolveNgModuleDep = resolveNgModuleDep;
function moduleTransitivelyPresent(ngModule, scope) {
    return ngModule._def.modules.indexOf(scope) > -1;
}
function targetsModule(ngModule, def) {
    return def.providedIn != null && (moduleTransitivelyPresent(ngModule, def.providedIn) ||
        def.providedIn === 'root' && ngModule._def.isRoot);
}
function _createProviderInstance(ngModule, providerDef) {
    var injectable;
    switch (providerDef.flags & 201347067 /* Types */) {
        case 512 /* TypeClassProvider */:
            injectable = _createClass(ngModule, providerDef.value, providerDef.deps);
            break;
        case 1024 /* TypeFactoryProvider */:
            injectable = _callFactory(ngModule, providerDef.value, providerDef.deps);
            break;
        case 2048 /* TypeUseExistingProvider */:
            injectable = resolveNgModuleDep(ngModule, providerDef.deps[0]);
            break;
        case 256 /* TypeValueProvider */:
            injectable = providerDef.value;
            break;
    }
    // The read of `ngOnDestroy` here is slightly expensive as it's megamorphic, so it should be
    // avoided if possible. The sequence of checks here determines whether ngOnDestroy needs to be
    // checked. It might not if the `injectable` isn't an object or if NodeFlags.OnDestroy is already
    // set (ngOnDestroy was detected statically).
    if (injectable !== UNDEFINED_VALUE && injectable != null && typeof injectable === 'object' &&
        !(providerDef.flags & 131072 /* OnDestroy */) && typeof injectable.ngOnDestroy === 'function') {
        providerDef.flags |= 131072 /* OnDestroy */;
    }
    return injectable === undefined ? UNDEFINED_VALUE : injectable;
}
function _createClass(ngModule, ctor, deps) {
    var len = deps.length;
    switch (len) {
        case 0:
            return new ctor();
        case 1:
            return new ctor(resolveNgModuleDep(ngModule, deps[0]));
        case 2:
            return new ctor(resolveNgModuleDep(ngModule, deps[0]), resolveNgModuleDep(ngModule, deps[1]));
        case 3:
            return new ctor(resolveNgModuleDep(ngModule, deps[0]), resolveNgModuleDep(ngModule, deps[1]), resolveNgModuleDep(ngModule, deps[2]));
        default:
            var depValues = new Array(len);
            for (var i = 0; i < len; i++) {
                depValues[i] = resolveNgModuleDep(ngModule, deps[i]);
            }
            return new (ctor.bind.apply(ctor, [void 0].concat(depValues)))();
    }
}
function _callFactory(ngModule, factory, deps) {
    var len = deps.length;
    switch (len) {
        case 0:
            return factory();
        case 1:
            return factory(resolveNgModuleDep(ngModule, deps[0]));
        case 2:
            return factory(resolveNgModuleDep(ngModule, deps[0]), resolveNgModuleDep(ngModule, deps[1]));
        case 3:
            return factory(resolveNgModuleDep(ngModule, deps[0]), resolveNgModuleDep(ngModule, deps[1]), resolveNgModuleDep(ngModule, deps[2]));
        default:
            var depValues = Array(len);
            for (var i = 0; i < len; i++) {
                depValues[i] = resolveNgModuleDep(ngModule, deps[i]);
            }
            return factory.apply(void 0, depValues);
    }
}
function callNgModuleLifecycle(ngModule, lifecycles) {
    var def = ngModule._def;
    var destroyed = new Set();
    for (var i = 0; i < def.providers.length; i++) {
        var provDef = def.providers[i];
        if (provDef.flags & 131072 /* OnDestroy */) {
            var instance = ngModule._providers[i];
            if (instance && instance !== UNDEFINED_VALUE) {
                var onDestroy = instance.ngOnDestroy;
                if (typeof onDestroy === 'function' && !destroyed.has(instance)) {
                    onDestroy.apply(instance);
                    destroyed.add(instance);
                }
            }
        }
    }
}
exports.callNgModuleLifecycle = callNgModuleLifecycle;
