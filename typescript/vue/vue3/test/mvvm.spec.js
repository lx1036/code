import {Vue3} from "../src";

describe('mvvm', function () {
  it('basic usage', function () {
    const vm = new Vue3({
      data() {
        return {a: 0};
      },
      render(h) {
        return h('div', null, this.a);
      }
    }).$mount();

    vm.a++;
    expect(vm.$el.textContent).toEqual('1');
    vm.a = 999;
    expect(vm.$el.textContent).toEqual('999');
  });
  
  it('deep object', () => {
    const vm = new Vue3({
      data() {
        return {a: {b: 0}};
      },
      render(h) {
        return h('div', null, this.a.b);
      }
    }).$mount();
    
    expect(vm.a.b).toEqual(0);
    vm.a.b++;
    expect(vm.a.b).toEqual(1);
    expect(vm.$el.textContent).toEqual('1');
    vm.a.b = 999;
    expect(vm.$el.textContent).toEqual('999');
  });
  
  it('add or delete object property', () => {
    const vm = new Vue3({
      data() {
        return {a: {}};
      },
      render(h) {
        return h('div', null, this.a.b);
      }
    }).$mount();
    
    expect(vm.$el.textContent).toEqual('undefined');
    
    vm.a.b = 0;
    expect(vm.a.b).toEqual(0);
    expect(vm.$el.textContent).toEqual('0');
  
    delete vm.a.b;
    expect(vm.a.b).toEqual(undefined);
    expect(vm.$el.textContent).toEqual('undefined');
  });
  
  it('array getter/setter', () => {
    const vm = new Vue3({
      data() {
        return {a: ['hello']};
      },
      render(h) {
        return h('div', null, this.a[0]);
      }
    }).$mount();
    
    expect(vm.a[0]).toEqual('hello');
    expect(vm.$el.textContent).toEqual('hello');
    vm.a[0] = 'world';
    expect(vm.a[0]).toEqual('world');
    expect(vm.$el.textContent).toEqual('world');
  });
  
  it('array push/pop/shift/unshift/splice/sort/reverse', () => {
    const vm = new Vue3({
      data() {
        return {a: ['hello']};
      },
      render(h) {
        return h('div', null, this.a[this.a.length - 1]);
      }
    }).$mount();
  
    expect(vm.a[0]).toEqual('hello');
    expect(vm.$el.textContent).toEqual('hello');
    vm.a.push('world');
    expect(vm.a[1]).toEqual('world');
    expect(vm.$el.textContent).toEqual('hello');
    vm.a.pop();
    expect(vm.a[0]).toEqual('hello');
    expect(vm.$el.textContent).toEqual('hello');
    vm.a.shift();
    expect(vm.a.length).toEqual(0);
    expect(vm.$el.textContent).toEqual('undefined');
    vm.a.unshift('hello');
    expect(vm.a[0]).toEqual('hello');
    expect(vm.$el.textContent).toEqual('hello');
    vm.a.splice(0, 1, 'world', 'hello');
    expect(vm.a[0]).toEqual('world');
    expect(vm.a[1]).toEqual('hello');
    expect(vm.$el.textContent).toEqual('hello');
    vm.a.sort();
    expect(vm.a[0]).toEqual('hello');
    expect(vm.a[1]).toEqual('world');
    expect(vm.$el.textContent).toEqual('world');
    vm.a.reverse();
    expect(vm.a[0]).toEqual('world');
    expect(vm.a[1]).toEqual('hello');
    expect(vm.$el.textContent).toEqual('hello');
  });
  
  it('setter/getter property of object in array', () => {
    const vm = new Vue3({
      data() {
        return {a: [{msg: 'hello'}]};
      },
      render(h) {
        return h('div', null, this.a[0].msg);
      }
    }).$mount();
    
    expect(vm.a[0].msg).toEqual('hello');
    expect(vm.$el.textContent).toEqual('hello');
    
    vm.a[0].msg = 'world';
    expect(vm.a[0].msg).toEqual('world');
    expect(vm.$el.textContent).toEqual('world');
  });
});
