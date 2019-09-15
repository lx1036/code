"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var hooks_1 = require("./hooks");
var container_1 = require("./interfaces/container");
var node_1 = require("./interfaces/node");
var projection_1 = require("./interfaces/projection");
var renderer_1 = require("./interfaces/renderer");
var view_1 = require("./interfaces/view");
var node_assert_1 = require("./node_assert");
var util_1 = require("./util");
var unusedValueToPlacateAjd = container_1.unusedValueExportToPlacateAjd + node_1.unusedValueExportToPlacateAjd + projection_1.unusedValueExportToPlacateAjd + renderer_1.unusedValueExportToPlacateAjd + view_1.unusedValueExportToPlacateAjd;
/**
 * Returns the first RNode following the given LNode in the same parent DOM element.
 *
 * This is needed in order to insert the given node with insertBefore.
 *
 * @param node The node whose following DOM node must be found.
 * @param stopNode A parent node at which the lookup in the tree should be stopped, or null if the
 * lookup should not be stopped until the result is found.
 * @returns RNode before which the provided node should be inserted or null if the lookup was
 * stopped
 * or if there is no native node after the given logical node in the same native parent.
 */
function findNextRNodeSibling(node, stopNode) {
    var currentNode = node;
    while (currentNode && currentNode !== stopNode) {
        var pNextOrParent = currentNode.pNextOrParent;
        if (pNextOrParent) {
            while (pNextOrParent.type !== 1 /* Projection */) {
                var nativeNode = findFirstRNode(pNextOrParent);
                if (nativeNode) {
                    return nativeNode;
                }
                pNextOrParent = (pNextOrParent.pNextOrParent);
            }
            currentNode = pNextOrParent;
        }
        else {
            var currentSibling = currentNode.next;
            while (currentSibling) {
                var nativeNode = findFirstRNode(currentSibling);
                if (nativeNode) {
                    return nativeNode;
                }
                currentSibling = currentSibling.next;
            }
            var parentNode = currentNode.parent;
            currentNode = null;
            if (parentNode) {
                var parentType = parentNode.type;
                if (parentType === 0 /* Container */ || parentType === 2 /* View */) {
                    currentNode = parentNode;
                }
            }
        }
    }
    return null;
}
/**
 * Get the next node in the LNode tree, taking into account the place where a node is
 * projected (in the shadow DOM) rather than where it comes from (in the light DOM).
 *
 * @param node The node whose next node in the LNode tree must be found.
 * @return LNode|null The next sibling in the LNode tree.
 */
function getNextLNodeWithProjection(node) {
    var pNextOrParent = node.pNextOrParent;
    if (pNextOrParent) {
        // The node is projected
        var isLastProjectedNode = pNextOrParent.type === 1 /* Projection */;
        // returns pNextOrParent if we are not at the end of the list, null otherwise
        return isLastProjectedNode ? null : pNextOrParent;
    }
    // returns node.next because the the node is not projected
    return node.next;
}
/**
 * Find the next node in the LNode tree, taking into account the place where a node is
 * projected (in the shadow DOM) rather than where it comes from (in the light DOM).
 *
 * If there is no sibling node, this function goes to the next sibling of the parent node...
 * until it reaches rootNode (at which point null is returned).
 *
 * @param initialNode The node whose following node in the LNode tree must be found.
 * @param rootNode The root node at which the lookup should stop.
 * @return LNode|null The following node in the LNode tree.
 */
function getNextOrParentSiblingNode(initialNode, rootNode) {
    var node = initialNode;
    var nextNode = getNextLNodeWithProjection(node);
    while (node && !nextNode) {
        // if node.pNextOrParent is not null here, it is not the next node
        // (because, at this point, nextNode is null, so it is the parent)
        node = node.pNextOrParent || node.parent;
        if (node === rootNode) {
            return null;
        }
        nextNode = node && getNextLNodeWithProjection(node);
    }
    return nextNode;
}
/**
 * Returns the first RNode inside the given LNode.
 *
 * @param node The node whose first DOM node must be found
 * @returns RNode The first RNode of the given LNode or null if there is none.
 */
function findFirstRNode(rootNode) {
    var node = rootNode;
    while (node) {
        var nextNode = null;
        if (node.type === 3 /* Element */) {
            // A LElementNode has a matching RNode in LElementNode.native
            return node.native;
        }
        else if (node.type === 0 /* Container */) {
            var lContainerNode = node;
            var childContainerData = lContainerNode.dynamicLContainerNode ?
                lContainerNode.dynamicLContainerNode.data :
                lContainerNode.data;
            nextNode = childContainerData.views.length ? childContainerData.views[0].child : null;
        }
        else if (node.type === 1 /* Projection */) {
            // For Projection look at the first projected node
            nextNode = node.data.head;
        }
        else {
            // Otherwise look at the first child
            nextNode = node.child;
        }
        node = nextNode === null ? getNextOrParentSiblingNode(node, rootNode) : nextNode;
    }
    return null;
}
function createTextNode(value, renderer) {
    return renderer_1.isProceduralRenderer(renderer) ? renderer.createText(util_1.stringify(value)) :
        renderer.createTextNode(util_1.stringify(value));
}
exports.createTextNode = createTextNode;
function addRemoveViewFromContainer(container, rootNode, insertMode, beforeNode) {
    ngDevMode && node_assert_1.assertNodeType(container, 0 /* Container */);
    ngDevMode && node_assert_1.assertNodeType(rootNode, 2 /* View */);
    var parentNode = container.data.renderParent;
    var parent = parentNode ? parentNode.native : null;
    var node = rootNode.child;
    if (parent) {
        while (node) {
            var nextNode = null;
            var renderer = container.view.renderer;
            if (node.type === 3 /* Element */) {
                if (insertMode) {
                    if (!node.native) {
                        // If the native element doesn't exist, this is a bound text node that hasn't yet been
                        // created because update mode has not run (occurs when a bound text node is a root
                        // node of a dynamically created view). See textBinding() in instructions for ctx.
                        // If the native element doesn't exist, this is a bound text node that hasn't yet been
                        // created because update mode has not run (occurs when a bound text node is a root
                        // node of a dynamically created view). See textBinding() in instructions for ctx.
                        node.native = createTextNode('', renderer);
                    }
                    renderer_1.isProceduralRenderer(renderer) ?
                        renderer.insertBefore(parent, (node.native), beforeNode) :
                        parent.insertBefore((node.native), beforeNode, true);
                }
                else {
                    renderer_1.isProceduralRenderer(renderer) ? renderer.removeChild(parent, (node.native)) :
                        parent.removeChild((node.native));
                }
                nextNode = node.next;
            }
            else if (node.type === 0 /* Container */) {
                // if we get to a container, it must be a root node of a view because we are only
                // propagating down into child views / containers and not child elements
                var childContainerData = node.data;
                childContainerData.renderParent = parentNode;
                nextNode = childContainerData.views.length ? childContainerData.views[0].child : null;
            }
            else if (node.type === 1 /* Projection */) {
                nextNode = node.data.head;
            }
            else {
                nextNode = node.child;
            }
            if (nextNode === null) {
                node = getNextOrParentSiblingNode(node, rootNode);
            }
            else {
                node = nextNode;
            }
        }
    }
}
exports.addRemoveViewFromContainer = addRemoveViewFromContainer;
/**
 * Traverses the tree of component views and containers to remove listeners and
 * call onDestroy callbacks.
 *
 * Notes:
 *  - Because it's used for onDestroy calls, it needs to be bottom-up.
 *  - Must process containers instead of their views to avoid splicing
 *  when views are destroyed and re-added.
 *  - Using a while loop because it's faster than recursion
 *  - Destroy only called on movement to sibling or movement to parent (laterally or up)
 *
 *  @param rootView The view to destroy
 */
function destroyViewTree(rootView) {
    var viewOrContainer = rootView;
    while (viewOrContainer) {
        var next = null;
        if (viewOrContainer.views && viewOrContainer.views.length) {
            next = viewOrContainer.views[0].data;
        }
        else if (viewOrContainer.child) {
            next = viewOrContainer.child;
        }
        else if (viewOrContainer.next) {
            cleanUpView(viewOrContainer);
            next = viewOrContainer.next;
        }
        if (next == null) {
            // If the viewOrContainer is the rootView, then the cleanup is done twice.
            // Without this check, ngOnDestroy would be called twice for a directive on an element.
            while (viewOrContainer && !viewOrContainer.next && viewOrContainer !== rootView) {
                cleanUpView(viewOrContainer);
                viewOrContainer = getParentState(viewOrContainer, rootView);
            }
            cleanUpView(viewOrContainer || rootView);
            next = viewOrContainer && viewOrContainer.next;
        }
        viewOrContainer = next;
    }
}
exports.destroyViewTree = destroyViewTree;
/**
 * Inserts a view into a container.
 *
 * This adds the view to the container's array of active views in the correct
 * position. It also adds the view's elements to the DOM if the container isn't a
 * root node of another view (in that case, the view's elements will be added when
 * the container's parent view is added later).
 *
 * @param container The container into which the view should be inserted
 * @param newView The view to insert
 * @param index The index at which to insert the view
 * @returns The inserted view
 */
function insertView(container, newView, index) {
    var state = container.data;
    var views = state.views;
    if (index > 0) {
        // This is a new view, we need to add it to the children.
        setViewNext(views[index - 1], newView);
    }
    if (index < views.length) {
        setViewNext(newView, views[index]);
        views.splice(index, 0, newView);
    }
    else {
        views.push(newView);
    }
    // If the container's renderParent is null, we know that it is a root node of its own parent view
    // and we should wait until that parent processes its nodes (otherwise, we will insert this view's
    // nodes twice - once now and once when its parent inserts its views).
    if (container.data.renderParent !== null) {
        var beforeNode = findNextRNodeSibling(newView, container);
        if (!beforeNode) {
            var containerNextNativeNode = container.native;
            if (containerNextNativeNode === undefined) {
                containerNextNativeNode = container.native = findNextRNodeSibling(container, null);
            }
            beforeNode = containerNextNativeNode;
        }
        addRemoveViewFromContainer(container, newView, true, beforeNode);
    }
    return newView;
}
exports.insertView = insertView;
/**
 * Removes a view from a container.
 *
 * This method splices the view from the container's array of active views. It also
 * removes the view's elements from the DOM and conducts cleanup (e.g. removing
 * listeners, calling onDestroys).
 *
 * @param container The container from which to remove a view
 * @param removeIndex The index of the view to remove
 * @returns The removed view
 */
function removeView(container, removeIndex) {
    var views = container.data.views;
    var viewNode = views[removeIndex];
    if (removeIndex > 0) {
        setViewNext(views[removeIndex - 1], viewNode.next);
    }
    views.splice(removeIndex, 1);
    viewNode.next = null;
    destroyViewTree(viewNode.data);
    addRemoveViewFromContainer(container, viewNode, false);
    // Notify query that view has been removed
    container.data.queries && container.data.queries.removeView(removeIndex);
    return viewNode;
}
exports.removeView = removeView;
/**
 * Sets a next on the view node, so views in for loops can easily jump from
 * one view to the next to add/remove elements. Also adds the LView (view.data)
 * to the view tree for easy traversal when cleaning up the view.
 *
 * @param view The view to set up
 * @param next The view's new next
 */
function setViewNext(view, next) {
    view.next = next;
    view.data.next = next ? next.data : null;
}
exports.setViewNext = setViewNext;
/**
 * Determines which LViewOrLContainer to jump to when traversing back up the
 * tree in destroyViewTree.
 *
 * Normally, the view's parent LView should be checked, but in the case of
 * embedded views, the container (which is the view node's parent, but not the
 * LView's parent) needs to be checked for a possible next property.
 *
 * @param state The LViewOrLContainer for which we need a parent state
 * @param rootView The rootView, so we don't propagate too far up the view tree
 * @returns The correct parent LViewOrLContainer
 */
function getParentState(state, rootView) {
    var node;
    if ((node = state.node) && node.type === 2 /* View */) {
        // if it's an embedded view, the state needs to go up to the container, in case the
        // container has a next
        return node.parent.data;
    }
    else {
        // otherwise, use parent view for containers or component views
        return state.parent === rootView ? null : state.parent;
    }
}
exports.getParentState = getParentState;
/**
 * Removes all listeners and call all onDestroys in a given view.
 *
 * @param view The LView to clean up
 */
function cleanUpView(view) {
    removeListeners(view);
    executeOnDestroys(view);
    executePipeOnDestroys(view);
}
/** Removes listeners and unsubscribes from output subscriptions */
function removeListeners(view) {
    var cleanup = (view.cleanup);
    if (cleanup != null) {
        for (var i = 0; i < cleanup.length - 1; i += 2) {
            if (typeof cleanup[i] === 'string') {
                cleanup[i + 1].removeEventListener(cleanup[i], cleanup[i + 2], cleanup[i + 3]);
                i += 2;
            }
            else {
                cleanup[i].call(cleanup[i + 1]);
            }
        }
        view.cleanup = null;
    }
}
/** Calls onDestroy hooks for this view */
function executeOnDestroys(view) {
    var tView = view.tView;
    var destroyHooks;
    if (tView != null && (destroyHooks = tView.destroyHooks) != null) {
        hooks_1.callHooks((view.directives), destroyHooks);
    }
}
/** Calls pipe destroy hooks for this view */
function executePipeOnDestroys(view) {
    var pipeDestroyHooks = view.tView && view.tView.pipeDestroyHooks;
    if (pipeDestroyHooks) {
        hooks_1.callHooks((view.data), pipeDestroyHooks);
    }
}
/**
 * Returns whether a native element should be inserted in the given parent.
 *
 * The native node can be inserted when its parent is:
 * - A regular element => Yes
 * - A component host element =>
 *    - if the `currentView` === the parent `view`: The element is in the content (vs the
 *      template)
 *      => don't add as the parent component will project if needed.
 *    - `currentView` !== the parent `view` => The element is in the template (vs the content),
 *      add it
 * - View element => delay insertion, will be done on `viewEnd()`
 *
 * @param parent The parent in which to insert the child
 * @param currentView The LView being processed
 * @return boolean Whether the child element should be inserted.
 */
function canInsertNativeNode(parent, currentView) {
    var parentIsElement = parent.type === 3 /* Element */;
    return parentIsElement &&
        (parent.view !== currentView || parent.data === null /* Regular Element. */);
}
exports.canInsertNativeNode = canInsertNativeNode;
/**
 * Appends the `child` element to the `parent`.
 *
 * The element insertion might be delayed {@link canInsertNativeNode}
 *
 * @param parent The parent to which to append the child
 * @param child The child that should be appended
 * @param currentView The current LView
 * @returns Whether or not the child was appended
 */
function appendChild(parent, child, currentView) {
    if (child !== null && canInsertNativeNode(parent, currentView)) {
        // We only add element if not in View or not projected.
        var renderer = currentView.renderer;
        renderer_1.isProceduralRenderer(renderer) ? renderer.appendChild(parent.native, child) :
            parent.native.appendChild(child);
        return true;
    }
    return false;
}
exports.appendChild = appendChild;
/**
 * Inserts the provided node before the correct element in the DOM.
 *
 * The element insertion might be delayed {@link canInsertNativeNode}
 *
 * @param node Node to insert
 * @param currentView Current LView
 */
function insertChild(node, currentView) {
    var parent = (node.parent);
    if (canInsertNativeNode(parent, currentView)) {
        var nativeSibling = findNextRNodeSibling(node, null);
        var renderer = currentView.renderer;
        renderer_1.isProceduralRenderer(renderer) ?
            renderer.insertBefore((parent.native), (node.native), nativeSibling) :
            parent.native.insertBefore((node.native), nativeSibling, false);
    }
}
exports.insertChild = insertChild;
/**
 * Appends a projected node to the DOM, or in the case of a projected container,
 * appends the nodes from all of the container's active views to the DOM.
 *
 * @param node The node to process
 * @param currentParent The last parent element to be processed
 * @param currentView Current LView
 */
function appendProjectedNode(node, currentParent, currentView) {
    if (node.type !== 0 /* Container */) {
        appendChild(currentParent, node.native, currentView);
    }
    else {
        // The node we are adding is a Container and we are adding it to Element which
        // is not a component (no more re-projection).
        // Alternatively a container is projected at the root of a component's template
        // and can't be re-projected (as not content of any component).
        // Assignee the final projection location in those cases.
        var lContainer = node.data;
        lContainer.renderParent = currentParent;
        var views = lContainer.views;
        for (var i = 0; i < views.length; i++) {
            addRemoveViewFromContainer(node, views[i], true, null);
        }
    }
    if (node.dynamicLContainerNode) {
        node.dynamicLContainerNode.data.renderParent = currentParent;
    }
}
exports.appendProjectedNode = appendProjectedNode;
//# sourceMappingURL=node_manipulation.js.map