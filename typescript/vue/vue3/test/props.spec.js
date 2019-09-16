import {Vue3} from "../src";

describe('props', () => {
  it('basic', () => {
    const vm = new Vue3({
      props: ['hello'],
      propsData: {
        hello: 'world',
      },
      render(h) {
        return h('div', null, this.hello);
      }
    }).$mount();
    
    expect(vm.hello).toEqual('world');
    expect(vm.$el.textContent).toEqual('world');
    
    vm.hello = 'world2';
    expect(vm.hello).toEqual('world2');
    expect(vm.$el.textContent).toEqual('world2');
  });
});
