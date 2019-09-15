"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var instructions_1 = require("./instructions");
/**
 * If the value hasn't been saved, calls the pure function to store and return the
 * value. If it has been saved, returns the saved value.
 *
 * @param pureFn Function that returns a value
 * @returns value
 */
function pureFunction0(pureFn, thisArg) {
    return instructions_1.getCreationMode() ? instructions_1.checkAndUpdateBinding(thisArg ? pureFn.call(thisArg) : pureFn()) :
        instructions_1.consumeBinding();
}
exports.pureFunction0 = pureFunction0;
/**
 * If the value of the provided exp has changed, calls the pure function to return
 * an updated value. Or if the value has not changed, returns cached value.
 *
 * @param pureFn Function that returns an updated value
 * @param exp Updated expression value
 * @returns Updated value
 */
function pureFunction1(pureFn, exp, thisArg) {
    return instructions_1.bindingUpdated(exp) ?
        instructions_1.checkAndUpdateBinding(thisArg ? pureFn.call(thisArg, exp) : pureFn(exp)) :
        instructions_1.consumeBinding();
}
exports.pureFunction1 = pureFunction1;
/**
 * If the value of any provided exp has changed, calls the pure function to return
 * an updated value. Or if no values have changed, returns cached value.
 *
 * @param pureFn
 * @param exp1
 * @param exp2
 * @returns Updated value
 */
function pureFunction2(pureFn, exp1, exp2, thisArg) {
    return instructions_1.bindingUpdated2(exp1, exp2) ?
        instructions_1.checkAndUpdateBinding(thisArg ? pureFn.call(thisArg, exp1, exp2) : pureFn(exp1, exp2)) :
        instructions_1.consumeBinding();
}
exports.pureFunction2 = pureFunction2;
/**
 * If the value of any provided exp has changed, calls the pure function to return
 * an updated value. Or if no values have changed, returns cached value.
 *
 * @param pureFn
 * @param exp1
 * @param exp2
 * @param exp3
 * @returns Updated value
 */
function pureFunction3(pureFn, exp1, exp2, exp3, thisArg) {
    var different = instructions_1.bindingUpdated2(exp1, exp2);
    return instructions_1.bindingUpdated(exp3) || different ?
        instructions_1.checkAndUpdateBinding(thisArg ? pureFn.call(thisArg, exp1, exp2, exp3) : pureFn(exp1, exp2, exp3)) :
        instructions_1.consumeBinding();
}
exports.pureFunction3 = pureFunction3;
/**
 * If the value of any provided exp has changed, calls the pure function to return
 * an updated value. Or if no values have changed, returns cached value.
 *
 * @param pureFn
 * @param exp1
 * @param exp2
 * @param exp3
 * @param exp4
 * @returns Updated value
 */
function pureFunction4(pureFn, exp1, exp2, exp3, exp4, thisArg) {
    return instructions_1.bindingUpdated4(exp1, exp2, exp3, exp4) ?
        instructions_1.checkAndUpdateBinding(thisArg ? pureFn.call(thisArg, exp1, exp2, exp3, exp4) : pureFn(exp1, exp2, exp3, exp4)) :
        instructions_1.consumeBinding();
}
exports.pureFunction4 = pureFunction4;
/**
 * If the value of any provided exp has changed, calls the pure function to return
 * an updated value. Or if no values have changed, returns cached value.
 *
 * @param pureFn
 * @param exp1
 * @param exp2
 * @param exp3
 * @param exp4
 * @param exp5
 * @returns Updated value
 */
function pureFunction5(pureFn, exp1, exp2, exp3, exp4, exp5, thisArg) {
    var different = instructions_1.bindingUpdated4(exp1, exp2, exp3, exp4);
    return instructions_1.bindingUpdated(exp5) || different ?
        instructions_1.checkAndUpdateBinding(thisArg ? pureFn.call(thisArg, exp1, exp2, exp3, exp4, exp5) :
            pureFn(exp1, exp2, exp3, exp4, exp5)) :
        instructions_1.consumeBinding();
}
exports.pureFunction5 = pureFunction5;
/**
 * If the value of any provided exp has changed, calls the pure function to return
 * an updated value. Or if no values have changed, returns cached value.
 *
 * @param pureFn
 * @param exp1
 * @param exp2
 * @param exp3
 * @param exp4
 * @param exp5
 * @param exp6
 * @returns Updated value
 */
function pureFunction6(pureFn, exp1, exp2, exp3, exp4, exp5, exp6, thisArg) {
    var different = instructions_1.bindingUpdated4(exp1, exp2, exp3, exp4);
    return instructions_1.bindingUpdated2(exp5, exp6) || different ?
        instructions_1.checkAndUpdateBinding(thisArg ? pureFn.call(thisArg, exp1, exp2, exp3, exp4, exp5, exp6) :
            pureFn(exp1, exp2, exp3, exp4, exp5, exp6)) :
        instructions_1.consumeBinding();
}
exports.pureFunction6 = pureFunction6;
/**
 * If the value of any provided exp has changed, calls the pure function to return
 * an updated value. Or if no values have changed, returns cached value.
 *
 * @param pureFn
 * @param exp1
 * @param exp2
 * @param exp3
 * @param exp4
 * @param exp5
 * @param exp6
 * @param exp7
 * @returns Updated value
 */
function pureFunction7(pureFn, exp1, exp2, exp3, exp4, exp5, exp6, exp7, thisArg) {
    var different = instructions_1.bindingUpdated4(exp1, exp2, exp3, exp4);
    different = instructions_1.bindingUpdated2(exp5, exp6) || different;
    return instructions_1.bindingUpdated(exp7) || different ?
        instructions_1.checkAndUpdateBinding(thisArg ? pureFn.call(thisArg, exp1, exp2, exp3, exp4, exp5, exp6, exp7) :
            pureFn(exp1, exp2, exp3, exp4, exp5, exp6, exp7)) :
        instructions_1.consumeBinding();
}
exports.pureFunction7 = pureFunction7;
/**
 * If the value of any provided exp has changed, calls the pure function to return
 * an updated value. Or if no values have changed, returns cached value.
 *
 * @param pureFn
 * @param exp1
 * @param exp2
 * @param exp3
 * @param exp4
 * @param exp5
 * @param exp6
 * @param exp7
 * @param exp8
 * @returns Updated value
 */
function pureFunction8(pureFn, exp1, exp2, exp3, exp4, exp5, exp6, exp7, exp8, thisArg) {
    var different = instructions_1.bindingUpdated4(exp1, exp2, exp3, exp4);
    return instructions_1.bindingUpdated4(exp5, exp6, exp7, exp8) || different ?
        instructions_1.checkAndUpdateBinding(thisArg ? pureFn.call(thisArg, exp1, exp2, exp3, exp4, exp5, exp6, exp7, exp8) :
            pureFn(exp1, exp2, exp3, exp4, exp5, exp6, exp7, exp8)) :
        instructions_1.consumeBinding();
}
exports.pureFunction8 = pureFunction8;
/**
 * pureFunction instruction that can support any number of bindings.
 *
 * If the value of any provided exp has changed, calls the pure function to return
 * an updated value. Or if no values have changed, returns cached value.
 *
 * @param pureFn A pure function that takes binding values and builds an object or array
 * containing those values.
 * @param exp An array of binding values
 * @returns Updated value
 */
function pureFunctionV(pureFn, exps, thisArg) {
    var different = false;
    for (var i = 0; i < exps.length; i++) {
        instructions_1.bindingUpdated(exps[i]) && (different = true);
    }
    return different ? instructions_1.checkAndUpdateBinding(pureFn.apply(thisArg, exps)) : instructions_1.consumeBinding();
}
exports.pureFunctionV = pureFunctionV;
//# sourceMappingURL=pure_function.js.map