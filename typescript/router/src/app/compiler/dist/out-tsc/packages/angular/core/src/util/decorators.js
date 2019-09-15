"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.ANNOTATIONS = '__annotations__';
exports.PARAMETERS = '__parameters__';
exports.PROP_METADATA = '__prop__metadata__';
/**
 * @suppress {globalThis}
 */
function makeDecorator(name, props, parentClass, chainFn, typeFn) {
    var metaCtor = makeMetadataCtor(props);
    function DecoratorFactory() {
        var args = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            args[_i] = arguments[_i];
        }
        if (this instanceof DecoratorFactory) {
            metaCtor.call.apply(metaCtor, [this].concat(args));
            return this;
        }
        var annotationInstance = new ((_a = DecoratorFactory).bind.apply(_a, [void 0].concat(args)))();
        var TypeDecorator = function TypeDecorator(cls) {
            typeFn && typeFn.apply(void 0, [cls].concat(args));
            // Use of Object.defineProperty is important since it creates non-enumerable property which
            // prevents the property is copied during subclassing.
            var annotations = cls.hasOwnProperty(exports.ANNOTATIONS) ?
                cls[exports.ANNOTATIONS] :
                Object.defineProperty(cls, exports.ANNOTATIONS, { value: [] })[exports.ANNOTATIONS];
            annotations.push(annotationInstance);
            return cls;
        };
        if (chainFn)
            chainFn(TypeDecorator);
        return TypeDecorator;
        var _a;
    }
    if (parentClass) {
        DecoratorFactory.prototype = Object.create(parentClass.prototype);
    }
    DecoratorFactory.prototype.ngMetadataName = name;
    DecoratorFactory.annotationCls = DecoratorFactory;
    return DecoratorFactory;
}
exports.makeDecorator = makeDecorator;
function makeMetadataCtor(props) {
    return function ctor() {
        var args = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            args[_i] = arguments[_i];
        }
        if (props) {
            var values = props.apply(void 0, args);
            for (var propName in values) {
                this[propName] = values[propName];
            }
        }
    };
}
function makeParamDecorator(name, props, parentClass) {
    var metaCtor = makeMetadataCtor(props);
    function ParamDecoratorFactory() {
        var args = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            args[_i] = arguments[_i];
        }
        if (this instanceof ParamDecoratorFactory) {
            metaCtor.apply(this, args);
            return this;
        }
        var annotationInstance = new ((_a = ParamDecoratorFactory).bind.apply(_a, [void 0].concat(args)))();
        ParamDecorator.annotation = annotationInstance;
        return ParamDecorator;
        function ParamDecorator(cls, unusedKey, index) {
            // Use of Object.defineProperty is important since it creates non-enumerable property which
            // prevents the property is copied during subclassing.
            var parameters = cls.hasOwnProperty(exports.PARAMETERS) ?
                cls[exports.PARAMETERS] :
                Object.defineProperty(cls, exports.PARAMETERS, { value: [] })[exports.PARAMETERS];
            // there might be gaps if some in between parameters do not have annotations.
            // we pad with nulls.
            while (parameters.length <= index) {
                parameters.push(null);
            }
            (parameters[index] = parameters[index] || []).push(annotationInstance);
            return cls;
        }
        var _a;
    }
    if (parentClass) {
        ParamDecoratorFactory.prototype = Object.create(parentClass.prototype);
    }
    ParamDecoratorFactory.prototype.ngMetadataName = name;
    ParamDecoratorFactory.annotationCls = ParamDecoratorFactory;
    return ParamDecoratorFactory;
}
exports.makeParamDecorator = makeParamDecorator;
function makePropDecorator(name, props, parentClass) {
    var metaCtor = makeMetadataCtor(props);
    function PropDecoratorFactory() {
        var args = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            args[_i] = arguments[_i];
        }
        if (this instanceof PropDecoratorFactory) {
            metaCtor.apply(this, args);
            return this;
        }
        var decoratorInstance = new ((_a = PropDecoratorFactory).bind.apply(_a, [void 0].concat(args)))();
        return function PropDecorator(target, name) {
            var constructor = target.constructor;
            // Use of Object.defineProperty is important since it creates non-enumerable property which
            // prevents the property is copied during subclassing.
            var meta = constructor.hasOwnProperty(exports.PROP_METADATA) ?
                constructor[exports.PROP_METADATA] :
                Object.defineProperty(constructor, exports.PROP_METADATA, { value: {} })[exports.PROP_METADATA];
            meta[name] = meta.hasOwnProperty(name) && meta[name] || [];
            meta[name].unshift(decoratorInstance);
        };
        var _a;
    }
    if (parentClass) {
        PropDecoratorFactory.prototype = Object.create(parentClass.prototype);
    }
    PropDecoratorFactory.prototype.ngMetadataName = name;
    PropDecoratorFactory.annotationCls = PropDecoratorFactory;
    return PropDecoratorFactory;
}
exports.makePropDecorator = makePropDecorator;
//# sourceMappingURL=decorators.js.map