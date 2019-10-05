

export const readonlyHandlers: ProxyHandler<any> = {

}


export const mutableHandlers: ProxyHandler<any> = {
  get: createGetter(false),
  set,
  deleteProperty,
  has,
  ownKeys
}

function createGetter(isReadonly: boolean) {
  return function get(target: any, key: string | symbol, receiver: any) {
  
  }
}
function deleteProperty(target: any, key: string | symbol): boolean {
  return false
}
function has(target: any, key: string | symbol): boolean {
  return false
}
function ownKeys(target: any): (string | number | symbol)[] {
  return []
}
function set(
  target: any,
  key: string | symbol,
  value: any,
  receiver: any
): boolean {
  return false
}
