"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var util_1 = require("../render3/util");
var html_sanitizer_1 = require("./html_sanitizer");
var style_sanitizer_1 = require("./style_sanitizer");
var url_sanitizer_1 = require("./url_sanitizer");
var BRAND = '__SANITIZER_TRUSTED_BRAND__';
/**
 * An `html` sanitizer which converts untrusted `html` **string** into trusted string by removing
 * dangerous content.
 *
 * This method parses the `html` and locates potentially dangerous content (such as urls and
 * javascript) and removes it.
 *
 * It is possible to mark a string as trusted by calling {@link bypassSanitizationTrustHtml}.
 *
 * @param unsafeHtml untrusted `html`, typically from the user.
 * @returns `html` string which is safe to display to user, because all of the dangerous javascript
 * and urls have been removed.
 */
function sanitizeHtml(unsafeHtml) {
    if (unsafeHtml instanceof String && unsafeHtml[BRAND] === 'Html') {
        return unsafeHtml.toString();
    }
    return html_sanitizer_1._sanitizeHtml(document, util_1.stringify(unsafeHtml));
}
exports.sanitizeHtml = sanitizeHtml;
/**
 * A `style` sanitizer which converts untrusted `style` **string** into trusted string by removing
 * dangerous content.
 *
 * This method parses the `style` and locates potentially dangerous content (such as urls and
 * javascript) and removes it.
 *
 * It is possible to mark a string as trusted by calling {@link bypassSanitizationTrustStyle}.
 *
 * @param unsafeStyle untrusted `style`, typically from the user.
 * @returns `style` string which is safe to bind to the `style` properties, because all of the
 * dangerous javascript and urls have been removed.
 */
function sanitizeStyle(unsafeStyle) {
    if (unsafeStyle instanceof String && unsafeStyle[BRAND] === 'Style') {
        return unsafeStyle.toString();
    }
    return style_sanitizer_1._sanitizeStyle(util_1.stringify(unsafeStyle));
}
exports.sanitizeStyle = sanitizeStyle;
/**
 * A `url` sanitizer which converts untrusted `url` **string** into trusted string by removing
 * dangerous
 * content.
 *
 * This method parses the `url` and locates potentially dangerous content (such as javascript) and
 * removes it.
 *
 * It is possible to mark a string as trusted by calling {@link bypassSanitizationTrustUrl}.
 *
 * @param unsafeUrl untrusted `url`, typically from the user.
 * @returns `url` string which is safe to bind to the `src` properties such as `<img src>`, because
 * all of the dangerous javascript has been removed.
 */
function sanitizeUrl(unsafeUrl) {
    if (unsafeUrl instanceof String && unsafeUrl[BRAND] === 'Url') {
        return unsafeUrl.toString();
    }
    return url_sanitizer_1._sanitizeUrl(util_1.stringify(unsafeUrl));
}
exports.sanitizeUrl = sanitizeUrl;
/**
 * A `url` sanitizer which only lets trusted `url`s through.
 *
 * This passes only `url`s marked trusted by calling {@link bypassSanitizationTrustResourceUrl}.
 *
 * @param unsafeResourceUrl untrusted `url`, typically from the user.
 * @returns `url` string which is safe to bind to the `src` properties such as `<img src>`, because
 * only trusted `url`s have been allowed to pass.
 */
function sanitizeResourceUrl(unsafeResourceUrl) {
    if (unsafeResourceUrl instanceof String &&
        unsafeResourceUrl[BRAND] === 'ResourceUrl') {
        return unsafeResourceUrl.toString();
    }
    throw new Error('unsafe value used in a resource URL context (see http://g.co/ng/security#xss)');
}
exports.sanitizeResourceUrl = sanitizeResourceUrl;
/**
 * A `script` sanitizer which only lets trusted javascript through.
 *
 * This passes only `script`s marked trusted by calling {@link bypassSanitizationTrustScript}.
 *
 * @param unsafeScript untrusted `script`, typically from the user.
 * @returns `url` string which is safe to bind to the `<script>` element such as `<img src>`,
 * because only trusted `scripts`s have been allowed to pass.
 */
function sanitizeScript(unsafeScript) {
    if (unsafeScript instanceof String && unsafeScript[BRAND] === 'Script') {
        return unsafeScript.toString();
    }
    throw new Error('unsafe value used in a script context');
}
exports.sanitizeScript = sanitizeScript;
/**
 * Mark `html` string as trusted.
 *
 * This function wraps the trusted string in `String` and brands it in a way which makes it
 * recognizable to {@link htmlSanitizer} to be trusted implicitly.
 *
 * @param trustedHtml `html` string which needs to be implicitly trusted.
 * @returns a `html` `String` which has been branded to be implicitly trusted.
 */
function bypassSanitizationTrustHtml(trustedHtml) {
    return bypassSanitizationTrustString(trustedHtml, 'Html');
}
exports.bypassSanitizationTrustHtml = bypassSanitizationTrustHtml;
/**
 * Mark `style` string as trusted.
 *
 * This function wraps the trusted string in `String` and brands it in a way which makes it
 * recognizable to {@link styleSanitizer} to be trusted implicitly.
 *
 * @param trustedStyle `style` string which needs to be implicitly trusted.
 * @returns a `style` `String` which has been branded to be implicitly trusted.
 */
function bypassSanitizationTrustStyle(trustedStyle) {
    return bypassSanitizationTrustString(trustedStyle, 'Style');
}
exports.bypassSanitizationTrustStyle = bypassSanitizationTrustStyle;
/**
 * Mark `script` string as trusted.
 *
 * This function wraps the trusted string in `String` and brands it in a way which makes it
 * recognizable to {@link scriptSanitizer} to be trusted implicitly.
 *
 * @param trustedScript `script` string which needs to be implicitly trusted.
 * @returns a `script` `String` which has been branded to be implicitly trusted.
 */
function bypassSanitizationTrustScript(trustedScript) {
    return bypassSanitizationTrustString(trustedScript, 'Script');
}
exports.bypassSanitizationTrustScript = bypassSanitizationTrustScript;
/**
 * Mark `url` string as trusted.
 *
 * This function wraps the trusted string in `String` and brands it in a way which makes it
 * recognizable to {@link urlSanitizer} to be trusted implicitly.
 *
 * @param trustedUrl `url` string which needs to be implicitly trusted.
 * @returns a `url` `String` which has been branded to be implicitly trusted.
 */
function bypassSanitizationTrustUrl(trustedUrl) {
    return bypassSanitizationTrustString(trustedUrl, 'Url');
}
exports.bypassSanitizationTrustUrl = bypassSanitizationTrustUrl;
/**
 * Mark `url` string as trusted.
 *
 * This function wraps the trusted string in `String` and brands it in a way which makes it
 * recognizable to {@link resourceUrlSanitizer} to be trusted implicitly.
 *
 * @param trustedResourceUrl `url` string which needs to be implicitly trusted.
 * @returns a `url` `String` which has been branded to be implicitly trusted.
 */
function bypassSanitizationTrustResourceUrl(trustedResourceUrl) {
    return bypassSanitizationTrustString(trustedResourceUrl, 'ResourceUrl');
}
exports.bypassSanitizationTrustResourceUrl = bypassSanitizationTrustResourceUrl;
function bypassSanitizationTrustString(trustedString, mode) {
    var trusted = new String(trustedString);
    trusted[BRAND] = mode;
    return trusted;
}
//# sourceMappingURL=sanitization.js.map