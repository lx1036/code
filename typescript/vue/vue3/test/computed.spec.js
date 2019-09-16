import {Vue3} from "../src";

describe('computed', () => {
  it('basic', () => {
    const vm = new Vue3({
      data() {
        return {a: 'hello', b: 'world'};
      },
      computed: {
        c() {
          return this.a + ' ' + this.b;
        }
      },
      render(createElement) {
        return createElement('div', {}, this.c);
      }
    }).$mount();
    
    expect(vm.c).toEqual('hello world');
    expect(vm.$el.textContent).toEqual('hello world');
    
    vm.a = 'bye';
    expect(vm.c).toEqual('bye world');
    expect(vm.$el.textContent).toEqual('bye world');
  });
  
  it('chain', () => {
    const vm = new Vue3({
      data() {
        return {a: 'hello', b: 'world'};
      },
      computed: {
        c() {
          return this.a + ' ' + this.b;
        },
        d() {
          return this.c + '!!!';
        },
      },
      render(createElement) {
        return createElement('div', {}, this.d);
      }
    }).$mount();
  
    expect(vm.d).toEqual('hello world!!!');
    expect(vm.$el.textContent).toEqual('hello world!!!');
  
    vm.a = 'bye';
    expect(vm.d).toEqual('bye world!!!');
    expect(vm.$el.textContent).toEqual('bye world!!!');
  });
});
