import {UnwrapNestedRefs} from "./ref";
import {mutableHandlers, readonlyHandlers} from "./baseHandlers";
import {mutableCollectionHandlers, readonlyCollectionHandlers} from "./collectionHandlers";
import {isObject, toTypeString} from "./shared";
import {ReactiveEffect} from "./effect";

// WeakMaps that store {raw <-> observed} pairs.
const reactiveToRaw: WeakMap<any, any> = new WeakMap()
const readonlyToRaw: WeakMap<any, any> = new WeakMap()
const rawToReadonly: WeakMap<any, any> = new WeakMap()
const rawToReactive: WeakMap<any, any> = new WeakMap()

// WeakSets for values that are marked readonly or non-reactive during
// observable creation.
const readonlyValues: WeakSet<any> = new WeakSet()
const nonReactiveValues: WeakSet<any> = new WeakSet()

export function readonly<T extends object>(target: T): Readonly<UnwrapNestedRefs<T>>
export function readonly(target: object) {
  // value is a mutable observable, retrieve its original and return
  // a readonly version.
  if (reactiveToRaw.has(target)) {
    target = reactiveToRaw.get(target)
  }
  return createReactiveObject(
    target,
    rawToReadonly,
    readonlyToRaw,
    readonlyHandlers,
    readonlyCollectionHandlers
  )
}

export function reactive<T extends object>(target: T): UnwrapNestedRefs<T>
export function reactive(target: object) {
// if trying to observe a readonly proxy, return the readonly version.
  if (readonlyToRaw.has(target)) {
    return target
  }
  // target is explicitly marked as readonly by user
  if (readonlyValues.has(target)) {
    return readonly(target)
  }
  
  return createReactiveObject(
    target,
    rawToReactive,
    reactiveToRaw,
    mutableHandlers,
    mutableCollectionHandlers
  )
}

export function isReactive(value: any): boolean {
  return reactiveToRaw.has(value) || readonlyToRaw.has(value)
}
const observableValueRE = /^\[object (?:Object|Array|Map|Set|WeakMap|WeakSet)\]$/
const canObserve = (value: any): boolean => {
  return (
    !value._isVue &&
    !value._isVNode &&
    observableValueRE.test(toTypeString(value)) &&
    !nonReactiveValues.has(value)
  )
}
// The main WeakMap that stores {target -> key -> dep} connections.
// Conceptually, it's easier to think of a dependency as a Dep class
// which maintains a Set of subscribers, but we simply store them as
// raw Sets to reduce memory overhead.
export type Dep = Set<ReactiveEffect>
const collectionTypes: Set<any> = new Set([Set, Map, WeakMap, WeakSet])
export const targetMap: WeakMap<any, KeyToDepMap> = new WeakMap()
export type KeyToDepMap = Map<string | symbol, Dep>
function createReactiveObject(
  target: any,
  toProxy: WeakMap<any, any>,
  toRaw: WeakMap<any, any>,
  baseHandlers: ProxyHandler<any>,
  collectionHandlers: ProxyHandler<any>
) {
  if (!isObject(target)) {
    // @ts-ignore
    if (__DEV__) { // close
      console.warn(`value cannot be made reactive: ${String(target)}`)
    }
    return target
  }
  
  // target already has corresponding Proxy
  let observed = toProxy.get(target)
  if (observed !== void 0) {
    return observed
  }
  // target is already a Proxy
  if (toRaw.has(target)) {
    return target
  }
  // only a whitelist of value types can be observed.
  if (!canObserve(target)) {
    return target
  }
  
  const handlers = collectionTypes.has(target.constructor)
    ? collectionHandlers
    : baseHandlers
  observed = new Proxy(target, handlers)
  toProxy.set(target, observed)
  toRaw.set(observed, target)
  if (!targetMap.has(target)) {
    targetMap.set(target, new Map())
  }
  return observed
}
