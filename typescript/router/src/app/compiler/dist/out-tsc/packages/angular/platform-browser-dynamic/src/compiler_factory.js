"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var compiler_1 = require("@angular/compiler");
var compiler_reflector_1 = require("./compiler_reflector");
exports.ERROR_COLLECTOR_TOKEN = new core_1.InjectionToken('ErrorCollector');
/**
 * A default provider for {@link PACKAGE_ROOT_URL} that maps to '/'.
 */
exports.DEFAULT_PACKAGE_URL_PROVIDER = {
    provide: core_1.PACKAGE_ROOT_URL,
    useValue: '/'
};
var _NO_RESOURCE_LOADER = {
    get: function (url) {
        throw new Error("No ResourceLoader implementation has been provided. Can't read the url \"" + url + "\"");
    }
};
var baseHtmlParser = new core_1.InjectionToken('HtmlParser');
var CompilerImpl = /** @class */ (function () {
    function CompilerImpl(injector, _metadataResolver, templateParser, styleCompiler, viewCompiler, ngModuleCompiler, summaryResolver, compileReflector, compilerConfig, console) {
        this._metadataResolver = _metadataResolver;
        this._delegate = new compiler_1.JitCompiler(_metadataResolver, templateParser, styleCompiler, viewCompiler, ngModuleCompiler, summaryResolver, compileReflector, compilerConfig, console, this.getExtraNgModuleProviders.bind(this));
        this.injector = injector;
    }
    CompilerImpl.prototype.getExtraNgModuleProviders = function () {
        return [this._metadataResolver.getProviderMetadata(new compiler_1.ProviderMeta(core_1.Compiler, { useValue: this }))];
    };
    CompilerImpl.prototype.compileModuleSync = function (moduleType) {
        return this._delegate.compileModuleSync(moduleType);
    };
    CompilerImpl.prototype.compileModuleAsync = function (moduleType) {
        return this._delegate.compileModuleAsync(moduleType);
    };
    CompilerImpl.prototype.compileModuleAndAllComponentsSync = function (moduleType) {
        var result = this._delegate.compileModuleAndAllComponentsSync(moduleType);
        return {
            ngModuleFactory: result.ngModuleFactory,
            componentFactories: result.componentFactories,
        };
    };
    CompilerImpl.prototype.compileModuleAndAllComponentsAsync = function (moduleType) {
        return this._delegate.compileModuleAndAllComponentsAsync(moduleType)
            .then(function (result) {
            return ({
                ngModuleFactory: result.ngModuleFactory,
                componentFactories: result.componentFactories,
            });
        });
    };
    CompilerImpl.prototype.loadAotSummaries = function (summaries) { this._delegate.loadAotSummaries(summaries); };
    CompilerImpl.prototype.hasAotSummary = function (ref) { return this._delegate.hasAotSummary(ref); };
    CompilerImpl.prototype.getComponentFactory = function (component) {
        return this._delegate.getComponentFactory(component);
    };
    CompilerImpl.prototype.clearCache = function () { this._delegate.clearCache(); };
    CompilerImpl.prototype.clearCacheFor = function (type) { this._delegate.clearCacheFor(type); };
    return CompilerImpl;
}());
exports.CompilerImpl = CompilerImpl;
/**
 * A set of providers that provide `JitCompiler` and its dependencies to use for
 * template compilation.
 */
exports.COMPILER_PROVIDERS = [
    { provide: compiler_1.CompileReflector, useValue: new compiler_reflector_1.JitReflector() },
    { provide: compiler_1.ResourceLoader, useValue: _NO_RESOURCE_LOADER },
    { provide: compiler_1.JitSummaryResolver, deps: [] },
    { provide: compiler_1.SummaryResolver, useExisting: compiler_1.JitSummaryResolver },
    { provide: core_1.ɵConsole, deps: [] },
    { provide: compiler_1.Lexer, deps: [] },
    { provide: compiler_1.Parser, deps: [compiler_1.Lexer] },
    {
        provide: baseHtmlParser,
        useClass: compiler_1.HtmlParser,
        deps: [],
    },
    {
        provide: compiler_1.I18NHtmlParser,
        useFactory: function (parser, translations, format, config, console) {
            translations = translations || '';
            var missingTranslation = translations ? config.missingTranslation : core_1.MissingTranslationStrategy.Ignore;
            return new compiler_1.I18NHtmlParser(parser, translations, format, missingTranslation, console);
        },
        deps: [
            baseHtmlParser,
            [new core_1.Optional(), new core_1.Inject(core_1.TRANSLATIONS)],
            [new core_1.Optional(), new core_1.Inject(core_1.TRANSLATIONS_FORMAT)],
            [compiler_1.CompilerConfig],
            [core_1.ɵConsole],
        ]
    },
    {
        provide: compiler_1.HtmlParser,
        useExisting: compiler_1.I18NHtmlParser,
    },
    {
        provide: compiler_1.TemplateParser, deps: [compiler_1.CompilerConfig, compiler_1.CompileReflector,
            compiler_1.Parser, compiler_1.ElementSchemaRegistry,
            compiler_1.I18NHtmlParser, core_1.ɵConsole]
    },
    { provide: compiler_1.DirectiveNormalizer, deps: [compiler_1.ResourceLoader, compiler_1.UrlResolver, compiler_1.HtmlParser, compiler_1.CompilerConfig] },
    { provide: compiler_1.CompileMetadataResolver, deps: [compiler_1.CompilerConfig, compiler_1.HtmlParser, compiler_1.NgModuleResolver,
            compiler_1.DirectiveResolver, compiler_1.PipeResolver,
            compiler_1.SummaryResolver,
            compiler_1.ElementSchemaRegistry,
            compiler_1.DirectiveNormalizer, core_1.ɵConsole,
            [core_1.Optional, compiler_1.StaticSymbolCache],
            compiler_1.CompileReflector,
            [core_1.Optional, exports.ERROR_COLLECTOR_TOKEN]] },
    exports.DEFAULT_PACKAGE_URL_PROVIDER,
    { provide: compiler_1.StyleCompiler, deps: [compiler_1.UrlResolver] },
    { provide: compiler_1.ViewCompiler, deps: [compiler_1.CompileReflector] },
    { provide: compiler_1.NgModuleCompiler, deps: [compiler_1.CompileReflector] },
    { provide: compiler_1.CompilerConfig, useValue: new compiler_1.CompilerConfig() },
    { provide: core_1.Compiler, useClass: CompilerImpl, deps: [core_1.Injector, compiler_1.CompileMetadataResolver,
            compiler_1.TemplateParser, compiler_1.StyleCompiler,
            compiler_1.ViewCompiler, compiler_1.NgModuleCompiler,
            compiler_1.SummaryResolver, compiler_1.CompileReflector, compiler_1.CompilerConfig,
            core_1.ɵConsole] },
    { provide: compiler_1.DomElementSchemaRegistry, deps: [] },
    { provide: compiler_1.ElementSchemaRegistry, useExisting: compiler_1.DomElementSchemaRegistry },
    { provide: compiler_1.UrlResolver, deps: [core_1.PACKAGE_ROOT_URL] },
    { provide: compiler_1.DirectiveResolver, deps: [compiler_1.CompileReflector] },
    { provide: compiler_1.PipeResolver, deps: [compiler_1.CompileReflector] },
    { provide: compiler_1.NgModuleResolver, deps: [compiler_1.CompileReflector] },
];
/**
 * @experimental
 */
var /**
 * @experimental
 */
JitCompilerFactory = /** @class */ (function () {
    /* @internal */
    function JitCompilerFactory(defaultOptions) {
        var compilerOptions = {
            useJit: true,
            defaultEncapsulation: core_1.ViewEncapsulation.Emulated,
            missingTranslation: core_1.MissingTranslationStrategy.Warning,
        };
        this._defaultOptions = [compilerOptions].concat(defaultOptions);
    }
    JitCompilerFactory.prototype.createCompiler = function (options) {
        if (options === void 0) { options = []; }
        var opts = _mergeOptions(this._defaultOptions.concat(options));
        var injector = core_1.Injector.create([
            exports.COMPILER_PROVIDERS, {
                provide: compiler_1.CompilerConfig,
                useFactory: function () {
                    return new compiler_1.CompilerConfig({
                        // let explicit values from the compiler options overwrite options
                        // from the app providers
                        useJit: opts.useJit,
                        jitDevMode: core_1.isDevMode(),
                        // let explicit values from the compiler options overwrite options
                        // from the app providers
                        defaultEncapsulation: opts.defaultEncapsulation,
                        missingTranslation: opts.missingTranslation,
                        preserveWhitespaces: opts.preserveWhitespaces,
                    });
                },
                deps: []
            },
            (opts.providers)
        ]);
        return injector.get(core_1.Compiler);
    };
    return JitCompilerFactory;
}());
exports.JitCompilerFactory = JitCompilerFactory;
function _mergeOptions(optionsArr) {
    return {
        useJit: _lastDefined(optionsArr.map(function (options) { return options.useJit; })),
        defaultEncapsulation: _lastDefined(optionsArr.map(function (options) { return options.defaultEncapsulation; })),
        providers: _mergeArrays(optionsArr.map(function (options) { return options.providers; })),
        missingTranslation: _lastDefined(optionsArr.map(function (options) { return options.missingTranslation; })),
        preserveWhitespaces: _lastDefined(optionsArr.map(function (options) { return options.preserveWhitespaces; })),
    };
}
function _lastDefined(args) {
    for (var i = args.length - 1; i >= 0; i--) {
        if (args[i] !== undefined) {
            return args[i];
        }
    }
    return undefined;
}
function _mergeArrays(parts) {
    var result = [];
    parts.forEach(function (part) { return part && result.push.apply(result, part); });
    return result;
}
//# sourceMappingURL=compiler_factory.js.map