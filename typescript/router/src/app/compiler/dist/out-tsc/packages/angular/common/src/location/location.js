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
var location_strategy_1 = require("./location_strategy");
/**
 * @description
 *
 * A service that applications can use to interact with a browser's URL.
 *
 * Depending on which {@link LocationStrategy} is used, `Location` will either persist
 * to the URL's path or the URL's hash segment.
 *
 * Note: it's better to use {@link Router#navigate} service to trigger route changes. Use
 * `Location` only if you need to interact with or create normalized URLs outside of
 * routing.
 *
 * `Location` is responsible for normalizing the URL against the application's base href.
 * A normalized URL is absolute from the URL host, includes the application's base href, and has no
 * trailing slash:
 * - `/my/app/user/123` is normalized
 * - `my/app/user/123` **is not** normalized
 * - `/my/app/user/123/` **is not** normalized
 *
 * ### Example
 * {@example common/location/ts/path_location_component.ts region='LocationComponent'}
 *
 */
var Location = /** @class */ (function () {
    function Location(platformStrategy) {
        var _this = this;
        /** @internal */
        this._subject = new core_1.EventEmitter();
        this._platformStrategy = platformStrategy;
        var browserBaseHref = this._platformStrategy.getBaseHref();
        this._baseHref = Location.stripTrailingSlash(_stripIndexHtml(browserBaseHref));
        this._platformStrategy.onPopState(function (ev) {
            _this._subject.emit({
                'url': _this.path(true),
                'pop': true,
                'state': ev.state,
                'type': ev.type,
            });
        });
    }
    /**
     * Returns the normalized URL path.
     */
    // TODO: vsavkin. Remove the boolean flag and always include hash once the deprecated router is
    // removed.
    /**
       * Returns the normalized URL path.
       */
    // TODO: vsavkin. Remove the boolean flag and always include hash once the deprecated router is
    // removed.
    Location.prototype.path = /**
       * Returns the normalized URL path.
       */
    // TODO: vsavkin. Remove the boolean flag and always include hash once the deprecated router is
    // removed.
    function (includeHash) {
        if (includeHash === void 0) { includeHash = false; }
        return this.normalize(this._platformStrategy.path(includeHash));
    };
    /**
     * Normalizes the given path and compares to the current normalized path.
     */
    /**
       * Normalizes the given path and compares to the current normalized path.
       */
    Location.prototype.isCurrentPathEqualTo = /**
       * Normalizes the given path and compares to the current normalized path.
       */
    function (path, query) {
        if (query === void 0) { query = ''; }
        return this.path() == this.normalize(path + Location.normalizeQueryParams(query));
    };
    /**
     * Given a string representing a URL, returns the normalized URL path without leading or
     * trailing slashes.
     */
    /**
       * Given a string representing a URL, returns the normalized URL path without leading or
       * trailing slashes.
       */
    Location.prototype.normalize = /**
       * Given a string representing a URL, returns the normalized URL path without leading or
       * trailing slashes.
       */
    function (url) {
        return Location.stripTrailingSlash(_stripBaseHref(this._baseHref, _stripIndexHtml(url)));
    };
    /**
     * Given a string representing a URL, returns the platform-specific external URL path.
     * If the given URL doesn't begin with a leading slash (`'/'`), this method adds one
     * before normalizing. This method will also add a hash if `HashLocationStrategy` is
     * used, or the `APP_BASE_HREF` if the `PathLocationStrategy` is in use.
     */
    /**
       * Given a string representing a URL, returns the platform-specific external URL path.
       * If the given URL doesn't begin with a leading slash (`'/'`), this method adds one
       * before normalizing. This method will also add a hash if `HashLocationStrategy` is
       * used, or the `APP_BASE_HREF` if the `PathLocationStrategy` is in use.
       */
    Location.prototype.prepareExternalUrl = /**
       * Given a string representing a URL, returns the platform-specific external URL path.
       * If the given URL doesn't begin with a leading slash (`'/'`), this method adds one
       * before normalizing. This method will also add a hash if `HashLocationStrategy` is
       * used, or the `APP_BASE_HREF` if the `PathLocationStrategy` is in use.
       */
    function (url) {
        if (url && url[0] !== '/') {
            url = '/' + url;
        }
        return this._platformStrategy.prepareExternalUrl(url);
    };
    // TODO: rename this method to pushState
    /**
     * Changes the browsers URL to the normalized version of the given URL, and pushes a
     * new item onto the platform's history.
     */
    // TODO: rename this method to pushState
    /**
       * Changes the browsers URL to the normalized version of the given URL, and pushes a
       * new item onto the platform's history.
       */
    Location.prototype.go = 
    // TODO: rename this method to pushState
    /**
       * Changes the browsers URL to the normalized version of the given URL, and pushes a
       * new item onto the platform's history.
       */
    function (path, query, state) {
        if (query === void 0) { query = ''; }
        if (state === void 0) { state = null; }
        this._platformStrategy.pushState(state, '', path, query);
    };
    /**
     * Changes the browsers URL to the normalized version of the given URL, and replaces
     * the top item on the platform's history stack.
     */
    /**
       * Changes the browsers URL to the normalized version of the given URL, and replaces
       * the top item on the platform's history stack.
       */
    Location.prototype.replaceState = /**
       * Changes the browsers URL to the normalized version of the given URL, and replaces
       * the top item on the platform's history stack.
       */
    function (path, query, state) {
        if (query === void 0) { query = ''; }
        if (state === void 0) { state = null; }
        this._platformStrategy.replaceState(state, '', path, query);
    };
    /**
     * Navigates forward in the platform's history.
     */
    /**
       * Navigates forward in the platform's history.
       */
    Location.prototype.forward = /**
       * Navigates forward in the platform's history.
       */
    function () { this._platformStrategy.forward(); };
    /**
     * Navigates back in the platform's history.
     */
    /**
       * Navigates back in the platform's history.
       */
    Location.prototype.back = /**
       * Navigates back in the platform's history.
       */
    function () { this._platformStrategy.back(); };
    /**
     * Subscribe to the platform's `popState` events.
     */
    /**
       * Subscribe to the platform's `popState` events.
       */
    Location.prototype.subscribe = /**
       * Subscribe to the platform's `popState` events.
       */
    function (onNext, onThrow, onReturn) {
        return this._subject.subscribe({ next: onNext, error: onThrow, complete: onReturn });
    };
    /**
     * Given a string of url parameters, prepend with '?' if needed, otherwise return parameters as
     * is.
     */
    /**
       * Given a string of url parameters, prepend with '?' if needed, otherwise return parameters as
       * is.
       */
    Location.normalizeQueryParams = /**
       * Given a string of url parameters, prepend with '?' if needed, otherwise return parameters as
       * is.
       */
    function (params) {
        return params && params[0] !== '?' ? '?' + params : params;
    };
    /**
     * Given 2 parts of a url, join them with a slash if needed.
     */
    /**
       * Given 2 parts of a url, join them with a slash if needed.
       */
    Location.joinWithSlash = /**
       * Given 2 parts of a url, join them with a slash if needed.
       */
    function (start, end) {
        if (start.length == 0) {
            return end;
        }
        if (end.length == 0) {
            return start;
        }
        var slashes = 0;
        if (start.endsWith('/')) {
            slashes++;
        }
        if (end.startsWith('/')) {
            slashes++;
        }
        if (slashes == 2) {
            return start + end.substring(1);
        }
        if (slashes == 1) {
            return start + end;
        }
        return start + '/' + end;
    };
    /**
     * If url has a trailing slash, remove it, otherwise return url as is. This
     * method looks for the first occurrence of either #, ?, or the end of the
     * line as `/` characters after any of these should not be replaced.
     */
    /**
       * If url has a trailing slash, remove it, otherwise return url as is. This
       * method looks for the first occurrence of either #, ?, or the end of the
       * line as `/` characters after any of these should not be replaced.
       */
    Location.stripTrailingSlash = /**
       * If url has a trailing slash, remove it, otherwise return url as is. This
       * method looks for the first occurrence of either #, ?, or the end of the
       * line as `/` characters after any of these should not be replaced.
       */
    function (url) {
        var match = url.match(/#|\?|$/);
        var pathEndIdx = match && match.index || url.length;
        var droppedSlashIdx = pathEndIdx - (url[pathEndIdx - 1] === '/' ? 1 : 0);
        return url.slice(0, droppedSlashIdx) + url.slice(pathEndIdx);
    };
    Location.decorators = [
        { type: core_1.Injectable },
    ];
    /** @nocollapse */
    Location.ctorParameters = function () { return [
        { type: location_strategy_1.LocationStrategy, },
    ]; };
    return Location;
}());
exports.Location = Location;
function _stripBaseHref(baseHref, url) {
    return baseHref && url.startsWith(baseHref) ? url.substring(baseHref.length) : url;
}
function _stripIndexHtml(url) {
    return url.replace(/\/index.html$/, '');
}
//# sourceMappingURL=location.js.map