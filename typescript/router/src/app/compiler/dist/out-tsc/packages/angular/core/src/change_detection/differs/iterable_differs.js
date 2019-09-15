"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var defs_1 = require("../../di/defs");
var metadata_1 = require("../../di/metadata");
var default_iterable_differ_1 = require("../differs/default_iterable_differ");
/**
 * A repository of different iterable diffing strategies used by NgFor, NgClass, and others.
 *
 */
var IterableDiffers = /** @class */ (function () {
    function IterableDiffers(factories) {
        this.factories = factories;
    }
    IterableDiffers.create = function (factories, parent) {
        if (parent != null) {
            var copied = parent.factories.slice();
            factories = factories.concat(copied);
        }
        return new IterableDiffers(factories);
    };
    /**
     * Takes an array of {@link IterableDifferFactory} and returns a provider used to extend the
     * inherited {@link IterableDiffers} instance with the provided factories and return a new
     * {@link IterableDiffers} instance.
     *
     * The following example shows how to extend an existing list of factories,
     * which will only be applied to the injector for this component and its children.
     * This step is all that's required to make a new {@link IterableDiffer} available.
     *
     * ### Example
     *
     * ```
     * @Component({
     *   viewProviders: [
     *     IterableDiffers.extend([new ImmutableListDiffer()])
     *   ]
     * })
     * ```
     */
    /**
       * Takes an array of {@link IterableDifferFactory} and returns a provider used to extend the
       * inherited {@link IterableDiffers} instance with the provided factories and return a new
       * {@link IterableDiffers} instance.
       *
       * The following example shows how to extend an existing list of factories,
       * which will only be applied to the injector for this component and its children.
       * This step is all that's required to make a new {@link IterableDiffer} available.
       *
       * ### Example
       *
       * ```
       * @Component({
       *   viewProviders: [
       *     IterableDiffers.extend([new ImmutableListDiffer()])
       *   ]
       * })
       * ```
       */
    IterableDiffers.extend = /**
       * Takes an array of {@link IterableDifferFactory} and returns a provider used to extend the
       * inherited {@link IterableDiffers} instance with the provided factories and return a new
       * {@link IterableDiffers} instance.
       *
       * The following example shows how to extend an existing list of factories,
       * which will only be applied to the injector for this component and its children.
       * This step is all that's required to make a new {@link IterableDiffer} available.
       *
       * ### Example
       *
       * ```
       * @Component({
       *   viewProviders: [
       *     IterableDiffers.extend([new ImmutableListDiffer()])
       *   ]
       * })
       * ```
       */
    function (factories) {
        return {
            provide: IterableDiffers,
            useFactory: function (parent) {
                if (!parent) {
                    // Typically would occur when calling IterableDiffers.extend inside of dependencies passed
                    // to
                    // bootstrap(), which would override default pipes instead of extending them.
                    throw new Error('Cannot extend IterableDiffers without a parent injector');
                }
                return IterableDiffers.create(factories, parent);
            },
            // Dependency technically isn't optional, but we can provide a better error message this way.
            deps: [[IterableDiffers, new metadata_1.SkipSelf(), new metadata_1.Optional()]]
        };
    };
    IterableDiffers.prototype.find = function (iterable) {
        var factory = this.factories.find(function (f) { return f.supports(iterable); });
        if (factory != null) {
            return factory;
        }
        else {
            throw new Error("Cannot find a differ supporting object '" + iterable + "' of type '" + getTypeNameForDebugging(iterable) + "'");
        }
    };
    IterableDiffers.ngInjectableDef = defs_1.defineInjectable({
        providedIn: 'root',
        factory: function () { return new IterableDiffers([new default_iterable_differ_1.DefaultIterableDifferFactory()]); }
    });
    return IterableDiffers;
}());
exports.IterableDiffers = IterableDiffers;
function getTypeNameForDebugging(type) {
    return type['name'] || typeof type;
}
exports.getTypeNameForDebugging = getTypeNameForDebugging;
//# sourceMappingURL=iterable_differs.js.map