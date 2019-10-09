

export interface Ref<T> {
  _isRef: true
  value: UnwrapNestedRefs<T>
}
export type UnwrapNestedRefs<T> = T extends Ref<any> ? T : UnwrapRef<T>

export type UnwrapRef<T> = Array<T> // close

export function ref<T>(raw: T): Ref<T> {

}
