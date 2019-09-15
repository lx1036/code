"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var injector_1 = require("../di/injector");
var assert_1 = require("./assert");
var instructions_1 = require("./instructions");
var node_assert_1 = require("./node_assert");
var node_manipulation_1 = require("./node_manipulation");
var util_1 = require("./util");
var view_ref_1 = require("./view_ref");
/**
 * If a directive is diPublic, bloomAdd sets a property on the instance with this constant as
 * the key and the directive's unique ID as the value. This allows us to map directives to their
 * bloom filter bit for DI.
 */
var NG_ELEMENT_ID = '__NG_ELEMENT_ID__';
/**
 * The number of slots in each bloom filter (used by DI). The larger this number, the fewer
 * directives that will share slots, and thus, the fewer false positives when checking for
 * the existence of a directive.
 */
var BLOOM_SIZE = 256;
/** Counter used to generate unique IDs for directives. */
var nextNgElementId = 0;
/**
 * Registers this directive as present in its node's injector by flipping the directive's
 * corresponding bit in the injector's bloom filter.
 *
 * @param injector The node injector in which the directive should be registered
 * @param type The directive to register
 */
function bloomAdd(injector, type) {
    var id = type[NG_ELEMENT_ID];
    // Set a unique ID on the directive type, so if something tries to inject the directive,
    // we can easily retrieve the ID and hash it into the bloom bit that should be checked.
    if (id == null) {
        id = type[NG_ELEMENT_ID] = nextNgElementId++;
    }
    // We only have BLOOM_SIZE (256) slots in our bloom filter (8 buckets * 32 bits each),
    // so all unique IDs must be modulo-ed into a number from 0 - 255 to fit into the filter.
    // This means that after 255, some directives will share slots, leading to some false positives
    // when checking for a directive's presence.
    var bloomBit = id % BLOOM_SIZE;
    // Create a mask that targets the specific bit associated with the directive.
    // JS bit operations are 32 bits, so this will be a number between 2^0 and 2^31, corresponding
    // to bit positions 0 - 31 in a 32 bit integer.
    var mask = 1 << bloomBit;
    // Use the raw bloomBit number to determine which bloom filter bucket we should check
    // e.g: bf0 = [0 - 31], bf1 = [32 - 63], bf2 = [64 - 95], bf3 = [96 - 127], etc
    if (bloomBit < 128) {
        // Then use the mask to flip on the bit (0-31) associated with the directive in that bucket
        bloomBit < 64 ? (bloomBit < 32 ? (injector.bf0 |= mask) : (injector.bf1 |= mask)) :
            (bloomBit < 96 ? (injector.bf2 |= mask) : (injector.bf3 |= mask));
    }
    else {
        bloomBit < 192 ? (bloomBit < 160 ? (injector.bf4 |= mask) : (injector.bf5 |= mask)) :
            (bloomBit < 224 ? (injector.bf6 |= mask) : (injector.bf7 |= mask));
    }
}
exports.bloomAdd = bloomAdd;
function getOrCreateNodeInjector() {
    ngDevMode && instructions_1.assertPreviousIsParent();
    return getOrCreateNodeInjectorForNode(instructions_1.getPreviousOrParentNode());
}
exports.getOrCreateNodeInjector = getOrCreateNodeInjector;
/**
 * Creates (or gets an existing) injector for a given element or container.
 *
 * @param node for which an injector should be retrieved / created.
 * @returns Node injector
 */
function getOrCreateNodeInjectorForNode(node) {
    var nodeInjector = node.nodeInjector;
    var parentInjector = node.parent && node.parent.nodeInjector;
    if (nodeInjector != parentInjector) {
        return nodeInjector;
    }
    return node.nodeInjector = {
        parent: parentInjector,
        node: node,
        bf0: 0,
        bf1: 0,
        bf2: 0,
        bf3: 0,
        bf4: 0,
        bf5: 0,
        bf6: 0,
        bf7: 0,
        cbf0: parentInjector == null ? 0 : parentInjector.cbf0 | parentInjector.bf0,
        cbf1: parentInjector == null ? 0 : parentInjector.cbf1 | parentInjector.bf1,
        cbf2: parentInjector == null ? 0 : parentInjector.cbf2 | parentInjector.bf2,
        cbf3: parentInjector == null ? 0 : parentInjector.cbf3 | parentInjector.bf3,
        cbf4: parentInjector == null ? 0 : parentInjector.cbf4 | parentInjector.bf4,
        cbf5: parentInjector == null ? 0 : parentInjector.cbf5 | parentInjector.bf5,
        cbf6: parentInjector == null ? 0 : parentInjector.cbf6 | parentInjector.bf6,
        cbf7: parentInjector == null ? 0 : parentInjector.cbf7 | parentInjector.bf7,
        templateRef: null,
        viewContainerRef: null,
        elementRef: null,
        changeDetectorRef: null
    };
}
exports.getOrCreateNodeInjectorForNode = getOrCreateNodeInjectorForNode;
/**
 * Makes a directive public to the DI system by adding it to an injector's bloom filter.
 *
 * @param di The node injector in which a directive will be added
 * @param def The definition of the directive to be made public
 */
function diPublicInInjector(di, def) {
    bloomAdd(di, def.type);
}
exports.diPublicInInjector = diPublicInInjector;
/**
 * Makes a directive public to the DI system by adding it to an injector's bloom filter.
 *
 * @param def The definition of the directive to be made public
 */
function diPublic(def) {
    diPublicInInjector(getOrCreateNodeInjector(), def);
}
exports.diPublic = diPublic;
function directiveInject(token, flags) {
    if (flags === void 0) { flags = 0 /* Default */; }
    return getOrCreateInjectable(getOrCreateNodeInjector(), token, flags);
}
exports.directiveInject = directiveInject;
/**
 * Creates an ElementRef and stores it on the injector.
 * Or, if the ElementRef already exists, retrieves the existing ElementRef.
 *
 * @returns The ElementRef instance to use
 */
function injectElementRef() {
    return getOrCreateElementRef(getOrCreateNodeInjector());
}
exports.injectElementRef = injectElementRef;
/**
 * Creates a TemplateRef and stores it on the injector. Or, if the TemplateRef already
 * exists, retrieves the existing TemplateRef.
 *
 * @returns The TemplateRef instance to use
 */
function injectTemplateRef() {
    return getOrCreateTemplateRef(getOrCreateNodeInjector());
}
exports.injectTemplateRef = injectTemplateRef;
/**
 * Creates a ViewContainerRef and stores it on the injector. Or, if the ViewContainerRef
 * already exists, retrieves the existing ViewContainerRef.
 *
 * @returns The ViewContainerRef instance to use
 */
function injectViewContainerRef() {
    return getOrCreateContainerRef(getOrCreateNodeInjector());
}
exports.injectViewContainerRef = injectViewContainerRef;
/** Returns a ChangeDetectorRef (a.k.a. a ViewRef) */
function injectChangeDetectorRef() {
    return getOrCreateChangeDetectorRef(getOrCreateNodeInjector(), null);
}
exports.injectChangeDetectorRef = injectChangeDetectorRef;
/**
 * Inject static attribute value into directive constructor.
 *
 * This method is used with `factory` functions which are generated as part of
 * `defineDirective` or `defineComponent`. The method retrieves the static value
 * of an attribute. (Dynamic attributes are not supported since they are not resolved
 *  at the time of injection and can change over time.)
 *
 * # Example
 * Given:
 * ```
 * @Component(...)
 * class MyComponent {
 *   constructor(@Attribute('title') title: string) { ... }
 * }
 * ```
 * When instantiated with
 * ```
 * <my-component title="Hello"></my-component>
 * ```
 *
 * Then factory method generated is:
 * ```
 * MyComponent.ngComponentDef = defineComponent({
 *   factory: () => new MyComponent(injectAttribute('title'))
 *   ...
 * })
 * ```
 *
 * @experimental
 */
function injectAttribute(attrName) {
    ngDevMode && instructions_1.assertPreviousIsParent();
    var lElement = instructions_1.getPreviousOrParentNode();
    ngDevMode && node_assert_1.assertNodeType(lElement, 3 /* Element */);
    var tElement = (lElement.tNode);
    ngDevMode && assert_1.assertNotNull(tElement, 'expecting tNode');
    var attrs = tElement.attrs;
    if (attrs) {
        for (var i = 0; i < attrs.length; i = i + 2) {
            if (attrs[i] == attrName) {
                return attrs[i + 1];
            }
        }
    }
    return undefined;
}
exports.injectAttribute = injectAttribute;
/**
 * Creates a ViewRef and stores it on the injector as ChangeDetectorRef (public alias).
 * Or, if it already exists, retrieves the existing instance.
 *
 * @returns The ChangeDetectorRef to use
 */
function getOrCreateChangeDetectorRef(di, context) {
    if (di.changeDetectorRef)
        return di.changeDetectorRef;
    var currentNode = di.node;
    if (instructions_1.isComponent((currentNode.tNode))) {
        return di.changeDetectorRef = view_ref_1.createViewRef(currentNode.data, context);
    }
    else if (currentNode.type === 3 /* Element */) {
        return di.changeDetectorRef = getOrCreateHostChangeDetector(currentNode.view.node);
    }
    return null;
}
exports.getOrCreateChangeDetectorRef = getOrCreateChangeDetectorRef;
/** Gets or creates ChangeDetectorRef for the closest host component */
function getOrCreateHostChangeDetector(currentNode) {
    var hostNode = getClosestComponentAncestor(currentNode);
    var hostInjector = hostNode.nodeInjector;
    var existingRef = hostInjector && hostInjector.changeDetectorRef;
    return existingRef ?
        existingRef :
        view_ref_1.createViewRef(hostNode.data, hostNode.view
            .directives[hostNode.tNode.flags >> 13 /* DirectiveStartingIndexShift */]);
}
/**
 * If the node is an embedded view, traverses up the view tree to return the closest
 * ancestor view that is attached to a component. If it's already a component node,
 * returns itself.
 */
function getClosestComponentAncestor(node) {
    while (node.type === 2 /* View */) {
        node = node.view.node;
    }
    return node;
}
/**
 * Searches for an instance of the given directive type up the injector tree and returns
 * that instance if found.
 *
 * Specifically, it gets the bloom filter bit associated with the directive (see bloomHashBit),
 * checks that bit against the bloom filter structure to identify an injector that might have
 * the directive (see bloomFindPossibleInjector), then searches the directives on that injector
 * for a match.
 *
 * If not found, it will propagate up to the next parent injector until the token
 * is found or the top is reached.
 *
 * @param di Node injector where the search should start
 * @param token The directive type to search for
 * @param flags Injection flags (e.g. CheckParent)
 * @returns The instance found
 */
function getOrCreateInjectable(di, token, flags) {
    var bloomHash = bloomHashBit(token);
    // If the token has a bloom hash, then it is a directive that is public to the injection system
    // (diPublic). If there is no hash, fall back to the module injector.
    if (bloomHash === null) {
        var moduleInjector = instructions_1.getPreviousOrParentNode().view.injector;
        var formerInjector = injector_1.setCurrentInjector(moduleInjector);
        try {
            return injector_1.inject(token, flags);
        }
        finally {
            injector_1.setCurrentInjector(formerInjector);
        }
    }
    else {
        var injector = di;
        while (injector) {
            // Get the closest potential matching injector (upwards in the injector tree) that
            // *potentially* has the token.
            injector = bloomFindPossibleInjector(injector, bloomHash);
            // If no injector is found, we *know* that there is no ancestor injector that contains the
            // token, so we abort.
            if (!injector) {
                break;
            }
            // At this point, we have an injector which *may* contain the token, so we step through the
            // directives associated with the injector's corresponding node to get the directive instance.
            var node = injector.node;
            var flags_1 = node.tNode.flags;
            var count = flags_1 & 4095 /* DirectiveCountMask */;
            if (count !== 0) {
                var start = flags_1 >> 13 /* DirectiveStartingIndexShift */;
                var end = start + count;
                var defs = (node.view.tView.directives);
                for (var i = start; i < end; i++) {
                    // Get the definition for the directive at this index and, if it is injectable (diPublic),
                    // and matches the given token, return the directive instance.
                    var directiveDef = defs[i];
                    if (directiveDef.type === token && directiveDef.diPublic) {
                        return instructions_1.getDirectiveInstance(node.view.directives[i]);
                    }
                }
            }
            // If we *didn't* find the directive for the token and we are searching the current node's
            // injector, it's possible the directive is on this node and hasn't been created yet.
            var instance = void 0;
            if (injector === di && (instance = searchMatchesQueuedForCreation(node, token))) {
                return instance;
            }
            // The def wasn't found anywhere on this node, so it might be a false positive.
            // Traverse up the tree and continue searching.
            injector = injector.parent;
        }
    }
    // No directive was found for the given token.
    // TODO: implement optional, check-self, and check-parent.
    throw new Error('Implement');
}
exports.getOrCreateInjectable = getOrCreateInjectable;
function searchMatchesQueuedForCreation(node, token) {
    var matches = node.view.tView.currentMatches;
    if (matches) {
        for (var i = 0; i < matches.length; i += 2) {
            var def = matches[i];
            if (def.type === token) {
                return instructions_1.resolveDirective(def, i + 1, matches, node.view.tView);
            }
        }
    }
    return null;
}
/**
 * Given a directive type, this function returns the bit in an injector's bloom filter
 * that should be used to determine whether or not the directive is present.
 *
 * When the directive was added to the bloom filter, it was given a unique ID that can be
 * retrieved on the class. Since there are only BLOOM_SIZE slots per bloom filter, the directive's
 * ID must be modulo-ed by BLOOM_SIZE to get the correct bloom bit (directives share slots after
 * BLOOM_SIZE is reached).
 *
 * @param type The directive type
 * @returns The bloom bit to check for the directive
 */
function bloomHashBit(type) {
    var id = type[NG_ELEMENT_ID];
    return typeof id === 'number' ? id % BLOOM_SIZE : null;
}
/**
 * Finds the closest injector that might have a certain directive.
 *
 * Each directive corresponds to a bit in an injector's bloom filter. Given the bloom bit to
 * check and a starting injector, this function traverses up injectors until it finds an
 * injector that contains a 1 for that bit in its bloom filter. A 1 indicates that the
 * injector may have that directive. It only *may* have the directive because directives begin
 * to share bloom filter bits after the BLOOM_SIZE is reached, and it could correspond to a
 * different directive sharing the bit.
 *
 * Note: We can skip checking further injectors up the tree if an injector's cbf structure
 * has a 0 for that bloom bit. Since cbf contains the merged value of all the parent
 * injectors, a 0 in the bloom bit indicates that the parents definitely do not contain
 * the directive and do not need to be checked.
 *
 * @param injector The starting node injector to check
 * @param  bloomBit The bit to check in each injector's bloom filter
 * @returns An injector that might have the directive
 */
function bloomFindPossibleInjector(startInjector, bloomBit) {
    // Create a mask that targets the specific bit associated with the directive we're looking for.
    // JS bit operations are 32 bits, so this will be a number between 2^0 and 2^31, corresponding
    // to bit positions 0 - 31 in a 32 bit integer.
    var mask = 1 << bloomBit;
    // Traverse up the injector tree until we find a potential match or until we know there *isn't* a
    // match.
    var injector = startInjector;
    while (injector) {
        // Our bloom filter size is 256 bits, which is eight 32-bit bloom filter buckets:
        // bf0 = [0 - 31], bf1 = [32 - 63], bf2 = [64 - 95], bf3 = [96 - 127], etc.
        // Get the bloom filter value from the appropriate bucket based on the directive's bloomBit.
        var value = void 0;
        if (bloomBit < 128) {
            value = bloomBit < 64 ? (bloomBit < 32 ? injector.bf0 : injector.bf1) :
                (bloomBit < 96 ? injector.bf2 : injector.bf3);
        }
        else {
            value = bloomBit < 192 ? (bloomBit < 160 ? injector.bf4 : injector.bf5) :
                (bloomBit < 224 ? injector.bf6 : injector.bf7);
        }
        // If the bloom filter value has the bit corresponding to the directive's bloomBit flipped on,
        // this injector is a potential match.
        if ((value & mask) === mask) {
            return injector;
        }
        // If the current injector does not have the directive, check the bloom filters for the ancestor
        // injectors (cbf0 - cbf7). These filters capture *all* ancestor injectors.
        if (bloomBit < 128) {
            value = bloomBit < 64 ? (bloomBit < 32 ? injector.cbf0 : injector.cbf1) :
                (bloomBit < 96 ? injector.cbf2 : injector.cbf3);
        }
        else {
            value = bloomBit < 192 ? (bloomBit < 160 ? injector.cbf4 : injector.cbf5) :
                (bloomBit < 224 ? injector.cbf6 : injector.cbf7);
        }
        // If the ancestor bloom filter value has the bit corresponding to the directive, traverse up to
        // find the specific injector. If the ancestor bloom filter does not have the bit, we can abort.
        injector = (value & mask) ? injector.parent : null;
    }
    return null;
}
exports.bloomFindPossibleInjector = bloomFindPossibleInjector;
var ReadFromInjectorFn = /** @class */ (function () {
    function ReadFromInjectorFn(read) {
        this.read = read;
    }
    return ReadFromInjectorFn;
}());
exports.ReadFromInjectorFn = ReadFromInjectorFn;
/**
 * Creates an ElementRef for a given node injector and stores it on the injector.
 * Or, if the ElementRef already exists, retrieves the existing ElementRef.
 *
 * @param di The node injector where we should store a created ElementRef
 * @returns The ElementRef instance to use
 */
function getOrCreateElementRef(di) {
    return di.elementRef || (di.elementRef = new ElementRef(di.node.type === 0 /* Container */ ? null : di.node.native));
}
exports.getOrCreateElementRef = getOrCreateElementRef;
exports.QUERY_READ_TEMPLATE_REF = new ReadFromInjectorFn(function (injector) { return getOrCreateTemplateRef(injector); });
exports.QUERY_READ_CONTAINER_REF = new ReadFromInjectorFn(function (injector) { return getOrCreateContainerRef(injector); });
exports.QUERY_READ_ELEMENT_REF = new ReadFromInjectorFn(function (injector) { return getOrCreateElementRef(injector); });
exports.QUERY_READ_FROM_NODE = new ReadFromInjectorFn(function (injector, node, directiveIdx) {
    ngDevMode && node_assert_1.assertNodeOfPossibleTypes(node, 0 /* Container */, 3 /* Element */);
    if (directiveIdx > -1) {
        return node.view.directives[directiveIdx];
    }
    else if (node.type === 3 /* Element */) {
        return getOrCreateElementRef(injector);
    }
    else if (node.type === 0 /* Container */) {
        return getOrCreateTemplateRef(injector);
    }
    throw new Error('fail');
});
/** A ref to a node's native element. */
var /** A ref to a node's native element. */
ElementRef = /** @class */ (function () {
    function ElementRef(nativeElement) {
        this.nativeElement = nativeElement;
    }
    return ElementRef;
}());
/**
 * Creates a ViewContainerRef and stores it on the injector. Or, if the ViewContainerRef
 * already exists, retrieves the existing ViewContainerRef.
 *
 * @returns The ViewContainerRef instance to use
 */
function getOrCreateContainerRef(di) {
    if (!di.viewContainerRef) {
        var vcRefHost = di.node;
        ngDevMode && node_assert_1.assertNodeOfPossibleTypes(vcRefHost, 0 /* Container */, 3 /* Element */);
        var lContainer = instructions_1.createLContainer((vcRefHost.parent), vcRefHost.view);
        var lContainerNode = instructions_1.createLNodeObject(0 /* Container */, vcRefHost.view, (vcRefHost.parent), undefined, lContainer, null);
        vcRefHost.dynamicLContainerNode = lContainerNode;
        instructions_1.addToViewTree(vcRefHost.view, lContainer);
        di.viewContainerRef = new ViewContainerRef(lContainerNode);
    }
    return di.viewContainerRef;
}
exports.getOrCreateContainerRef = getOrCreateContainerRef;
/**
 * A ref to a container that enables adding and removing views from that container
 * imperatively.
 */
var /**
 * A ref to a container that enables adding and removing views from that container
 * imperatively.
 */
ViewContainerRef = /** @class */ (function () {
    function ViewContainerRef(_lContainerNode) {
        this._lContainerNode = _lContainerNode;
        this._viewRefs = [];
    }
    ViewContainerRef.prototype.clear = function () {
        var lContainer = this._lContainerNode.data;
        while (lContainer.views.length) {
            this.remove(0);
        }
    };
    ViewContainerRef.prototype.get = function (index) { return this._viewRefs[index] || null; };
    Object.defineProperty(ViewContainerRef.prototype, "length", {
        get: function () {
            var lContainer = this._lContainerNode.data;
            return lContainer.views.length;
        },
        enumerable: true,
        configurable: true
    });
    ViewContainerRef.prototype.createEmbeddedView = function (templateRef, context, index) {
        var viewRef = templateRef.createEmbeddedView(context || {});
        this.insert(viewRef, index);
        return viewRef;
    };
    ViewContainerRef.prototype.createComponent = function (componentFactory, index, injector, projectableNodes, ngModule) {
        throw util_1.notImplemented();
    };
    ViewContainerRef.prototype.insert = function (viewRef, index) {
        var lViewNode = viewRef._lViewNode;
        var adjustedIdx = this._adjustIndex(index);
        node_manipulation_1.insertView(this._lContainerNode, lViewNode, adjustedIdx);
        // invalidate cache of next sibling RNode (we do similar operation in the containerRefreshEnd
        // instruction)
        this._lContainerNode.native = undefined;
        this._viewRefs.splice(adjustedIdx, 0, viewRef);
        lViewNode.parent = this._lContainerNode;
        // If the view is dynamic (has a template), it needs to be counted both at the container
        // level and at the node above the container.
        if (lViewNode.data.template !== null) {
            // Increment the container view count.
            this._lContainerNode.data.dynamicViewCount++;
            // Look for the parent node and increment its dynamic view count.
            if (this._lContainerNode.parent !== null && this._lContainerNode.parent.data !== null) {
                ngDevMode && node_assert_1.assertNodeOfPossibleTypes(this._lContainerNode.parent, 2 /* View */, 3 /* Element */);
                this._lContainerNode.parent.data.dynamicViewCount++;
            }
        }
        return viewRef;
    };
    ViewContainerRef.prototype.move = function (viewRef, newIndex) {
        var index = this.indexOf(viewRef);
        this.detach(index);
        this.insert(viewRef, this._adjustIndex(newIndex));
        return viewRef;
    };
    ViewContainerRef.prototype.indexOf = function (viewRef) { return this._viewRefs.indexOf(viewRef); };
    ViewContainerRef.prototype.remove = function (index) {
        this.detach(index);
        // TODO(ml): proper destroy of the ViewRef, i.e. recursively destroy the LviewNode and its
        // children, delete DOM nodes and QueryList, trigger hooks (onDestroy), destroy the renderer,
        // detach projected nodes
    };
    ViewContainerRef.prototype.detach = function (index) {
        var adjustedIdx = this._adjustIndex(index, -1);
        node_manipulation_1.removeView(this._lContainerNode, adjustedIdx);
        return this._viewRefs.splice(adjustedIdx, 1)[0] || null;
    };
    ViewContainerRef.prototype._adjustIndex = function (index, shift) {
        if (shift === void 0) { shift = 0; }
        if (index == null) {
            return this._lContainerNode.data.views.length + shift;
        }
        if (ngDevMode) {
            assert_1.assertGreaterThan(index, -1, 'index must be positive');
            // +1 because it's legal to insert at the end.
            // +1 because it's legal to insert at the end.
            assert_1.assertLessThan(index, this._lContainerNode.data.views.length + 1 + shift, 'index');
        }
        return index;
    };
    return ViewContainerRef;
}());
/**
 * Creates a TemplateRef and stores it on the injector. Or, if the TemplateRef already
 * exists, retrieves the existing TemplateRef.
 *
 * @param di The node injector where we should store a created TemplateRef
 * @returns The TemplateRef instance to use
 */
function getOrCreateTemplateRef(di) {
    ngDevMode && node_assert_1.assertNodeType(di.node, 0 /* Container */);
    var data = di.node.data;
    var tView = di.node.view.tView;
    return di.templateRef || (di.templateRef = new TemplateRef(getOrCreateElementRef(di), (data.template), instructions_1.getRenderer(), tView.directiveRegistry, tView.pipeRegistry));
}
exports.getOrCreateTemplateRef = getOrCreateTemplateRef;
var TemplateRef = /** @class */ (function () {
    function TemplateRef(elementRef, template, _renderer, _directives, _pipes) {
        this._renderer = _renderer;
        this._directives = _directives;
        this._pipes = _pipes;
        this.elementRef = elementRef;
        this._template = template;
    }
    TemplateRef.prototype.createEmbeddedView = function (context) {
        var viewNode = instructions_1.renderEmbeddedTemplate(null, this._template, context, this._renderer, this._directives, this._pipes);
        return view_ref_1.addDestroyable(new view_ref_1.EmbeddedViewRef(viewNode, this._template, context));
    };
    return TemplateRef;
}());
//# sourceMappingURL=di.js.map