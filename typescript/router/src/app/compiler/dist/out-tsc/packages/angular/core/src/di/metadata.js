"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var decorators_1 = require("../util/decorators");
/**
 * Inject decorator and metadata.
 *
 *
 * @Annotation
 */
exports.Inject = decorators_1.makeParamDecorator('Inject', function (token) { return ({ token: token }); });
/**
 * Optional decorator and metadata.
 *
 *
 * @Annotation
 */
exports.Optional = decorators_1.makeParamDecorator('Optional');
/**
 * Self decorator and metadata.
 *
 *
 * @Annotation
 */
exports.Self = decorators_1.makeParamDecorator('Self');
/**
 * SkipSelf decorator and metadata.
 *
 *
 * @Annotation
 */
exports.SkipSelf = decorators_1.makeParamDecorator('SkipSelf');
/**
 * Host decorator and metadata.
 *
 *
 * @Annotation
 */
exports.Host = decorators_1.makeParamDecorator('Host');
//# sourceMappingURL=metadata.js.map