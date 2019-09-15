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
 * A service that can be used to get and add meta tags.
 *
 * @experimental
 */
var Meta = /** @class */ (function () {
    function Meta(_doc) {
        this._doc = _doc;
        this._dom = dom_adapter_1.getDOM();
    }
    Meta.prototype.addTag = function (tag, forceCreation) {
        if (forceCreation === void 0) { forceCreation = false; }
        if (!tag)
            return null;
        return this._getOrCreateElement(tag, forceCreation);
    };
    Meta.prototype.addTags = function (tags, forceCreation) {
        var _this = this;
        if (forceCreation === void 0) { forceCreation = false; }
        if (!tags)
            return [];
        return tags.reduce(function (result, tag) {
            if (tag) {
                result.push(_this._getOrCreateElement(tag, forceCreation));
            }
            return result;
        }, []);
    };
    Meta.prototype.getTag = function (attrSelector) {
        if (!attrSelector)
            return null;
        return this._dom.querySelector(this._doc, "meta[" + attrSelector + "]") || null;
    };
    Meta.prototype.getTags = function (attrSelector) {
        if (!attrSelector)
            return [];
        var list = this._dom.querySelectorAll(this._doc, "meta[" + attrSelector + "]");
        return list ? [].slice.call(list) : [];
    };
    Meta.prototype.updateTag = function (tag, selector) {
        if (!tag)
            return null;
        selector = selector || this._parseSelector(tag);
        var meta = (this.getTag(selector));
        if (meta) {
            return this._setMetaElementAttributes(tag, meta);
        }
        return this._getOrCreateElement(tag, true);
    };
    Meta.prototype.removeTag = function (attrSelector) { this.removeTagElement((this.getTag(attrSelector))); };
    Meta.prototype.removeTagElement = function (meta) {
        if (meta) {
            this._dom.remove(meta);
        }
    };
    Meta.prototype._getOrCreateElement = function (meta, forceCreation) {
        if (forceCreation === void 0) { forceCreation = false; }
        if (!forceCreation) {
            var selector = this._parseSelector(meta);
            var elem = (this.getTag(selector));
            // It's allowed to have multiple elements with the same name so it's not enough to
            // just check that element with the same name already present on the page. We also need to
            // check if element has tag attributes
            if (elem && this._containsAttributes(meta, elem))
                return elem;
        }
        var element = this._dom.createElement('meta');
        this._setMetaElementAttributes(meta, element);
        var head = this._dom.getElementsByTagName(this._doc, 'head')[0];
        this._dom.appendChild(head, element);
        return element;
    };
    Meta.prototype._setMetaElementAttributes = function (tag, el) {
        var _this = this;
        Object.keys(tag).forEach(function (prop) { return _this._dom.setAttribute(el, prop, tag[prop]); });
        return el;
    };
    Meta.prototype._parseSelector = function (tag) {
        var attr = tag.name ? 'name' : 'property';
        return attr + "=\"" + tag[attr] + "\"";
    };
    Meta.prototype._containsAttributes = function (tag, elem) {
        var _this = this;
        return Object.keys(tag).every(function (key) { return _this._dom.getAttribute(elem, key) === tag[key]; });
    };
    Meta.decorators = [
        { type: core_1.Injectable },
    ];
    /** @nocollapse */
    Meta.ctorParameters = function () { return [
        { type: undefined, decorators: [{ type: core_1.Inject, args: [dom_tokens_1.DOCUMENT,] },] },
    ]; };
    return Meta;
}());
exports.Meta = Meta;
//# sourceMappingURL=meta.js.map