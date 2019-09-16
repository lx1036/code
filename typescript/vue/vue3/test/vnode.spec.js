import {Vue3} from "../src";


describe('vnode render', function () {
  it('basic usage', function () {
    const vm = new Vue3({
      render(h) {
        return h('div', null, 'hello');
      }
    }).$mount();

    expect(vm.$el.tagName).toEqual('DIV');
    expect(vm.$el.textContent).toBe('hello');
  });

  it('basic usage with children', function () {
    const vm = new Vue3({
      render(h) {
        return h('div', {id: 'parent'}, [
          h('p', {class: 'hello', id: 'hello'}, 'hello'),
          h('span', {class: 'world', id: 'world'}, 'world'),
        ]);
      }
    }).$mount();

    expect(vm.$el.tagName).toEqual('DIV');
    expect(vm.$el.id).toEqual('parent');
    expect(vm.$el.children.length).toEqual(2);
    expect(vm.$el.firstChild.tagName).toBe('P');
    expect(vm.$el.firstChild.id).toBe('hello');
    expect(vm.$el.firstChild.classList).toContain('hello');
    expect(vm.$el.lastChild.tagName).toBe('SPAN');
    expect(vm.$el.lastChild.id).toBe('world');
    expect(vm.$el.lastChild.classList).toContain('world');
  });
});
