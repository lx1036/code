"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @experimental i18n support is experimental.
 */
exports.LOCALE_DATA = {};
/**
 * Register global data to be used internally by Angular. See the
 * {@linkDocs guide/i18n#i18n-pipes "I18n guide"} to know how to import additional locale data.
 *
 * @experimental i18n support is experimental.
 */
// The signature registerLocaleData(data: any, extraData?: any) is deprecated since v5.1
function registerLocaleData(data, localeId, extraData) {
    if (typeof localeId !== 'string') {
        extraData = localeId;
        localeId = data[0 /* LocaleId */];
    }
    localeId = localeId.toLowerCase().replace(/_/g, '-');
    exports.LOCALE_DATA[localeId] = data;
    if (extraData) {
        exports.LOCALE_DATA[localeId][19 /* ExtraData */] = extraData;
    }
}
exports.registerLocaleData = registerLocaleData;
//# sourceMappingURL=locale_data.js.map