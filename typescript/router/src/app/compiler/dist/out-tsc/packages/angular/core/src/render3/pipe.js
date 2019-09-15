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
var pure_function_1 = require("./pure_function");
/**
 * Create a pipe.
 *
 * @param index Pipe index where the pipe will be stored.
 * @param pipeName The name of the pipe
 * @returns T the instance of the pipe.
 */
function pipe(index, pipeName) {
    var tView = instructions_1.getTView();
    var pipeDef;
    if (tView.firstTemplatePass) {
        pipeDef = getPipeDef(pipeName, tView.pipeRegistry);
        tView.data[index] = pipeDef;
        if (pipeDef.onDestroy) {
            (tView.pipeDestroyHooks || (tView.pipeDestroyHooks = [])).push(index, pipeDef.onDestroy);
        }
    }
    else {
        pipeDef = tView.data[index];
    }
    var pipeInstance = pipeDef.n();
    instructions_1.store(index, pipeInstance);
    return pipeInstance;
}
exports.pipe = pipe;
/**
 * Searches the pipe registry for a pipe with the given name. If one is found,
 * returns the pipe. Otherwise, an error is thrown because the pipe cannot be resolved.
 *
 * @param name Name of pipe to resolve
 * @param registry Full list of available pipes
 * @returns Matching PipeDef
 */
function getPipeDef(name, registry) {
    if (registry) {
        for (var i = 0; i < registry.length; i++) {
            var pipeDef = registry[i];
            if (name === pipeDef.name) {
                return pipeDef;
            }
        }
    }
    throw new Error("Pipe with name '" + name + "' not found!");
}
/**
 * Invokes a pipe with 1 arguments.
 *
 * This instruction acts as a guard to {@link PipeTransform#transform} invoking
 * the pipe only when an input to the pipe changes.
 *
 * @param index Pipe index where the pipe was stored on creation.
 * @param v1 1st argument to {@link PipeTransform#transform}.
 */
function pipeBind1(index, v1) {
    var pipeInstance = instructions_1.load(index);
    return isPure(index) ? pure_function_1.pureFunction1(pipeInstance.transform, v1, pipeInstance) :
        pipeInstance.transform(v1);
}
exports.pipeBind1 = pipeBind1;
/**
 * Invokes a pipe with 2 arguments.
 *
 * This instruction acts as a guard to {@link PipeTransform#transform} invoking
 * the pipe only when an input to the pipe changes.
 *
 * @param index Pipe index where the pipe was stored on creation.
 * @param v1 1st argument to {@link PipeTransform#transform}.
 * @param v2 2nd argument to {@link PipeTransform#transform}.
 */
function pipeBind2(index, v1, v2) {
    var pipeInstance = instructions_1.load(index);
    return isPure(index) ? pure_function_1.pureFunction2(pipeInstance.transform, v1, v2, pipeInstance) :
        pipeInstance.transform(v1, v2);
}
exports.pipeBind2 = pipeBind2;
/**
 * Invokes a pipe with 3 arguments.
 *
 * This instruction acts as a guard to {@link PipeTransform#transform} invoking
 * the pipe only when an input to the pipe changes.
 *
 * @param index Pipe index where the pipe was stored on creation.
 * @param v1 1st argument to {@link PipeTransform#transform}.
 * @param v2 2nd argument to {@link PipeTransform#transform}.
 * @param v3 4rd argument to {@link PipeTransform#transform}.
 */
function pipeBind3(index, v1, v2, v3) {
    var pipeInstance = instructions_1.load(index);
    return isPure(index) ? pure_function_1.pureFunction3(pipeInstance.transform.bind(pipeInstance), v1, v2, v3) :
        pipeInstance.transform(v1, v2, v3);
}
exports.pipeBind3 = pipeBind3;
/**
 * Invokes a pipe with 4 arguments.
 *
 * This instruction acts as a guard to {@link PipeTransform#transform} invoking
 * the pipe only when an input to the pipe changes.
 *
 * @param index Pipe index where the pipe was stored on creation.
 * @param v1 1st argument to {@link PipeTransform#transform}.
 * @param v2 2nd argument to {@link PipeTransform#transform}.
 * @param v3 3rd argument to {@link PipeTransform#transform}.
 * @param v4 4th argument to {@link PipeTransform#transform}.
 */
function pipeBind4(index, v1, v2, v3, v4) {
    var pipeInstance = instructions_1.load(index);
    return isPure(index) ? pure_function_1.pureFunction4(pipeInstance.transform, v1, v2, v3, v4, pipeInstance) :
        pipeInstance.transform(v1, v2, v3, v4);
}
exports.pipeBind4 = pipeBind4;
/**
 * Invokes a pipe with variable number of arguments.
 *
 * This instruction acts as a guard to {@link PipeTransform#transform} invoking
 * the pipe only when an input to the pipe changes.
 *
 * @param index Pipe index where the pipe was stored on creation.
 * @param values Array of arguments to pass to {@link PipeTransform#transform} method.
 */
function pipeBindV(index, values) {
    var pipeInstance = instructions_1.load(index);
    return isPure(index) ? pure_function_1.pureFunctionV(pipeInstance.transform, values, pipeInstance) :
        pipeInstance.transform.apply(pipeInstance, values);
}
exports.pipeBindV = pipeBindV;
function isPure(index) {
    return instructions_1.getTView().data[index].pure;
}
//# sourceMappingURL=pipe.js.map