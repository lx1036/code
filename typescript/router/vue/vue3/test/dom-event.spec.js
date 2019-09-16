import {Vue3} from "../src";


describe('DOM event', function () {
  let cb = jasmine.createSpy('cb');

  it('event listener', function () {
    let vm = new Vue3({
      render(h) {
        return h('button', {class: 'btn', on: {'click': cb}}, []);
      }
    }).$mount();

    document.body.appendChild(vm.$el);
    const btn = document.querySelector('.btn');
    expect(btn.tagName).toEqual('BUTTON');
    btn.click();
    expect(cb).toHaveBeenCalled();
    
    document.body.removeChild(vm.$el);
  });
});
