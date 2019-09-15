"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var errors_1 = require("./errors");
/**
 *
 * @description
 * Provides a hook for centralized exception handling.
 *
 * The default implementation of `ErrorHandler` prints error messages to the `console`. To
 * intercept error handling, write a custom exception handler that replaces this default as
 * appropriate for your app.
 *
 * ### Example
 *
 * ```
 * class MyErrorHandler implements ErrorHandler {
 *   handleError(error) {
 *     // do something with the exception
 *   }
 * }
 *
 * @NgModule({
 *   providers: [{provide: ErrorHandler, useClass: MyErrorHandler}]
 * })
 * class MyModule {}
 * ```
 *
 *
 */
var /**
 *
 * @description
 * Provides a hook for centralized exception handling.
 *
 * The default implementation of `ErrorHandler` prints error messages to the `console`. To
 * intercept error handling, write a custom exception handler that replaces this default as
 * appropriate for your app.
 *
 * ### Example
 *
 * ```
 * class MyErrorHandler implements ErrorHandler {
 *   handleError(error) {
 *     // do something with the exception
 *   }
 * }
 *
 * @NgModule({
 *   providers: [{provide: ErrorHandler, useClass: MyErrorHandler}]
 * })
 * class MyModule {}
 * ```
 *
 *
 */
ErrorHandler = /** @class */ (function () {
    function ErrorHandler() {
        /**
           * @internal
           */
        this._console = console;
    }
    ErrorHandler.prototype.handleError = function (error) {
        var originalError = this._findOriginalError(error);
        var context = this._findContext(error);
        // Note: Browser consoles show the place from where console.error was called.
        // We can use this to give users additional information about the error.
        var errorLogger = errors_1.getErrorLogger(error);
        errorLogger(this._console, "ERROR", error);
        if (originalError) {
            errorLogger(this._console, "ORIGINAL ERROR", originalError);
        }
        if (context) {
            errorLogger(this._console, 'ERROR CONTEXT', context);
        }
    };
    /** @internal */
    /** @internal */
    ErrorHandler.prototype._findContext = /** @internal */
    function (error) {
        if (error) {
            return errors_1.getDebugContext(error) ? errors_1.getDebugContext(error) :
                this._findContext(errors_1.getOriginalError(error));
        }
        return null;
    };
    /** @internal */
    /** @internal */
    ErrorHandler.prototype._findOriginalError = /** @internal */
    function (error) {
        var e = errors_1.getOriginalError(error);
        while (e && errors_1.getOriginalError(e)) {
            e = errors_1.getOriginalError(e);
        }
        return e;
    };
    return ErrorHandler;
}());
exports.ErrorHandler = ErrorHandler;
function wrappedError(message, originalError) {
    var msg = message + " caused by: " + (originalError instanceof Error ? originalError.message : originalError);
    var error = Error(msg);
    error[errors_1.ERROR_ORIGINAL_ERROR] = originalError;
    return error;
}
exports.wrappedError = wrappedError;
//# sourceMappingURL=error_handler.js.map