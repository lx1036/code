"use strict";
var __assign = (this && this.__assign) || Object.assign || function(t) {
    for (var s, i = 1, n = arguments.length; i < n; i++) {
        s = arguments[i];
        for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p))
            t[p] = s[p];
    }
    return t;
};
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var rxjs_1 = require("rxjs");
var operators_1 = require("rxjs/operators");
var error_handler_1 = require("../src/error_handler");
var util_1 = require("../src/util");
var lang_1 = require("../src/util/lang");
var application_init_1 = require("./application_init");
var application_tokens_1 = require("./application_tokens");
var console_1 = require("./console");
var di_1 = require("./di");
var compiler_1 = require("./linker/compiler");
var component_factory_1 = require("./linker/component_factory");
var component_factory_resolver_1 = require("./linker/component_factory_resolver");
var ng_module_factory_1 = require("./linker/ng_module_factory");
var profile_1 = require("./profile/profile");
var testability_1 = require("./testability/testability");
var ng_zone_1 = require("./zone/ng_zone");
var _devMode = true;
var _runModeLocked = false;
var _platform;
exports.ALLOW_MULTIPLE_PLATFORMS = new di_1.InjectionToken('AllowMultipleToken');
/**
 * Disable Angular's development mode, which turns off assertions and other
 * checks within the framework.
 *
 * One important assertion this disables verifies that a change detection pass
 * does not result in additional changes to any bindings (also known as
 * unidirectional data flow).
 *
 *
 */
function enableProdMode() {
    if (_runModeLocked) {
        throw new Error('Cannot enable prod mode after platform setup.');
    }
    _devMode = false;
}
exports.enableProdMode = enableProdMode;
/**
 * Returns whether Angular is in development mode. After called once,
 * the value is locked and won't change any more.
 *
 * By default, this is true, unless a user calls `enableProdMode` before calling this.
 *
 * @experimental APIs related to application bootstrap are currently under review.
 */
function isDevMode() {
    _runModeLocked = true;
    return _devMode;
}
exports.isDevMode = isDevMode;
/**
 * A token for third-party components that can register themselves with NgProbe.
 *
 * @experimental
 */
var /**
 * A token for third-party components that can register themselves with NgProbe.
 *
 * @experimental
 */
NgProbeToken = /** @class */ (function () {
    function NgProbeToken(name, token) {
        this.name = name;
        this.token = token;
    }
    return NgProbeToken;
}());
exports.NgProbeToken = NgProbeToken;
/**
 * Creates a platform.
 * Platforms have to be eagerly created via this function.
 *
 * @experimental APIs related to application bootstrap are currently under review.
 */
function createPlatform(injector) {
    if (_platform && !_platform.destroyed &&
        !_platform.injector.get(exports.ALLOW_MULTIPLE_PLATFORMS, false)) {
        throw new Error('There can be only one platform. Destroy the previous one to create a new one.');
    }
    _platform = injector.get(PlatformRef);
    var inits = injector.get(application_tokens_1.PLATFORM_INITIALIZER, null);
    if (inits)
        inits.forEach(function (init) { return init(); });
    return _platform;
}
exports.createPlatform = createPlatform;
/**
 * Creates a factory for a platform
 *
 * @experimental APIs related to application bootstrap are currently under review.
 */
function createPlatformFactory(parentPlatformFactory, name, providers) {
    if (providers === void 0) { providers = []; }
    var desc = "Platform: " + name;
    var marker = new di_1.InjectionToken(desc);
    return function (extraProviders) {
        if (extraProviders === void 0) { extraProviders = []; }
        var platform = getPlatform();
        if (!platform || platform.injector.get(exports.ALLOW_MULTIPLE_PLATFORMS, false)) {
            if (parentPlatformFactory) {
                parentPlatformFactory(providers.concat(extraProviders).concat({ provide: marker, useValue: true }));
            }
            else {
                var injectedProviders = providers.concat(extraProviders).concat({ provide: marker, useValue: true });
                createPlatform(di_1.Injector.create({ providers: injectedProviders, name: desc }));
            }
        }
        return assertPlatform(marker);
    };
}
exports.createPlatformFactory = createPlatformFactory;
/**
 * Checks that there currently is a platform which contains the given token as a provider.
 *
 * @experimental APIs related to application bootstrap are currently under review.
 */
function assertPlatform(requiredToken) {
    var platform = getPlatform();
    if (!platform) {
        throw new Error('No platform exists!');
    }
    if (!platform.injector.get(requiredToken, null)) {
        throw new Error('A platform with a different configuration has been created. Please destroy it first.');
    }
    return platform;
}
exports.assertPlatform = assertPlatform;
/**
 * Destroy the existing platform.
 *
 * @experimental APIs related to application bootstrap are currently under review.
 */
function destroyPlatform() {
    if (_platform && !_platform.destroyed) {
        _platform.destroy();
    }
}
exports.destroyPlatform = destroyPlatform;
/**
 * Returns the current platform.
 *
 * @experimental APIs related to application bootstrap are currently under review.
 */
function getPlatform() {
    return _platform && !_platform.destroyed ? _platform : null;
}
exports.getPlatform = getPlatform;
/**
 * The Angular platform is the entry point for Angular on a web page. Each page
 * has exactly one platform, and services (such as reflection) which are common
 * to every Angular application running on the page are bound in its scope.
 *
 * A page's platform is initialized implicitly when a platform is created via a platform factory
 * (e.g. {@link platformBrowser}), or explicitly by calling the {@link createPlatform} function.
 *
 *
 */
var PlatformRef = /** @class */ (function () {
    /** @internal */
    function PlatformRef(_injector) {
        this._injector = _injector;
        this._modules = [];
        this._destroyListeners = [];
        this._destroyed = false;
    }
    /**
     * Creates an instance of an `@NgModule` for the given platform
     * for offline compilation.
     *
     * ## Simple Example
     *
     * ```typescript
     * my_module.ts:
     *
     * @NgModule({
     *   imports: [BrowserModule]
     * })
     * class MyModule {}
     *
     * main.ts:
     * import {MyModuleNgFactory} from './my_module.ngfactory';
     * import {platformBrowser} from '@angular/platform-browser';
     *
     * let moduleRef = platformBrowser().bootstrapModuleFactory(MyModuleNgFactory);
     * ```
     *
     * @experimental APIs related to application bootstrap are currently under review.
     */
    /**
       * Creates an instance of an `@NgModule` for the given platform
       * for offline compilation.
       *
       * ## Simple Example
       *
       * ```typescript
       * my_module.ts:
       *
       * @NgModule({
       *   imports: [BrowserModule]
       * })
       * class MyModule {}
       *
       * main.ts:
       * import {MyModuleNgFactory} from './my_module.ngfactory';
       * import {platformBrowser} from '@angular/platform-browser';
       *
       * let moduleRef = platformBrowser().bootstrapModuleFactory(MyModuleNgFactory);
       * ```
       *
       * @experimental APIs related to application bootstrap are currently under review.
       */
    PlatformRef.prototype.bootstrapModuleFactory = /**
       * Creates an instance of an `@NgModule` for the given platform
       * for offline compilation.
       *
       * ## Simple Example
       *
       * ```typescript
       * my_module.ts:
       *
       * @NgModule({
       *   imports: [BrowserModule]
       * })
       * class MyModule {}
       *
       * main.ts:
       * import {MyModuleNgFactory} from './my_module.ngfactory';
       * import {platformBrowser} from '@angular/platform-browser';
       *
       * let moduleRef = platformBrowser().bootstrapModuleFactory(MyModuleNgFactory);
       * ```
       *
       * @experimental APIs related to application bootstrap are currently under review.
       */
    function (moduleFactory, options) {
        var _this = this;
        // Note: We need to create the NgZone _before_ we instantiate the module,
        // as instantiating the module creates some providers eagerly.
        // So we create a mini parent injector that just contains the new NgZone and
        // pass that as parent to the NgModuleFactory.
        var ngZoneOption = options ? options.ngZone : undefined;
        var ngZone = getNgZone(ngZoneOption);
        var providers = [{ provide: ng_zone_1.NgZone, useValue: ngZone }];
        // Attention: Don't use ApplicationRef.run here,
        // as we want to be sure that all possible constructor calls are inside `ngZone.run`!
        return ngZone.run(function () {
            var ngZoneInjector = di_1.Injector.create({ providers: providers, parent: _this.injector, name: moduleFactory.moduleType.name });
            var moduleRef = moduleFactory.create(ngZoneInjector);
            var exceptionHandler = moduleRef.injector.get(error_handler_1.ErrorHandler, null);
            if (!exceptionHandler) {
                throw new Error('No ErrorHandler. Is platform module (BrowserModule) included?');
            }
            moduleRef.onDestroy(function () { return remove(_this._modules, moduleRef); });
            ngZone.runOutsideAngular(function () {
                return ngZone.onError.subscribe({ next: function (error) { exceptionHandler.handleError(error); } });
            });
            return _callAndReportToErrorHandler(exceptionHandler, (ngZone), function () {
                var initStatus = moduleRef.injector.get(application_init_1.ApplicationInitStatus);
                initStatus.runInitializers();
                return initStatus.donePromise.then(function () {
                    _this._moduleDoBootstrap(moduleRef);
                    return moduleRef;
                });
            });
        });
    };
    /**
     * Creates an instance of an `@NgModule` for a given platform using the given runtime compiler.
     *
     * ## Simple Example
     *
     * ```typescript
     * @NgModule({
     *   imports: [BrowserModule]
     * })
     * class MyModule {}
     *
     * let moduleRef = platformBrowser().bootstrapModule(MyModule);
     * ```
     *
     */
    /**
       * Creates an instance of an `@NgModule` for a given platform using the given runtime compiler.
       *
       * ## Simple Example
       *
       * ```typescript
       * @NgModule({
       *   imports: [BrowserModule]
       * })
       * class MyModule {}
       *
       * let moduleRef = platformBrowser().bootstrapModule(MyModule);
       * ```
       *
       */
    PlatformRef.prototype.bootstrapModule = /**
       * Creates an instance of an `@NgModule` for a given platform using the given runtime compiler.
       *
       * ## Simple Example
       *
       * ```typescript
       * @NgModule({
       *   imports: [BrowserModule]
       * })
       * class MyModule {}
       *
       * let moduleRef = platformBrowser().bootstrapModule(MyModule);
       * ```
       *
       */
    function (moduleType, compilerOptions) {
        var _this = this;
        if (compilerOptions === void 0) { compilerOptions = []; }
        var compilerFactory = this.injector.get(compiler_1.CompilerFactory);
        var options = optionsReducer({}, compilerOptions);
        var compiler = compilerFactory.createCompiler([options]);
        return compiler.compileModuleAsync(moduleType)
            .then(function (moduleFactory) { return _this.bootstrapModuleFactory(moduleFactory, options); });
    };
    PlatformRef.prototype._moduleDoBootstrap = function (moduleRef) {
        var appRef = moduleRef.injector.get(ApplicationRef);
        if (moduleRef._bootstrapComponents.length > 0) {
            moduleRef._bootstrapComponents.forEach(function (f) { return appRef.bootstrap(f); });
        }
        else if (moduleRef.instance.ngDoBootstrap) {
            moduleRef.instance.ngDoBootstrap(appRef);
        }
        else {
            throw new Error("The module " + util_1.stringify(moduleRef.instance.constructor) + " was bootstrapped, but it does not declare \"@NgModule.bootstrap\" components nor a \"ngDoBootstrap\" method. " +
                "Please define one of these.");
        }
        this._modules.push(moduleRef);
    };
    /**
     * Register a listener to be called when the platform is disposed.
     */
    /**
       * Register a listener to be called when the platform is disposed.
       */
    PlatformRef.prototype.onDestroy = /**
       * Register a listener to be called when the platform is disposed.
       */
    function (callback) { this._destroyListeners.push(callback); };
    Object.defineProperty(PlatformRef.prototype, "injector", {
        /**
         * Retrieve the platform {@link Injector}, which is the parent injector for
         * every Angular application on the page and provides singleton providers.
         */
        get: /**
           * Retrieve the platform {@link Injector}, which is the parent injector for
           * every Angular application on the page and provides singleton providers.
           */
        function () { return this._injector; },
        enumerable: true,
        configurable: true
    });
    /**
     * Destroy the Angular platform and all Angular applications on the page.
     */
    /**
       * Destroy the Angular platform and all Angular applications on the page.
       */
    PlatformRef.prototype.destroy = /**
       * Destroy the Angular platform and all Angular applications on the page.
       */
    function () {
        if (this._destroyed) {
            throw new Error('The platform has already been destroyed!');
        }
        this._modules.slice().forEach(function (module) { return module.destroy(); });
        this._destroyListeners.forEach(function (listener) { return listener(); });
        this._destroyed = true;
    };
    Object.defineProperty(PlatformRef.prototype, "destroyed", {
        get: function () { return this._destroyed; },
        enumerable: true,
        configurable: true
    });
    PlatformRef.decorators = [
        { type: di_1.Injectable },
    ];
    /** @nocollapse */
    PlatformRef.ctorParameters = function () { return [
        { type: di_1.Injector, },
    ]; };
    return PlatformRef;
}());
exports.PlatformRef = PlatformRef;
function getNgZone(ngZoneOption) {
    var ngZone;
    if (ngZoneOption === 'noop') {
        ngZone = new ng_zone_1.NoopNgZone();
    }
    else {
        ngZone = (ngZoneOption === 'zone.js' ? undefined : ngZoneOption) ||
            new ng_zone_1.NgZone({ enableLongStackTrace: isDevMode() });
    }
    return ngZone;
}
function _callAndReportToErrorHandler(errorHandler, ngZone, callback) {
    try {
        var result = callback();
        if (lang_1.isPromise(result)) {
            return result.catch(function (e) {
                ngZone.runOutsideAngular(function () { return errorHandler.handleError(e); });
                // rethrow as the exception handler might not do it
                throw e;
            });
        }
        return result;
    }
    catch (e) {
        ngZone.runOutsideAngular(function () { return errorHandler.handleError(e); });
        // rethrow as the exception handler might not do it
        throw e;
    }
}
function optionsReducer(dst, objs) {
    if (Array.isArray(objs)) {
        dst = objs.reduce(optionsReducer, dst);
    }
    else {
        dst = __assign({}, dst, objs);
    }
    return dst;
}
/**
 * A reference to an Angular application running on a page.
 *
 *
 */
var ApplicationRef = /** @class */ (function () {
    /** @internal */
    function ApplicationRef(_zone, _console, _injector, _exceptionHandler, _componentFactoryResolver, _initStatus) {
        var _this = this;
        this._zone = _zone;
        this._console = _console;
        this._injector = _injector;
        this._exceptionHandler = _exceptionHandler;
        this._componentFactoryResolver = _componentFactoryResolver;
        this._initStatus = _initStatus;
        this._bootstrapListeners = [];
        this._views = [];
        this._runningTick = false;
        this._enforceNoNewChanges = false;
        this._stable = true;
        /**
           * Get a list of component types registered to this application.
           * This list is populated even before the component is created.
           */
        this.componentTypes = [];
        /**
           * Get a list of components registered to this application.
           */
        this.components = [];
        this._enforceNoNewChanges = isDevMode();
        this._zone.onMicrotaskEmpty.subscribe({ next: function () { _this._zone.run(function () { _this.tick(); }); } });
        var isCurrentlyStable = new rxjs_1.Observable(function (observer) {
            _this._stable = _this._zone.isStable && !_this._zone.hasPendingMacrotasks &&
                !_this._zone.hasPendingMicrotasks;
            _this._zone.runOutsideAngular(function () {
                observer.next(_this._stable);
                observer.complete();
            });
        });
        var isStable = new rxjs_1.Observable(function (observer) {
            // Create the subscription to onStable outside the Angular Zone so that
            // the callback is run outside the Angular Zone.
            var stableSub;
            _this._zone.runOutsideAngular(function () {
                stableSub = _this._zone.onStable.subscribe(function () {
                    ng_zone_1.NgZone.assertNotInAngularZone();
                    // Check whether there are no pending macro/micro tasks in the next tick
                    // to allow for NgZone to update the state.
                    // Check whether there are no pending macro/micro tasks in the next tick
                    // to allow for NgZone to update the state.
                    util_1.scheduleMicroTask(function () {
                        if (!_this._stable && !_this._zone.hasPendingMacrotasks &&
                            !_this._zone.hasPendingMicrotasks) {
                            _this._stable = true;
                            observer.next(true);
                        }
                    });
                });
            });
            var unstableSub = _this._zone.onUnstable.subscribe(function () {
                ng_zone_1.NgZone.assertInAngularZone();
                if (_this._stable) {
                    _this._stable = false;
                    _this._zone.runOutsideAngular(function () { observer.next(false); });
                }
            });
            return function () {
                stableSub.unsubscribe();
                unstableSub.unsubscribe();
            };
        });
        this.isStable =
            rxjs_1.merge(isCurrentlyStable, isStable.pipe(operators_1.share()));
    }
    /**
     * Bootstrap a new component at the root level of the application.
     *
     * ### Bootstrap process
     *
     * When bootstrapping a new root component into an application, Angular mounts the
     * specified application component onto DOM elements identified by the [componentType]'s
     * selector and kicks off automatic change detection to finish initializing the component.
     *
     * Optionally, a component can be mounted onto a DOM element that does not match the
     * [componentType]'s selector.
     *
     * ### Example
     * {@example core/ts/platform/platform.ts region='longform'}
     */
    /**
       * Bootstrap a new component at the root level of the application.
       *
       * ### Bootstrap process
       *
       * When bootstrapping a new root component into an application, Angular mounts the
       * specified application component onto DOM elements identified by the [componentType]'s
       * selector and kicks off automatic change detection to finish initializing the component.
       *
       * Optionally, a component can be mounted onto a DOM element that does not match the
       * [componentType]'s selector.
       *
       * ### Example
       * {@example core/ts/platform/platform.ts region='longform'}
       */
    ApplicationRef.prototype.bootstrap = /**
       * Bootstrap a new component at the root level of the application.
       *
       * ### Bootstrap process
       *
       * When bootstrapping a new root component into an application, Angular mounts the
       * specified application component onto DOM elements identified by the [componentType]'s
       * selector and kicks off automatic change detection to finish initializing the component.
       *
       * Optionally, a component can be mounted onto a DOM element that does not match the
       * [componentType]'s selector.
       *
       * ### Example
       * {@example core/ts/platform/platform.ts region='longform'}
       */
    function (componentOrFactory, rootSelectorOrNode) {
        var _this = this;
        if (!this._initStatus.done) {
            throw new Error('Cannot bootstrap as there are still asynchronous initializers running. Bootstrap components in the `ngDoBootstrap` method of the root module.');
        }
        var componentFactory;
        if (componentOrFactory instanceof component_factory_1.ComponentFactory) {
            componentFactory = componentOrFactory;
        }
        else {
            componentFactory =
                (this._componentFactoryResolver.resolveComponentFactory(componentOrFactory));
        }
        this.componentTypes.push(componentFactory.componentType);
        // Create a factory associated with the current module if it's not bound to some other
        var ngModule = componentFactory instanceof component_factory_resolver_1.ComponentFactoryBoundToModule ?
            null :
            this._injector.get(ng_module_factory_1.NgModuleRef);
        var selectorOrNode = rootSelectorOrNode || componentFactory.selector;
        var compRef = componentFactory.create(di_1.Injector.NULL, [], selectorOrNode, ngModule);
        compRef.onDestroy(function () { _this._unloadComponent(compRef); });
        var testability = compRef.injector.get(testability_1.Testability, null);
        if (testability) {
            compRef.injector.get(testability_1.TestabilityRegistry)
                .registerApplication(compRef.location.nativeElement, testability);
        }
        this._loadComponent(compRef);
        if (isDevMode()) {
            this._console.log("Angular is running in the development mode. Call enableProdMode() to enable the production mode.");
        }
        return compRef;
    };
    /**
     * Invoke this method to explicitly process change detection and its side-effects.
     *
     * In development mode, `tick()` also performs a second change detection cycle to ensure that no
     * further changes are detected. If additional changes are picked up during this second cycle,
     * bindings in the app have side-effects that cannot be resolved in a single change detection
     * pass.
     * In this case, Angular throws an error, since an Angular application can only have one change
     * detection pass during which all change detection must complete.
     */
    /**
       * Invoke this method to explicitly process change detection and its side-effects.
       *
       * In development mode, `tick()` also performs a second change detection cycle to ensure that no
       * further changes are detected. If additional changes are picked up during this second cycle,
       * bindings in the app have side-effects that cannot be resolved in a single change detection
       * pass.
       * In this case, Angular throws an error, since an Angular application can only have one change
       * detection pass during which all change detection must complete.
       */
    ApplicationRef.prototype.tick = /**
       * Invoke this method to explicitly process change detection and its side-effects.
       *
       * In development mode, `tick()` also performs a second change detection cycle to ensure that no
       * further changes are detected. If additional changes are picked up during this second cycle,
       * bindings in the app have side-effects that cannot be resolved in a single change detection
       * pass.
       * In this case, Angular throws an error, since an Angular application can only have one change
       * detection pass during which all change detection must complete.
       */
    function () {
        var _this = this;
        if (this._runningTick) {
            throw new Error('ApplicationRef.tick is called recursively');
        }
        var scope = ApplicationRef._tickScope();
        try {
            this._runningTick = true;
            this._views.forEach(function (view) { return view.detectChanges(); });
            if (this._enforceNoNewChanges) {
                this._views.forEach(function (view) { return view.checkNoChanges(); });
            }
        }
        catch (e) {
            // Attention: Don't rethrow as it could cancel subscriptions to Observables!
            this._zone.runOutsideAngular(function () { return _this._exceptionHandler.handleError(e); });
        }
        finally {
            this._runningTick = false;
            profile_1.wtfLeave(scope);
        }
    };
    /**
     * Attaches a view so that it will be dirty checked.
     * The view will be automatically detached when it is destroyed.
     * This will throw if the view is already attached to a ViewContainer.
     */
    /**
       * Attaches a view so that it will be dirty checked.
       * The view will be automatically detached when it is destroyed.
       * This will throw if the view is already attached to a ViewContainer.
       */
    ApplicationRef.prototype.attachView = /**
       * Attaches a view so that it will be dirty checked.
       * The view will be automatically detached when it is destroyed.
       * This will throw if the view is already attached to a ViewContainer.
       */
    function (viewRef) {
        var view = viewRef;
        this._views.push(view);
        view.attachToAppRef(this);
    };
    /**
     * Detaches a view from dirty checking again.
     */
    /**
       * Detaches a view from dirty checking again.
       */
    ApplicationRef.prototype.detachView = /**
       * Detaches a view from dirty checking again.
       */
    function (viewRef) {
        var view = viewRef;
        remove(this._views, view);
        view.detachFromAppRef();
    };
    ApplicationRef.prototype._loadComponent = function (componentRef) {
        this.attachView(componentRef.hostView);
        this.tick();
        this.components.push(componentRef);
        // Get the listeners lazily to prevent DI cycles.
        var listeners = this._injector.get(application_tokens_1.APP_BOOTSTRAP_LISTENER, []).concat(this._bootstrapListeners);
        listeners.forEach(function (listener) { return listener(componentRef); });
    };
    ApplicationRef.prototype._unloadComponent = function (componentRef) {
        this.detachView(componentRef.hostView);
        remove(this.components, componentRef);
    };
    /** @internal */
    /** @internal */
    ApplicationRef.prototype.ngOnDestroy = /** @internal */
    function () {
        // TODO(alxhub): Dispose of the NgZone.
        this._views.slice().forEach(function (view) { return view.destroy(); });
    };
    Object.defineProperty(ApplicationRef.prototype, "viewCount", {
        /**
         * Returns the number of attached views.
         */
        get: /**
           * Returns the number of attached views.
           */
        function () { return this._views.length; },
        enumerable: true,
        configurable: true
    });
    /** @internal */
    ApplicationRef._tickScope = profile_1.wtfCreateScope('ApplicationRef#tick()');
    ApplicationRef.decorators = [
        { type: di_1.Injectable },
    ];
    /** @nocollapse */
    ApplicationRef.ctorParameters = function () { return [
        { type: ng_zone_1.NgZone, },
        { type: console_1.Console, },
        { type: di_1.Injector, },
        { type: error_handler_1.ErrorHandler, },
        { type: component_factory_resolver_1.ComponentFactoryResolver, },
        { type: application_init_1.ApplicationInitStatus, },
    ]; };
    return ApplicationRef;
}());
exports.ApplicationRef = ApplicationRef;
function remove(list, el) {
    var index = list.indexOf(el);
    if (index > -1) {
        list.splice(index, 1);
    }
}
//# sourceMappingURL=application_ref.js.map