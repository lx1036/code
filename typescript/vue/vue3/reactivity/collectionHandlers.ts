import {hasOwn, isObject} from "./shared";
import {OperationTypes} from "./operations";
import {reactive, readonly} from "./reactive";

const toReactive = (value: any) => (isObject(value) ? reactive(value) : value)
const toReadonly = (value: any) => (isObject(value) ? readonly(value) : value)

const mutableInstrumentations: any = {
  get(key: any) {
    return get(this, key, toReactive)
  },
  get size() {
    return size(this)
  },
  has,
  add,
  set,
  delete: deleteEntry,
  clear,
  forEach: createForEach(false)
}

const readonlyInstrumentations: any = {
  get(key: any) {
    return get(this, key, toReadonly)
  },
  get size() {
    return size(this)
  },
  has,
  add: createReadonlyMethod(add, OperationTypes.ADD),
  set: createReadonlyMethod(set, OperationTypes.SET),
  delete: createReadonlyMethod(deleteEntry, OperationTypes.DELETE),
  clear: createReadonlyMethod(clear, OperationTypes.CLEAR),
  forEach: createForEach(true)
}

function createForEach(isReadonly: boolean) {

}
function clear(this: any) {

}
function set(this: any, key: any, value: any) {

}
function get(target: any, key: any, wrap: (t: any) => any): any {

}
function has(this: any, key: any): boolean {
  return false
}
function size(target: any) {

}
function add(this: any, value: any) {

}
function createReadonlyMethod(
  method: Function,
  type: OperationTypes
): Function {
  return function(this: any, ...args: any[]) {
  
  }
}

function deleteEntry(this: any, key: any) {

}
export const mutableCollectionHandlers: ProxyHandler<any> = {
  get: createInstrumentationGetter(mutableInstrumentations)
}

export const readonlyCollectionHandlers: ProxyHandler<any> = {
  get: createInstrumentationGetter(readonlyInstrumentations)
}

function createInstrumentationGetter(instrumentations: any) {
  return function getInstrumented(
    target: any,
    key: string | symbol,
    receiver: any
  ) {
    target =
      hasOwn(instrumentations, key) && key in target ? instrumentations : target
    return Reflect.get(target, key, receiver)
  }
}
