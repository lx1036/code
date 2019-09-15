import {Vue3} from "../src";

describe('Data Proxy', function () {
  it('vm._data.a=vm.a', function () {
    let vm = new Vue3({
      data() {
        return {a: 2};
      }
    });

    expect(vm.a).toEqual(2);
  });
});
