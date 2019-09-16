

console.log('a');
const _global = <any>(typeof window === 'undefined' ? global : window);

console.log(Object.keys(_global));

export function test() {
  console.log('test');
}