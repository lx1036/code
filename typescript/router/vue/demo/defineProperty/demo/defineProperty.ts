// cd /Users/liuxiang/rightcapital/rightcapital/router/vue/demo/defineProperty/demo
// ../../../../node_modules/.bin/ts-node defineProperty.ts

let obj = {name: 'lx1036'};
Object.defineProperty(obj, 'name', {
  set: v => {
    console.log(v);
  },
  get: () => {
    console.log('get');
  }
});

let name1 = obj.name;
obj.name = 'lx1037';
// get lx1037
