"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = Object.setPrototypeOf ||
        ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
        function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var location_1 = require("./location");
var location_strategy_1 = require("./location_strategy");
var platform_location_1 = require("./platform_location");
/**
 * @description
 * A {@link LocationStrategy} used to configure the {@link Location} service to
 * represent its state in the
 * [hash fragment](https://en.wikipedia.org/wiki/Uniform_Resource_Locator#Syntax)
 * of the browser's URL.
 *
 * For instance, if you call `location.go('/foo')`, the browser's URL will become
 * `example.com#/foo`.
 *
 * ### Example
 *
 * {@example common/location/ts/hash_location_component.ts region='LocationComponent'}
 *
 *
 */
var HashLocationStrategy = /** @class */ (function (_super) {
    __extends(HashLocationStrategy, _super);
    function HashLocationStrategy(_platformLocation, _baseHref) {
        var _this = _super.call(this) || this;
        _this._platformLocation = _platformLocation;
        _this._baseHref = '';
        if (_baseHref != null) {
            _this._baseHref = _baseHref;
        }
        return _this;
    }
    HashLocationStrategy.prototype.onPopState = function (fn) {
        this._platformLocation.onPopState(fn);
        this._platformLocation.onHashChange(fn);
    };
    HashLocationStrategy.prototype.getBaseHref = function () { return this._baseHref; };
    HashLocationStrategy.prototype.path = function (includeHash) {
        if (includeHash === void 0) { includeHash = false; }
        // the hash value is always prefixed with a `#`
        // and if it is empty then it will stay empty
        var path = this._platformLocation.hash;
        if (path == null)
            path = '#';
        return path.length > 0 ? path.substring(1) : path;
    };
    HashLocationStrategy.prototype.prepareExternalUrl = function (internal) {
        var url = location_1.Location.joinWithSlash(this._baseHref, internal);
        return url.length > 0 ? ('#' + url) : url;
    };
    HashLocationStrategy.prototype.pushState = function (state, title, path, queryParams) {
        var url = this.prepareExternalUrl(path + location_1.Location.normalizeQueryParams(queryParams));
        if (url.length == 0) {
            url = this._platformLocation.pathname;
        }
        this._platformLocation.pushState(state, title, url);
    };
    HashLocationStrategy.prototype.replaceState = function (state, title, path, queryParams) {
        var url = this.prepareExternalUrl(path + location_1.Location.normalizeQueryParams(queryParams));
        if (url.length == 0) {
            url = this._platformLocation.pathname;
        }
        this._platformLocation.replaceState(state, title, url);
    };
    HashLocationStrategy.prototype.forward = function () { this._platformLocation.forward(); };
    HashLocationStrategy.prototype.back = function () { this._platformLocation.back(); };
    HashLocationStrategy.decorators = [
        { type: core_1.Injectable },
    ];
    /** @nocollapse */
    HashLocationStrategy.ctorParameters = function () { return [
        { type: platform_location_1.PlatformLocation, },
        { type: undefined, decorators: [{ type: core_1.Optional }, { type: core_1.Inject, args: [location_strategy_1.APP_BASE_HREF,] },] },
    ]; };
    return HashLocationStrategy;
}(location_strategy_1.LocationStrategy));
exports.HashLocationStrategy = HashLocationStrategy;
//# sourceMappingURL=hash_location_strategy.js.map