"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
/** Called when directives inject each other (creating a circular dependency) */
function throwCyclicDependencyError(token) {
    throw new Error("Cannot instantiate cyclic dependency! " + token);
}
exports.throwCyclicDependencyError = throwCyclicDependencyError;
/** Called when there are multiple component selectors that match a given node */
function throwMultipleComponentError(tNode) {
    throw new Error("Multiple components match node with tagname " + tNode.tagName);
}
exports.throwMultipleComponentError = throwMultipleComponentError;
/** Throws an ExpressionChangedAfterChecked error if checkNoChanges mode is on. */
function throwErrorIfNoChangesMode(creationMode, checkNoChangesMode, oldValue, currValue) {
    if (checkNoChangesMode) {
        var msg = "ExpressionChangedAfterItHasBeenCheckedError: Expression has changed after it was checked. Previous value: '" + oldValue + "'. Current value: '" + currValue + "'.";
        if (creationMode) {
            msg +=
                " It seems like the view has been created after its parent and its children have been dirty checked." +
                    " Has it been created in a change detection hook ?";
        }
        // TODO: include debug context
        throw new Error(msg);
    }
}
exports.throwErrorIfNoChangesMode = throwErrorIfNoChangesMode;
//# sourceMappingURL=errors.js.map