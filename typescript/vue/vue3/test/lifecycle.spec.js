import {Vue3} from "../src";


describe('lifecycle', function () {
  let cb = jasmine.createSpy('cb');

  it('mounted', function () {
    new Vue3({
      mounted() {
        cb();
      },
      render(h) {
        return h('div', null, 'hello');
      },
    }).$mount();

    expect(cb).toHaveBeenCalled();
  });
});
