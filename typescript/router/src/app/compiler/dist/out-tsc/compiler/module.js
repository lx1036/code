"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
require("../packages/angular/goog");
require("hammerjs");
var platform_browser_dynamic_1 = require("@angular/platform-browser-dynamic");
/**
 * https://www.zhihu.com/question/58083132/answer/155731764
 * Angular under the water
 *
 * yarn ngc -p src/app/compiler/tsconfig.json
 */
var AppComp = /** @class */ (function () {
    function AppComp() {
        this.name = 'lx1036';
    }
    AppComp.decorators = [
        { type: core_1.Component, args: [{
                    selector: 'app',
                    template: "\n    <p>{{name}}</p>\n  "
                },] },
    ];
    return AppComp;
}());
exports.AppComp = AppComp;
var AppModule = /** @class */ (function () {
    function AppModule() {
    }
    AppModule.decorators = [
        { type: core_1.NgModule, args: [{
                    declarations: [AppComp]
                },] },
    ];
    return AppModule;
}());
exports.AppModule = AppModule;
platform_browser_dynamic_1.platformBrowserDynamic().bootstrapModule(AppModule).then(function (ngModuleRef) { return console.log(ngModuleRef); });
//# sourceMappingURL=module.js.map