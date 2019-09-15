"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var util_1 = require("../util");
var forward_ref_1 = require("./forward_ref");
var injection_token_1 = require("./injection_token");
var injector_1 = require("./injector");
var scope_1 = require("./scope");
/**
 * Marker which indicates that a value has not yet been created from the factory function.
 */
var NOT_YET = {};
/**
 * Marker which indicates that the factory function for a token is in the process of being called.
 *
 * If the injector is asked to inject a token with its value set to CIRCULAR, that indicates
 * injection of a dependency has recursively attempted to inject the original token, and there is
 * a circular dependency among the providers.
 */
var CIRCULAR = {};
var EMPTY_ARRAY = [];
/**
 * A lazily initialized NullInjector.
 */
var NULL_INJECTOR = undefined;
function getNullInjector() {
    if (NULL_INJECTOR === undefined) {
        NULL_INJECTOR = new injector_1.NullInjector();
    }
    return NULL_INJECTOR;
}
/**
 * Create a new `Injector` which is configured using `InjectorType`s.
 *
 * @experimental
 */
function createInjector(defType, parent) {
    if (parent === void 0) { parent = null; }
    parent = parent || getNullInjector();
    return new R3Injector(defType, parent);
}
exports.createInjector = createInjector;
var R3Injector = /** @class */ (function () {
    function R3Injector(def, parent) {
        var _this = this;
        this.parent = parent;
        /**
           * Map of tokens to records which contain the instances of those tokens.
           */
        this.records = new Map();
        /**
           * The transitive set of `InjectorType`s which define this injector.
           */
        this.injectorDefTypes = new Set();
        /**
           * Set of values instantiated by this injector which contain `ngOnDestroy` lifecycle hooks.
           */
        this.onDestroy = new Set();
        /**
           * Flag indicating that this injector was previously destroyed.
           */
        this.destroyed = false;
        // Start off by creating Records for every provider declared in every InjectorType
        // included transitively in `def`.
        deepForEach([def], function (injectorDef) { return _this.processInjectorType(injectorDef, new Set()); });
        // Make sure the INJECTOR token provides this injector.
        this.records.set(injector_1.INJECTOR, makeRecord(undefined, this));
        // Detect whether this injector has the APP_ROOT_SCOPE token and thus should provide
        // any injectable scoped to APP_ROOT_SCOPE.
        this.isRootInjector = this.records.has(scope_1.APP_ROOT);
        // Eagerly instantiate the InjectorType classes themselves.
        this.injectorDefTypes.forEach(function (defType) { return _this.get(defType); });
    }
    /**
     * Destroy the injector and release references to every instance or provider associated with it.
     *
     * Also calls the `OnDestroy` lifecycle hooks of every instance that was created for which a
     * hook was found.
     */
    /**
       * Destroy the injector and release references to every instance or provider associated with it.
       *
       * Also calls the `OnDestroy` lifecycle hooks of every instance that was created for which a
       * hook was found.
       */
    R3Injector.prototype.destroy = /**
       * Destroy the injector and release references to every instance or provider associated with it.
       *
       * Also calls the `OnDestroy` lifecycle hooks of every instance that was created for which a
       * hook was found.
       */
    function () {
        this.assertNotDestroyed();
        // Set destroyed = true first, in case lifecycle hooks re-enter destroy().
        this.destroyed = true;
        try {
            // Call all the lifecycle hooks.
            this.onDestroy.forEach(function (service) { return service.ngOnDestroy(); });
        }
        finally {
            // Release all references.
            this.records.clear();
            this.onDestroy.clear();
            this.injectorDefTypes.clear();
        }
    };
    R3Injector.prototype.get = function (token, notFoundValue, flags) {
        if (notFoundValue === void 0) { notFoundValue = injector_1.THROW_IF_NOT_FOUND; }
        if (flags === void 0) { flags = 0 /* Default */; }
        this.assertNotDestroyed();
        // Set the injection context.
        var previousInjector = injector_1.setCurrentInjector(this);
        try {
            // Check for the SkipSelf flag.
            if (!(flags & 4 /* SkipSelf */)) {
                // SkipSelf isn't set, check if the record belongs to this injector.
                var record = this.records.get(token);
                if (record === undefined) {
                    // No record, but maybe the token is scoped to this injector. Look for an ngInjectableDef
                    // with a scope matching this injector.
                    var def = couldBeInjectableType(token) &&
                        token.ngInjectableDef ||
                        undefined;
                    if (def !== undefined && this.injectableDefInScope(def)) {
                        // Found an ngInjectableDef and it's scoped to this injector. Pretend as if it was here
                        // all along.
                        record = injectableDefRecord(token);
                        this.records.set(token, record);
                    }
                }
                // If a record was found, get the instance for it and return it.
                if (record !== undefined) {
                    return this.hydrate(token, record);
                }
            }
            // Select the next injector based on the Self flag - if self is set, the next injector is
            // the NullInjector, otherwise it's the parent.
            var next = !(flags & 2 /* Self */) ? this.parent : getNullInjector();
            return this.parent.get(token, notFoundValue);
        }
        finally {
            // Lastly, clean up the state by restoring the previous injector.
            // Lastly, clean up the state by restoring the previous injector.
            injector_1.setCurrentInjector(previousInjector);
        }
    };
    R3Injector.prototype.assertNotDestroyed = function () {
        if (this.destroyed) {
            throw new Error('Injector has already been destroyed.');
        }
    };
    /**
     * Add an `InjectorType` or `InjectorDefTypeWithProviders` and all of its transitive providers
     * to this injector.
     */
    /**
       * Add an `InjectorType` or `InjectorDefTypeWithProviders` and all of its transitive providers
       * to this injector.
       */
    R3Injector.prototype.processInjectorType = /**
       * Add an `InjectorType` or `InjectorDefTypeWithProviders` and all of its transitive providers
       * to this injector.
       */
    function (defOrWrappedDef, parents) {
        var _this = this;
        defOrWrappedDef = forward_ref_1.resolveForwardRef(defOrWrappedDef);
        // Either the defOrWrappedDef is an InjectorType (with ngInjectorDef) or an
        // InjectorDefTypeWithProviders (aka ModuleWithProviders). Detecting either is a megamorphic
        // read, so care is taken to only do the read once.
        // First attempt to read the ngInjectorDef.
        var def = defOrWrappedDef.ngInjectorDef;
        // If that's not present, then attempt to read ngModule from the InjectorDefTypeWithProviders.
        var ngModule = (def == null) && defOrWrappedDef.ngModule || undefined;
        // Determine the InjectorType. In the case where `defOrWrappedDef` is an `InjectorType`,
        // then this is easy. In the case of an InjectorDefTypeWithProviders, then the definition type
        // is the `ngModule`.
        var defType = (ngModule === undefined) ? defOrWrappedDef : ngModule;
        // If defOrWrappedType was an InjectorDefTypeWithProviders, then .providers may hold some
        // extra providers.
        var providers = (ngModule !== undefined) && defOrWrappedDef.providers ||
            EMPTY_ARRAY;
        // Finally, if defOrWrappedType was an `InjectorDefTypeWithProviders`, then the actual
        // `InjectorDef` is on its `ngModule`.
        if (ngModule !== undefined) {
            def = ngModule.ngInjectorDef;
        }
        // If no definition was found, throw.
        if (def == null) {
            throw new Error("Type " + util_1.stringify(defType) + " is missing an ngInjectorDef definition.");
        }
        // Check for circular dependencies.
        if (parents.has(defType)) {
            throw new Error("Circular dependency: type " + util_1.stringify(defType) + " ends up importing itself.");
        }
        // Track the InjectorType and add a provider for it.
        this.injectorDefTypes.add(defType);
        this.records.set(defType, makeRecord(def.factory));
        // Add providers in the same way that @NgModule resolution did:
        // First, include providers from any imports.
        if (def.imports != null) {
            // Before processing defType's imports, add it to the set of parents. This way, if it ends
            // up deeply importing itself, this can be detected.
            parents.add(defType);
            try {
                deepForEach(def.imports, function (imported) { return _this.processInjectorType(imported, parents); });
            }
            finally {
                // Remove it from the parents set when finished.
                parents.delete(defType);
            }
        }
        // Next, include providers listed on the definition itself.
        if (def.providers != null) {
            deepForEach(def.providers, function (provider) { return _this.processProvider(provider); });
        }
        // Finally, include providers from an InjectorDefTypeWithProviders if there was one.
        deepForEach(providers, function (provider) { return _this.processProvider(provider); });
    };
    /**
     * Process a `SingleProvider` and add it.
     */
    /**
       * Process a `SingleProvider` and add it.
       */
    R3Injector.prototype.processProvider = /**
       * Process a `SingleProvider` and add it.
       */
    function (provider) {
        // Determine the token from the provider. Either it's its own token, or has a {provide: ...}
        // property.
        provider = forward_ref_1.resolveForwardRef(provider);
        var token = isTypeProvider(provider) ? provider : forward_ref_1.resolveForwardRef(provider.provide);
        // Construct a `Record` for the provider.
        var record = providerToRecord(provider);
        if (!isTypeProvider(provider) && provider.multi === true) {
            // If the provider indicates that it's a multi-provider, process it specially.
            // First check whether it's been defined already.
            var multiRecord_1 = this.records.get(token);
            if (multiRecord_1) {
                // It has. Throw a nice error if
                if (multiRecord_1.multi === undefined) {
                    throw new Error("Mixed multi-provider for " + token + ".");
                }
            }
            else {
                token = provider;
                multiRecord_1 = makeRecord(undefined, NOT_YET, true);
                multiRecord_1.factory = function () { return injector_1.injectArgs((multiRecord_1.multi)); };
                this.records.set(token, multiRecord_1);
            }
            token = provider;
            multiRecord_1.multi.push(provider);
        }
        var existing = this.records.get(token);
        if (existing && existing.multi !== undefined) {
            throw new Error("Mixed multi-provider for " + token);
        }
        this.records.set(token, record);
    };
    R3Injector.prototype.hydrate = function (token, record) {
        if (record.value === CIRCULAR) {
            throw new Error("Circular dep for " + util_1.stringify(token));
        }
        else if (record.value === NOT_YET) {
            record.value = CIRCULAR;
            record.value = record.factory();
        }
        if (typeof record.value === 'object' && record.value && hasOnDestroy(record.value)) {
            this.onDestroy.add(record.value);
        }
        return record.value;
    };
    R3Injector.prototype.injectableDefInScope = function (def) {
        if (!def.providedIn) {
            return false;
        }
        else if (typeof def.providedIn === 'string') {
            return def.providedIn === 'any' || (def.providedIn === 'root' && this.isRootInjector);
        }
        else {
            return this.injectorDefTypes.has(def.providedIn);
        }
    };
    return R3Injector;
}());
exports.R3Injector = R3Injector;
function injectableDefRecord(token) {
    var def = token.ngInjectableDef;
    if (def === undefined) {
        throw new Error("Type " + util_1.stringify(token) + " is missing an ngInjectableDef definition.");
    }
    return makeRecord(def.factory);
}
function providerToRecord(provider) {
    var token = forward_ref_1.resolveForwardRef(provider);
    var value = NOT_YET;
    var factory = undefined;
    if (isTypeProvider(provider)) {
        return injectableDefRecord(provider);
    }
    else {
        token = forward_ref_1.resolveForwardRef(provider.provide);
        if (isValueProvider(provider)) {
            value = provider.useValue;
        }
        else if (isExistingProvider(provider)) {
            factory = function () { return injector_1.inject(provider.useExisting); };
        }
        else if (isFactoryProvider(provider)) {
            factory = function () { return provider.useFactory.apply(provider, injector_1.injectArgs(provider.deps || [])); };
        }
        else {
            var classRef_1 = provider.useClass || token;
            if (hasDeps(provider)) {
                factory = function () { return new ((classRef_1).bind.apply((classRef_1), [void 0].concat(injector_1.injectArgs(provider.deps))))(); };
            }
            else {
                return injectableDefRecord(classRef_1);
            }
        }
    }
    return makeRecord(factory, value);
}
function makeRecord(factory, value, multi) {
    if (value === void 0) { value = NOT_YET; }
    if (multi === void 0) { multi = false; }
    return {
        factory: factory,
        value: value,
        multi: multi ? [] : undefined,
    };
}
function deepForEach(input, fn) {
    input.forEach(function (value) { return Array.isArray(value) ? deepForEach(value, fn) : fn(value); });
}
function isValueProvider(value) {
    return injector_1.USE_VALUE in value;
}
function isExistingProvider(value) {
    return !!value.useExisting;
}
function isFactoryProvider(value) {
    return !!value.useFactory;
}
function isClassProvider(value) {
    return !!value.useClass;
}
function isTypeProvider(value) {
    return typeof value === 'function';
}
function hasDeps(value) {
    return !!value.deps;
}
function hasOnDestroy(value) {
    return typeof value === 'object' && value != null && value.ngOnDestroy &&
        typeof value.ngOnDestroy === 'function';
}
function couldBeInjectableType(value) {
    return (typeof value === 'function') ||
        (typeof value === 'object' && value instanceof injection_token_1.InjectionToken);
}
//# sourceMappingURL=r3_injector.js.map