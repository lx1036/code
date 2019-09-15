"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var assert_1 = require("./assert");
/**
 * If this is the first template pass, any ngOnInit or ngDoCheck hooks will be queued into
 * TView.initHooks during directiveCreate.
 *
 * The directive index and hook type are encoded into one number (1st bit: type, remaining bits:
 * directive index), then saved in the even indices of the initHooks array. The odd indices
 * hold the hook functions themselves.
 *
 * @param index The index of the directive in LView.data
 * @param hooks The static hooks map on the directive def
 * @param tView The current TView
 */
function queueInitHooks(index, onInit, doCheck, tView) {
    ngDevMode &&
        assert_1.assertEqual(tView.firstTemplatePass, true, 'Should only be called on first template pass');
    if (onInit) {
        (tView.initHooks || (tView.initHooks = [])).push(index, onInit);
    }
    if (doCheck) {
        (tView.initHooks || (tView.initHooks = [])).push(index, doCheck);
        (tView.checkHooks || (tView.checkHooks = [])).push(index, doCheck);
    }
}
exports.queueInitHooks = queueInitHooks;
/**
 * Loops through the directives on a node and queues all their hooks except ngOnInit
 * and ngDoCheck, which are queued separately in directiveCreate.
 */
function queueLifecycleHooks(flags, currentView) {
    var tView = currentView.tView;
    if (tView.firstTemplatePass === true) {
        var start = flags >> 13 /* DirectiveStartingIndexShift */;
        var count = flags & 4095 /* DirectiveCountMask */;
        var end = start + count;
        // It's necessary to loop through the directives at elementEnd() (rather than processing in
        // directiveCreate) so we can preserve the current hook order. Content, view, and destroy
        // hooks for projected components and directives must be called *before* their hosts.
        for (var i = start; i < end; i++) {
            var def = tView.directives[i];
            queueContentHooks(def, tView, i);
            queueViewHooks(def, tView, i);
            queueDestroyHooks(def, tView, i);
        }
    }
}
exports.queueLifecycleHooks = queueLifecycleHooks;
/** Queues afterContentInit and afterContentChecked hooks on TView */
function queueContentHooks(def, tView, i) {
    if (def.afterContentInit) {
        (tView.contentHooks || (tView.contentHooks = [])).push(i, def.afterContentInit);
    }
    if (def.afterContentChecked) {
        (tView.contentHooks || (tView.contentHooks = [])).push(i, def.afterContentChecked);
        (tView.contentCheckHooks || (tView.contentCheckHooks = [])).push(i, def.afterContentChecked);
    }
}
/** Queues afterViewInit and afterViewChecked hooks on TView */
function queueViewHooks(def, tView, i) {
    if (def.afterViewInit) {
        (tView.viewHooks || (tView.viewHooks = [])).push(i, def.afterViewInit);
    }
    if (def.afterViewChecked) {
        (tView.viewHooks || (tView.viewHooks = [])).push(i, def.afterViewChecked);
        (tView.viewCheckHooks || (tView.viewCheckHooks = [])).push(i, def.afterViewChecked);
    }
}
/** Queues onDestroy hooks on TView */
function queueDestroyHooks(def, tView, i) {
    if (def.onDestroy != null) {
        (tView.destroyHooks || (tView.destroyHooks = [])).push(i, def.onDestroy);
    }
}
/**
 * Calls onInit and doCheck calls if they haven't already been called.
 *
 * @param currentView The current view
 */
function executeInitHooks(currentView, tView, creationMode) {
    if (currentView.lifecycleStage === 1 /* Init */) {
        executeHooks((currentView.directives), tView.initHooks, tView.checkHooks, creationMode);
        currentView.lifecycleStage = 2 /* AfterInit */;
    }
}
exports.executeInitHooks = executeInitHooks;
/**
 * Iterates over afterViewInit and afterViewChecked functions and calls them.
 *
 * @param currentView The current view
 */
function executeHooks(data, allHooks, checkHooks, creationMode) {
    var hooksToCall = creationMode ? allHooks : checkHooks;
    if (hooksToCall) {
        callHooks(data, hooksToCall);
    }
}
exports.executeHooks = executeHooks;
/**
 * Calls lifecycle hooks with their contexts, skipping init hooks if it's not
 * creation mode.
 *
 * @param currentView The current view
 * @param arr The array in which the hooks are found
 */
function callHooks(data, arr) {
    for (var i = 0; i < arr.length; i += 2) {
        arr[i + 1].call(data[arr[i]]);
    }
}
exports.callHooks = callHooks;
//# sourceMappingURL=hooks.js.map