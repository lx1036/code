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
var dom_adapter_1 = require("../dom/dom_adapter");
var dom_tokens_1 = require("../dom/dom_tokens");
/**
 * A service that can be used to get and set the title of a current HTML document.
 *
 * Since an Angular application can't be bootstrapped on the entire HTML document (`<html>` tag)
 * it is not possible to bind to the `text` property of the `HTMLTitleElement` elements
 * (representing the `<title>` tag). Instead, this service can be used to set and get the current
 * title value.
 *
 * @experimental
 */
var Title = /** @class */ (function () {
    function Title(_doc) {
        this._doc = _doc;
    }
    /**
     * Get the title of the current HTML document.
     */
    /**
       * Get the title of the current HTML document.
       */
    Title.prototype.getTitle = /**
       * Get the title of the current HTML document.
       */
    function () { return dom_adapter_1.getDOM().getTitle(this._doc); };
    /**
     * Set the title of the current HTML document.
     * @param newTitle
     */
    /**
       * Set the title of the current HTML document.
       * @param newTitle
       */
    Title.prototype.setTitle = /**
       * Set the title of the current HTML document.
       * @param newTitle
       */
    function (newTitle) { dom_adapter_1.getDOM().setTitle(this._doc, newTitle); };
    Title.decorators = [
        { type: core_1.Injectable },
    ];
    /** @nocollapse */
    Title.ctorParameters = function () { return [
        { type: undefined, decorators: [{ type: core_1.Inject, args: [dom_tokens_1.DOCUMENT,] },] },
    ]; };
    return Title;
}());
exports.Title = Title;
//# sourceMappingURL=title.js.map