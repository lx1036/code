import {mockWarn} from "./jestUtils";
import {reactive, isReactive} from './reactive';

/**
 * @see [Vue3响应式系统源码解析(上)](https://zhuanlan.zhihu.com/p/85678790)
 */
describe('reactive', () => {
  mockWarn()
  
  test('Object', () => {
    const original = { foo: 1 }
    const observed = reactive(original)
    expect(observed).not.toBe(original)
    expect(isReactive(observed)).toBe(true)
    expect(isReactive(original)).toBe(false)
    // get
    console.log(observed.foo)
    expect(observed.foo).toBe(1)
    // has
    expect('foo' in observed).toBe(true)
    // ownKeys
    expect(Object.keys(observed)).toEqual(['foo'])
  });
  
  test('Array', () => {
    const original: any[] = [{ foo: 1 }]
    const observed = reactive(original)
    expect(observed).not.toBe(original)
    expect(isReactive(observed)).toBe(true)
    expect(isReactive(original)).toBe(false)
    expect(isReactive(observed[0])).toBe(true)
    // get
    expect(observed[0].foo).toBe(1)
    // has
    expect(0 in observed).toBe(true)
    // ownKeys
    expect(Object.keys(observed)).toEqual(['0'])
  })
  
  test('Nested array', () => {
    const original: any[] = [{ foo: 1, a: {b: {c: 1}}, arr: [{d: {}}]}]
    const observed = reactive(original)
    expect(observed).not.toBe(original)
    expect(isReactive(observed)).toBe(true)
    expect(isReactive(original)).toBe(false)
    expect(isReactive(observed[0])).toBe(true)
    // observed.a.b 是reactive
    expect(isReactive(observed[0].a.b)).toBe(true)
    // observed[0].arr[0].d 是reactive
    expect(isReactive(observed[0].arr[0].d)).toBe(true)
    // get
    expect(observed[0].foo).toBe(1)
    // has
    expect(0 in observed).toBe(true)
    // ownKeys
    expect(Object.keys(observed)).toEqual(['0'])
  });
  
  test('cloned reactive Array should point to observed values', () => {
    const original = [{ foo: 1 }]
    const observed = reactive(original)
    const clone = observed.slice()
    expect(isReactive(clone[0])).toBe(true)
    expect(clone[0]).not.toBe(original[0])
    expect(clone[0]).toBe(observed[0])
  })
});
