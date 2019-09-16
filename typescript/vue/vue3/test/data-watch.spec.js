import {Vue3} from "../src";


describe('Watch data change', function () {
  it('cb is called', function () {
    let vm = new Vue3({
      data() {
        return {a: 2}
      }
    });

    vm.$watch('a', (prev, val) => {
      expect(prev).toEqual(2);
      expect(val).toEqual(3);
    });

    vm.a = 3;
  });
});
